package singolang

import (
	"fmt"
	"os"
	"strings"
)

const senv string = "SINGULARITYENV"

// EnvOptions configures the environment with which to run the container
type EnvOptions struct {
	EnvVars map[string]string
	PrependPath []string
	AppendPath []string
	ReplacePath	string
}

// DefaultEnvOptions gives a blank env
func DefaultEnvOptions() *EnvOptions {
	return &EnvOptions {
		EnvVars: make(map[string]string),
		PrependPath: []string{},
		AppendPath: []string{},
	}
}

// UpdateEnv clears all 
func (opts *EnvOptions) UpdateEnv() error {
	opts.unsetAll()
	err := opts.ProcessEnvVars()
	return err
}

// ClearEnv will unset all environment variables which contain "SINGULARITYENV_". 
// It iterates over all environment variables and as such is costly
func ClearEnv() error {
	var err error
	for _, v := range os.Environ() {
		k := strings.Split(v, "=")
		if strings.Contains(k[0], "SINGULARITYENV_") {
			err = os.Unsetenv(k[0])
		}
	}
	return err
}

func (opts *EnvOptions) ProcessEnvVars() error {
	var err error
	for k, v := range opts.EnvVars {
		varName := fmt.Sprintf("%s_%s", senv, k)
		err = os.Setenv(varName, v)
	}
	return err
}

func (opts *EnvOptions) processPathMod() error {
	var err error
	for _, prepath := range opts.PrependPath {
		err = prependPath(prepath)
	}
	for _, apppath := range opts.AppendPath {
		err = appendPath(apppath)
	} 
	return err
}

func prependPath(path string) error {
	err := os.Setenv(fmt.Sprintf("%s_PREPEND_PATH", senv), path)
	return err
}

func appendPath(path string) error {
	err := os.Setenv(fmt.Sprintf("%s_APPEND_PATH", senv), path)
	return err
}

func (opts *EnvOptions) unsetAll() {
	for k := range opts.EnvVars {
		varName := fmt.Sprintf("%s_%s", senv, k)
		os.Unsetenv(varName)
	}
	os.Unsetenv(fmt.Sprintf("%s_PREPEND_PATH", senv))
	os.Unsetenv(fmt.Sprintf("%s_APPEND_PATH", senv))
}