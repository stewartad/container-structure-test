package singolang

import (
	"fmt"
	"strings"
	"log"
)

// Instance holds information about a currently running image instance
type Instance struct {
	Name     string
	ImageURI string
	Protocol string
	Image    string
	Cleanenv bool
	ImgLabels	 map[string]string
	ImgEnvVars	 map[string]string
	EnvOpts	 *EnvOptions
	Sudo	 bool
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
	if i.Protocol != "" {
		return fmt.Sprintf("%s:\\%s", i.Protocol, i.Image)
	}
	return i.Image
}

// GetInstance returns a new Instance with image information
func getInstance(image string, name string, options ...string) *Instance {
	i := new(Instance)
	i.parseImageName(image)

	if name != "" {
		i.Name = name
	}
	i.Cleanenv = true
	i.ImgEnvVars = make(map[string]string)
	i.ImgLabels = make(map[string]string)
	i.EnvOpts = DefaultEnvOptions()
	i.Sudo = false

	i.Options = options
	return i
}

func (i *Instance) IsRunning() bool {
	cmd := []string{"singularity", "instance", "list"}
	opts := runCommandOptions{
		sudo: i.Sudo,
		quietout: true,
		quieterr: true,
	}
	stdout, _, _, err := runCommand(cmd, &opts) 
	if err != nil {
		return false
	}
	name := fmt.Sprintf("%s", i.Name)
	output := string(stdout.Bytes())
	return strings.Contains(output, name)
}

// parseImageName processes the image name and protocol
func (i *Instance) parseImageName(image string) {
	i.ImageURI = image
	i.Protocol, i.Image = SplitURI(image)
}


func (i *Instance) updateEnv() {

}

// Start starts an instance
// Does not support startscript args
func (i *Instance) Start(sudo bool) error {
	cmd := initCommand("instance", "start")

	cmd = append(cmd, i.ImageURI, i.Name)

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
	cmd = append(cmd, i.Name)

	stdout, stderr, status, err := runCommand(cmd, &instanceOpts)
	// TODO: use these
	log.Printf("instance stdout: %s\n", stdout)
	log.Printf("instance stderr: %s\n", stderr)
	_, _, _ = stdout, stderr, status
	return err
}

func (i *Instance) RetrieveLabels() error {
	i.ImgLabels = make(map[string]string)
	cmd := []string{"singularity", "inspect", "--labels", i.Image}
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
	cmd := []string{"singularity", "exec", i.Image, "env"}
	stdout, _, _, err := runCommand(cmd, defaultRunCommandOptions())
	output := string(stdout.Bytes())
	if err != nil {
		return err
	}

	for _, env := range strings.Split(output, "\n") {
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
	m["name"] = i.Name
	m["imageURI"] = i.ImageURI
	m["protocol"] = i.Protocol
	m["image"] = i.Image
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
