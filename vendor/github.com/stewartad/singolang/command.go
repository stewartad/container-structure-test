package singolang

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

// InitCommand creates a slice of strings of which the first element is "singularity" followed by all args
func initCommand(args ...string) []string {
	cmd := []string{"singularity"}
	cmd = append(cmd, args...)
	// append quiet or debug if flags are set in client
	return cmd
}

type runCommandOptions struct {
	sudo     bool
	quietout bool
	quieterr bool
}

//
func defaultRunCommandOptions() *runCommandOptions {
	return &runCommandOptions{
		sudo:     false,
		quietout: true,
		quieterr: true,
	}
}

/*RunCommand runs a terminal command
cmd - a slice of strings, of which the first element must be the command name
	all subsequent elements are the arguments
opts - runCommandOptions struct defining options to be used
*/
func runCommand(cmd []string, opts *runCommandOptions) (bytes.Buffer, bytes.Buffer, int, error) {
	// add sudo to front of command if requested
	if opts.sudo {
		cmd = append([]string{"sudo"}, cmd...)
	}
	name := cmd[0]

	

	// create command instance
	process := exec.Command(name, cmd[1:]...)

	// buffers to store stdout and stderr to be returned
	var stdoutBuf, stderrBuf bytes.Buffer

	// get stdout and stderr pipes
	stdout, err := process.StdoutPipe()
	if err != nil {
		log.Printf("Error getting stdout: %s\n", err)
	}
	stderr, err := process.StderrPipe()
	if err != nil {
		log.Printf("Error getting stderr: %s\n", err)
	}

	process.Start()

	var errStdout, errStderr error
	var outWriter, errWriter io.Writer
	if opts.quietout {
		// only write stdout to buffer, not to screen
		outWriter = io.Writer(&stdoutBuf)
	} else {
		// write stdout to both buffer and screen
		outWriter = io.MultiWriter(os.Stdout, &stdoutBuf)
	}
	if opts.quieterr {
		// only write stderr to buffer, not to screen
		errWriter = io.Writer(&stderrBuf)
	} else {
		// write stderr to both buffer and screen
		errWriter = io.MultiWriter(os.Stderr, &stderrBuf)
	}

	// capture stdout concurrently
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(outWriter, stdout)
		wg.Done()
	}()

	// wait for stdout to be captured, then capture stderr
	_, errStderr = io.Copy(errWriter, stderr)
	wg.Wait()

	var waitStatus syscall.WaitStatus

	// wait for command to exit
	err = process.Wait()
	// handle erros
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			if !opts.quieterr {	
				log.Printf("cmd: %s\n", cmd)
				log.Printf("stdout: %s", string(stdoutBuf.Bytes()))
				log.Printf("stderr: %s", string(stderrBuf.Bytes()))
				log.Printf("Command failed with %s\n", err)
			}
			return stdoutBuf, stderrBuf, waitStatus.ExitStatus(), err
		}
	}
	
	waitStatus = process.ProcessState.Sys().(syscall.WaitStatus)
	if errStdout != nil || errStderr != nil {
		log.Printf("Failed to capture strout or stderr")
	}

	// return stdout and stderr as strings
	return stdoutBuf, stderrBuf, waitStatus.ExitStatus(), err
}