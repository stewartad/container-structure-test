package drivers

import (
	"os"
	"fmt"
	"path/filepath"
	"archive/tar"
	"io"
	"bytes"
	"bufio"
	"strings"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"

	"github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	_"github.com/GoogleContainerTools/container-structure-test/pkg/utils"

	singularity "github.com/stewartad/singolang"
)

type SingularityDriver struct {
	originalImage 	string
	currentImage 	string
	currentInstance	*singularity.Instance
	cli 			singularity.Client
	env				map[string]string
	save			bool
	runtime			string
}

func NewSingularityDriver(args DriverConfig) (Driver, error) {
	newCli, teardown := singularity.NewClient()
	_ = teardown
	instance, err := newCli.NewInstance(args.Image, "testing-base", singularity.DefaultEnvOptions())
	if err != nil {
		return &SingularityDriver{}, nil
	}
	// instance.Start(newCli.Sudo)

	return &SingularityDriver{
		originalImage:	args.Image,
		currentImage:	args.Image,
		currentInstance: instance,
		cli:			*newCli,
		env:			nil,
		save:			args.Save,
		runtime:		args.Runtime,
	}, nil
}

func (d *SingularityDriver) Setup(envVars []unversioned.EnvVar, fullCommands[][]string) error {
	return nil
}

func (d *SingularityDriver) Teardown(fullCommands [][]string) error {
	return nil
}

func (d *SingularityDriver) SetEnv(envVars []unversioned.EnvVar) error {
	env := d.processEnvVars(envVars)
	container, err := d.cli.NewInstance(d.currentImage, "testing-current", &singularity.EnvOptions{
		EnvVars: convertSliceToMap(env),
		AppendPath: []string{},
		PrependPath: []string{},
		ReplacePath: "",
	})
	if err != nil {
		return errors.Wrap(err, "Error creating container")
	}
	// container.Start(d.cli.Sudo)
	d.currentInstance.Stop(d.cli.Sudo)
	d.currentInstance = container
	return nil
}

func (d *SingularityDriver) ProcessCommand(envVars []unversioned.EnvVar, fullCommand []string) (string, string, int, error) {
	var env []string
	for _, envVar := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", envVar.Key, envVar.Value))
	}
	
	stdout, stderr, exitCode, err := d.exec(env, fullCommand)
	if err != nil {
		return "", "", -1, err
	}

	if stdout != "" {
		logrus.Infof("stdout:\n%s", stdout)
	}
	if stderr != "" {
		logrus.Infof("stderr:\n%s", stderr)
	}
	return stdout, stderr, exitCode, nil
}

func (d *SingularityDriver) exec(env []string, command []string) (string, string, int, error) {

	sudo := d.cli.Sudo
	d.currentInstance.Start(sudo)
	defer d.currentInstance.Stop(sudo)

	opts := singularity.DefaultExecOptions()
	opts.Env = &singularity.EnvOptions{
		EnvVars: convertSliceToMap(env),
	}

	stdout, stderr, code, err := d.currentInstance.Execute(command, opts, sudo)
	return stdout, stderr, code, err
}

func (d *SingularityDriver) retrieveTar(target string) (*tar.Reader, error, func()) {
	// instanceName := "testing"
	// _, err := d.cli.NewInstance(d.currentImage, instanceName, singularity.DefaultEnvOptions())
	// if err != nil {
	// 	return nil, err, func() {}
	// }
	// defer d.cli.StopInstance(instanceName)
	sudo := d.cli.Sudo
	d.currentInstance.Start(sudo)
	defer d.currentInstance.Stop(sudo)

	t, read, err := d.currentInstance.CopyTarball(target)
	if err != nil {
		return nil, err, func() {}
	}
	// defer os.RemoveAll(filepath.Dir(t))

	return read, nil, func() {
		os.RemoveAll(filepath.Dir(t))
	}
}

