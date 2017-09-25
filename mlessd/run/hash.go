package run

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
)

func hashDir(dir string) ([]byte, error) {
	var errors errorList
	hash := sha1.New()
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		hash.Write([]byte(path))
		if err != nil {
			fmt.Fprintf(hash, "[%s]", err)
			errors = append(errors, err)
			return nil
		}
		fmt.Fprintf(hash, "[%d:%d:%s]", info.Size(), info.Mode(), info.ModTime())
		return nil
	})
	if err != nil {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		err = errors
	}
	return hash.Sum(nil), err
}

type errorList []error

func (e errorList) Error() string {
	b := bytes.Buffer{}
	for i, err := range e {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(err.Error())
	}
	return b.String()
}
