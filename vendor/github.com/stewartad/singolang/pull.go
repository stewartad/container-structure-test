package singolang

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PullOptions provide a set of options to configure the pull command.
//
type PullOptions struct {
	Name       string
	Pullfolder string
	Force      bool
}

type pullError struct {
	msg string
}

func (e *pullError) Error() string {
	return e.msg
}

// Pull pulls image from singularityhub or dockerhub and builds it.
// It stores the image in pullfolder, naming it name.ext
func (c *Client) Pull(image string, opts *PullOptions) (string, error) {
	cmd := initCommand("pull")

	if opts.Force {
		cmd = append(cmd, "-F")
	}

	match, err := regexp.MatchString("^(shub|docker)://", image)
	if err != nil {
		return "", err
	}
	if !match {
		return "", &pullError{msg: "Pull only valid for singularity hub and docker hub"}
	}

	name := opts.Name

	if opts.Name == "" {
		name = GetFilename(image, "")
	}
	if opts.Pullfolder != "" {
		name = filepath.Join(opts.Pullfolder, name)
	}

	cmd = append(cmd, name)
	cmd = append(cmd, image)

	fmt.Printf("%s\n", strings.Join(cmd, " "))

	runCommand(cmd, defaultRunCommandOptions())

	finalImage := filepath.Join(opts.Pullfolder, filepath.Base(name))

	_, err = os.Stat(finalImage)
	if err != nil {
		return "", err
	}
	fmt.Println(finalImage)
	return finalImage, nil
}
