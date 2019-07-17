package singolang

import (
	"fmt"
	"strings"
)

// ExecOptions provide flags simulating options int he singularity command line
type ExecOptions struct {
	Pwd   	string
	Quiet	bool
	Cleanenv bool
	Env		*EnvOptions
}



// DefaultExecOptions provides a default options struct
func DefaultExecOptions() *ExecOptions {
	return &ExecOptions{
		Pwd:   "",
		Quiet: true,
		Cleanenv: true,
		Env: DefaultEnvOptions(),
	}
}

type existError struct {
	instance string
}

func (e *existError) Error() string {
	return fmt.Sprintf("%s is not a loaded instance", e.instance)
}

// Execute runs a command inside a container
func (i *Instance) execute(command []string, opts *ExecOptions, sudo bool) (string, string, int, error) {
	// TODO: check install

	cmd := initCommand("exec")
	instance := i.name

	// TODO: bind paths

	if opts.Cleanenv {
		cmd = append(cmd, "--cleanenv")
	}
	if opts.Pwd != "" {
		cmd = append(cmd, "--pwd", opts.Pwd)
	}

	var image string
	if !strings.HasPrefix(instance, "instance://") {
		image = strings.Join([]string{"instance://", instance}, "")
	} else {
		image = instance
	}

	err := opts.Env.ProcessEnvVars()
	if err != nil {
		return "", "", -1, err
	}
	err = opts.Env.processPathMod()
	if err != nil {
		return "", "", -1, err
	}

	// deferred functions execute LIFO
	defer i.EnvOpts.ProcessEnvVars()
	defer opts.Env.unsetAll()
	
	cmd = append(cmd, image)
	cmd = append(cmd, command...)

	fmt.Println(cmd)

	stdout, stderr, status, err := runCommand(cmd, &runCommandOptions{
		sudo:     sudo,
		quieterr: opts.Quiet,
		quietout: opts.Quiet,
	})
	// TODO: use status
	_ = status
	if err != nil {
		return string(stdout.Bytes()), string(stderr.Bytes()), -1, err
	}

	return string(stdout.Bytes()), string(stderr.Bytes()), 0, nil

}