// Package formation handle the defenitions of the serverless environment.
package formation

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/awslabs/goformation"
	"github.com/awslabs/goformation/cloudformation"
	"github.com/inconshreveable/log15"
	"github.com/mitchellh/mapstructure"

	"github.com/minutelab/mless/util/ldebug"
)

var confProcessor func(fn *Function) error

// SetConfProcessor set soemthing that can farther process function configuration
func SetConfProcessor(f func(fn *Function) error) {
	confProcessor = f
}

// Functions represent a SAM template file
// It contain the functions and has the ability to reload itself when needed.
type Functions struct {
	fname         string // template file name
	baseDir       string // base directory of the template
	desktopDir    string // directory on desktop that correspond to baseDir
	lock          sync.Mutex
	refreshTicker *time.Ticker
	modTime       time.Time
	functions     map[string]Function
}

// Function represent a lambda function defenition.
// It is mostly cloudformation.AWSServerlessFunction defenition
// But also mless specific paramters, relation to the containing template
// and cached variables
type Function struct {
	cloudformation.AWSServerlessFunction
	Mless    Mless
	Warnings []string `json:",omitempty"`
	template *Functions
	dir      string // code directory relative to the template file
}

// Mless contain mless specific paratmers for a function
type Mless struct {
	Debugger interface{} `json:",omitempty"` // details of debugger to invoke (nil means no debugger)
}

// New Initialize a new template and read the initial content
func New(fname, desktop string, refresh time.Duration) (*Functions, error) {
	baseDir := path.Dir(fname)

	switch info, err := os.Stat(baseDir); {
	case err != nil:
		return nil, fmt.Errorf("Template directory does not exist: %s", err)
	case !info.IsDir():
		return nil, fmt.Errorf("Template 'directory' is not a dirirectory: %s", baseDir)
	}

	template := &Functions{
		fname:      fname,
		baseDir:    baseDir,
		desktopDir: desktop,
	}

	if refresh > 0 {
		template.refreshTicker = time.NewTicker(refresh)
		go func() {
			for range template.refreshTicker.C {
				template.Refresh()
			}
		}()
	}

	return template, template.Refresh()
}

// Get a function defentition by name (or nil if the function does not exit)
func (t *Functions) Get(name string) *Function {
	t.lock.Lock()
	defer t.lock.Unlock()

	if f, ok := t.functions[name]; ok {
		return &f
	}
	return nil
}

// Close cleanup up after the tempalte (stop automatic refresh)
func (t *Functions) Close() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.refreshTicker != nil {
		t.refreshTicker.Stop()
		t.refreshTicker = nil
	}
	return nil
}

// Refresh attempt to re-read the templates
func (t *Functions) Refresh() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	funcs, mod, err := t.read(t.modTime)
	if err == errSame {
		// log15.Debug("Refresh - same file")
		return nil
	}

	if !mod.IsZero() {
		t.functions = funcs
		t.modTime = mod
	}

	if err != nil {
		fmt.Println("Error reading template: ", err)
	} else {
		fmt.Println("Read template")
		fmt.Println(ldebug.DumpJSON(t.functions))
	}
	return err
}

var errSame = errors.New("same file")

func (t *Functions) read(prev time.Time) (map[string]Function, time.Time, error) {
	info, err := os.Stat(t.fname)
	if err != nil {
		return nil, time.Time{}, err
	}
	modTime := info.ModTime()
	if prev == modTime {
		return nil, time.Time{}, errSame
	}

	cfTemplate, err := goformation.Open(t.fname)
	if err != nil {
		return nil, modTime, err
	}

	funcs := make(map[string]Function)

	for name, f := range cfTemplate.GetAllAWSServerlessFunctionResources() {
		fn := Function{
			AWSServerlessFunction: f,
			template:              t,
		}

		fn.fix(name)

		if mless, err := loadMlessParams(cfTemplate, name); err == nil {
			fn.Mless = *mless
		} else {
			fn.warn(err)
		}

		if confProcessor != nil {
			if err := confProcessor(&fn); err != nil {
				fn.warn(err)
			}
		}
		funcs[name] = fn
	}
	return funcs, modTime, nil
}

func (f *Function) fix(name string) {
	f.FunctionName = name
	if f.CodeUri != nil && f.CodeUri.String != nil {
		f.dir = *f.CodeUri.String
	}

	switch info, err := os.Stat(f.Code()); {
	case err != nil:
		f.warn(fmt.Errorf("no code for function: %s", err))
	case !info.IsDir():
		f.warn(fmt.Errorf("code directory is not a directory: %s", f.Code())) // TODO: should support compressed archive
	}

	// TODO: verify runtime
}

// Code is the directory where the code reside
func (f *Function) Code() string {
	return path.Join(f.template.baseDir, f.dir)
}

// Desktop return the directory in the desktop where we have the sourcecode
func (f *Function) Desktop() string {
	return path.Join(f.template.desktopDir, f.dir) // TODO: will not work with windows desktop
}

func (f *Function) warn(w error) {
	log15.Error("Adding warning", "w", w)
	f.Warnings = append(f.Warnings, w.Error())
}

func loadMlessParams(template *cloudformation.Template, name string) (*Mless, error) {
	rawFunc, ok := template.Resources[name]
	if !ok {
		return nil, nil
	}

	var f struct {
		Properties struct {
			Mless Mless
		}
	}
	if err := mapstructure.Decode(rawFunc, &f); err != nil {
		return nil, err
	}
	return &f.Properties.Mless, nil
}