func (d *SingularityDriver) StatFile(path string) (os.FileInfo, error) {
	read, err, cleanup := d.retrieveTar(path)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	for {
		head, err := read.Next()
		if err == io.EOF {
			break
		}

		/*
		* BEGIN FILE LOGIC HERE
		* EVERYTHING ELSE IS BOILER PLATE
		*/
		switch head.Typeflag {
		case tar.TypeDir, tar.TypeReg, tar.TypeLink, tar.TypeSymlink:
			
			if filepath.Clean(head.Name) == filepath.Base(path) {
				return head.FileInfo(), nil
			}
		default:
			continue
		}
		/*
		 * END FILE LOGIC
		 */
	}

	return nil, fmt.Errorf("File %s not found in image", path)
}

func (d *SingularityDriver) ReadFile(path string) ([]byte, error) {
	read, err, cleanup := d.retrieveTar(path)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	for {
		head, err := read.Next()
		if err == io.EOF {
			break
		}

		/*
		* BEGIN FILE LOGIC HERE
		* EVERYTHING ELSE IS BOILER PLATE
		*/
		switch head.Typeflag {
		case tar.TypeDir:
			if filepath.Clean(head.Name) == filepath.Base(path) {
				return nil, fmt.Errorf("Cannot read specified path: %s is a directory, not a file", path)
			}
		case tar.TypeSymlink:
			return d.ReadFile(head.Linkname)
		case tar.TypeReg, tar.TypeLink:
			if filepath.Clean(head.Name) == filepath.Base(path) {
				var b bytes.Buffer
				stream := bufio.NewWriter(&b)
				io.Copy(stream, read)
				return b.Bytes(), nil
			}
		default:
			continue
		}
		/*
		 * END FILE LOGIC
		 */
	}

	return nil, fmt.Errorf("File %s not found in image", path)
}

func (d *SingularityDriver) ReadDir(path string) ([]os.FileInfo, error) {
	read, err, cleanup := d.retrieveTar(path)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	var infos []os.FileInfo
	for {
		header, err := read.Next()
		if err == io.EOF {
			break
		}
		if header.Typeflag == tar.TypeDir {
			// we only want top level dirs here, no recursion. to get these, remove
			// trailing separator and split on separator. there should only be two parts.
			parts := strings.Split(strings.TrimSuffix(header.Name, string(os.PathSeparator)), string(os.PathSeparator))
			if len(parts) == 2 {
				infos = append(infos, header.FileInfo())
			}
		}
	}

	return infos, nil
}

func (d *SingularityDriver) GetConfig() (unversioned.Config, error) {
	return unversioned.Config{
		Env: d.currentInstance.GetEnv().EnvVars,
		Labels: d.currentInstance.ImgLabels,
	}, nil
}

func (d *SingularityDriver) Destroy() {
	d.cli.StopAllInstances()
}

// returns a func that consumes a string, and returns the value associated with
// that string when treated as a key in the image's environment.
func retrieveSingularityEnv(d *SingularityDriver) func(string) string {
	return func(envVar string) string {
		var env map[string]string
		if env == nil {
			image := d.currentInstance
			// convert env to map for processing
			env = image.EnvOpts.EnvVars
		}
		return env[envVar]
	}
}

// returns the value associated with the provided key in the image's environment
func (d *SingularityDriver) retrieveEnvVar(envVar string) string {
	// since we're only retrieving these during processing, we can use a closure to cache this
	return retrieveSingularityEnv(d)(envVar)
}

// given a list of env vars, return a new list with each var's value appended to it
// in the form 'key==val'. we do this because docker expects them to be passed this way.
func (d *SingularityDriver) processEnvVars(vars []unversioned.EnvVar) []string {
	if len(vars) == 0 {
		return nil
	}

	env := []string{}

	for _, envVar := range vars {
		expandedVal := os.Expand(envVar.Value, d.retrieveEnvVar)
		env = append(env, fmt.Sprintf("%s=%s", envVar.Key, expandedVal))
	}
	return env
}