// Package envstr encode an environemnt to be used as hash key.
package envstr

import (
	"sort"
	"strings"
)

var replacer = strings.NewReplacer(`\`, `\\`, `=`, `\=`, `;`, `\;`)

// Encode an environment map in a way that is safe to compare,
// to know if the environemnt is the same or not
func Encode(env map[string]string) string {
	list := make([]string, len(env))
	i := 0
	for k, v := range env {
		list[i] = replacer.Replace(k) + "=" + replacer.Replace(v)
		i++
	}
	sort.Strings(list)
	return strings.Join(list, ";")
}
