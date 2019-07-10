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

	"github.com/sirupsen/logrus"

	"github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	_"github.com/GoogleContainerTools/container-structure-test/pkg/utils"

	singularity "github.com/stewartad/singolang/client"
)

type SingularityDriver struct {
	originalImage 	string
	currentImage 	string
	cli 			singularity.Client
	env				map[string]string
	save			bool
	runtime			string
}

func NewSingularityDriver(args DriverConfig) (Driver, error) {
	newCli, teardown := singularity.NewClient()
	_ = teardown

	return &SingularityDriver{
		originalImage:	args.Image,
		currentImage:	args.Image,
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
	// TODO: process env variables
	instanceName := "testing"
	err := d.cli.NewInstance(d.currentImage, instanceName)
	if err != nil {
		return "", "", -1, err
	}
	defer d.cli.StopInstance(instanceName)

	stdout, stderr, code, err := d.cli.Execute(instanceName, command, singularity.DefaultExecOptions())
	return stdout, stderr, code, err
}

func (d *SingularityDriver) retrieveTar(target string) (*tar.Reader, error, func()) {
	instanceName := "testing"
	err := d.cli.NewInstance(d.currentImage, instanceName)
	if err != nil {
		return nil, err, func() {}
	}
	defer d.cli.StopInstance(instanceName)
	t, read, err := d.cli.CopyTarball(instanceName, target)
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
	return unversioned.Config{}, nil
}

func (d *SingularityDriver) Destroy() {

}