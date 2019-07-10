package client

import (
	"fmt"
	"strings"
)

// ExecOptions provide flags simulating options int he singularity command line
type ExecOptions struct {
	pwd   string
	quiet bool
}

// DefaultExecOptions provides a default options struct
func DefaultExecOptions() *ExecOptions {
	return &ExecOptions{
		pwd:   "",
		quiet: true,
	}
}

type existError struct {
	instance string
}

func (e *existError) Error() string {
	return fmt.Sprintf("%s is not a loaded instance", e.instance)
}

// Execute runs a command inside a container
func (c *Client) Execute(instance string, command []string, opts *ExecOptions) (string, string, int, error) {
	// TODO: check install

	cmd := initCommand("exec")

	_, exists := c.instances[instance]
	if !exists {
		return "", "", -1, &existError{instance}
	}

	// TODO: bind paths

	if opts.pwd != "" {
		cmd = append(cmd, "--pwd", opts.pwd)
	}

	var image string
	if !strings.HasPrefix(instance, "instance://") {
		image = strings.Join([]string{"instance://", instance}, "")
	} else {
		image = instance
	}

	cmd = append(cmd, image)
	cmd = append(cmd, command...)

	stdout, stderr, status, err := runCommand(cmd, &runCommandOptions{
		sudo:     c.Sudo,
		quieterr: opts.quiet,
		quietout: opts.quiet,
	})
	// TODO: use status
	_ = status
	if err != nil {
		return string(stdout.Bytes()), string(stderr.Bytes()), -1, err
	}

	return string(stdout.Bytes()), string(stderr.Bytes()), 0, nil

}
