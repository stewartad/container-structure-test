package drivers

import (
	"os"
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/GoogleContainerTools/container-structure-test/pkg/types/unversioned"
	_ "github.com/GoogleContainerTools/container-structure-test/pkg/utils"

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
		logrus.Infof("stdout: %s", stdout)
	}
	if stderr != "" {
		logrus.Infof("stderr: %s", stderr)
	}
	return stdout, stderr, exitCode, nil
}

func (d *SingularityDriver) exec(env []string, command []string) (string, string, int, error) {
	instanceName := "testing"
	err := d.cli.NewInstance(d.currentImage, instanceName)
	if err != nil {
		return "", "", -1, err
	}
	defer d.cli.StopInstance(instanceName)

	stdout, stderr, code, err := d.cli.Execute(instanceName, command, true)
	return stdout, stderr, code, err
}

func (d *SingularityDriver) StatFile(path string) (os.FileInfo, error) {
	return nil, nil
}

func (d *SingularityDriver) ReadFile(path string) ([]byte, error) {
	return nil, nil
}

func (d *SingularityDriver) ReadDir(path string) ([]os.FileInfo, error) {
	return nil, nil
}

func (d *SingularityDriver) GetConfig() (unversioned.Config, error) {
	return unversioned.Config{}, nil
}

func (d *SingularityDriver) Destroy() {

}