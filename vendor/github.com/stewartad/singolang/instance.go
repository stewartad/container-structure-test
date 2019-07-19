package singolang

import (
	"fmt"
	"strings"
)

// Instance holds information about a currently running image instance
type Instance struct {
	name     string
	imageURI string
	protocol string
	image    string
	Cleanenv bool
	ImgLabels	 map[string]string
	ImgEnvVars	 map[string]string
	EnvOpts	 *EnvOptions
	cmd      []string
	Options  []string
	Metadata []string // might go unused
}

var instanceOpts = runCommandOptions{
	sudo:     false,
	quietout: false,
	quieterr: false,
}

func (i *Instance) String() string {
	if i.protocol != "" {
		return fmt.Sprintf("%s:\\%s", i.protocol, i.image)
	}
	return i.image
}

// GetInstance returns a new Instance with image information
func getInstance(image string, name string, options ...string) *Instance {
	i := new(Instance)
	i.parseImageName(image)

	if name != "" {
		i.name = name
	}
	i.Cleanenv = true
	i.ImgEnvVars = make(map[string]string)
	i.ImgLabels = make(map[string]string)
	i.EnvOpts = DefaultEnvOptions()

	i.Options = options
	return i
}

// parseImageName processes the image name and protocol
func (i *Instance) parseImageName(image string) {
	i.imageURI = image
	i.protocol, i.image = SplitURI(image)
}


func (i *Instance) updateEnv() {

}

// Start starts an instance
// Does not support startscript args
func (i *Instance) Start(sudo bool) error {
	cmd := initCommand("instance", "start")

	cmd = append(cmd, i.imageURI, i.name)

	if i.Cleanenv {
		cmd = append(cmd, "--cleanenv")
	}

	stdout, stderr, status, err := runCommand(cmd, &instanceOpts)
	// TODO: use these
	_, _, _ = stdout, stderr, status

	return err
}

// Stop stops an instance.
func (i *Instance) Stop(sudo bool) error {
	cmd := initCommand("instance", "stop")
	cmd = append(cmd, i.name)

	stdout, stderr, status, err := runCommand(cmd, &instanceOpts)
	// TODO: use these
	_, _, _ = stdout, stderr, status
	return err
}

func (i *Instance) RetrieveLabels() error {
	i.ImgLabels = make(map[string]string)
	cmd := []string{"singularity", "inspect", "--labels", i.image}
	stdout, _, _, err := runCommand(cmd, defaultRunCommandOptions())

	for _, label := range strings.Split(string(stdout.Bytes()), "\n") {
		v := strings.Split(label, ":")
		if len(v) > 1 {
			i.ImgLabels[v[0]] = v[1]
		}
	}
	return err
}

// RetrieveEnv retrieves all env variables in the instance and stores them in a map
func (i *Instance) RetrieveEnv() error {
	i.ImgEnvVars = make(map[string]string)
	cmd := []string{"env"}
	stdout, _, _, err := i.execute(cmd, DefaultExecOptions(), false)
	
	if err != nil {
		return err
	}

	for _, env := range strings.Split(stdout, "\n") {
		v := strings.Split(env, "=")
		if len(v) > 1 {
			i.ImgEnvVars[v[0]] = v[1]
		}
	}

	return err
}

/*
 * Getters for Instance fields
 */

// GetInfo returns the information about an Instance
func (i *Instance) GetInfo() map[string]string {
	m := make(map[string]string)
	m["name"] = i.name
	m["imageURI"] = i.imageURI
	m["protocol"] = i.protocol
	m["image"] = i.image
	m["cmd"] = strings.Join(i.cmd, " ")
	return m
}

func (i *Instance) SetEnv(opts *EnvOptions) {
	i.EnvOpts = opts
	i.RetrieveEnv()
	i.RetrieveLabels()
	i.EnvOpts.ProcessEnvVars()
}

// GetEnv gets the instance environment
func (i *Instance) GetEnv() *EnvOptions {
	return i.EnvOpts
}

// GetCmd returns a slice of strings that represent the full command created when i.Start() was called.
// This slice can immediately be passed into RunCommand() to be ran again
func (i *Instance) GetCmd() []string {
	return i.cmd
}
