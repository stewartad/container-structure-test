package singolang

import (
	"fmt"
)

type instanceError struct {
	name   string
	action string
}

func (e *instanceError) Error() string {
	return fmt.Sprintf("Error %sing Instance: %s. Must be stopped manually in Singularity", e.name, e.action)
}

// Client is a struct to hold information about the current client
type Client struct {
	Instances map[string]*Instance
	Sudo      bool // either everything or nothing you do is sudo
	Cleanenv  bool
}

// NewClient creates and returns a new client as well as a teardown function.
// Assign this teardown function and defer it to exit cleanly
func NewClient() (*Client, func(c *Client)) {
	return &Client{
			Instances: make(map[string]*Instance),
			Sudo:      false,
			Cleanenv:  true,
		},
		func(c *Client) { c.StopAllInstances() }
}

// TODO: Clean this up
func (c *Client) String() string {
	baseClient := "[singularity-golang]"

	return baseClient
}

// NewInstance creates a new instance and adds it to the client, if it is able to be started
func (c *Client) NewInstance(image string, name string, env *EnvOptions) (*Instance, error) {
	i := getInstance(image, name, c.Sudo)
	i.EnvOpts = env
	i.RetrieveEnv()
	i.RetrieveLabels()
	i.EnvOpts.ProcessEnvVars()
	c.Instances[name] = i
	return i, nil
}

// StopAllInstances stops all instances created in the client
func (c *Client) StopAllInstances() error {
	var err error
	for _, v := range c.Instances {
		err = v.Stop()
	}
	return err
}

// ListAllInstances lists all currently running Singularity instances.
// It is equivalent to running `singularity instance list` on the host system
func ListAllInstances() string {
	cmd := initCommand("instance", "list")

	output, _, _, err := runCommand(cmd, defaultRunCommandOptions())
	if err != nil {
		return ""
	}
	return string(output.Bytes())
}
