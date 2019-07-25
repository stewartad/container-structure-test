package singolang

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"os"
	"log"
)

// CopyTarball creates a Tar archive of a directory or file and places inside the container and places it in /tmp.
// It returns the path to the output file, and a reader for the archive
func (i *Instance) CopyTarball(path string) (string, *tar.Reader, error) {
	// Set up filepaths
	path = filepath.Clean(path)
	file := filepath.Base(path)
	parentDir := filepath.Dir(path)
	dir := filepath.Join(os.TempDir(), i.Name)

	// Make temp directory
	Mkdirp(dir)
	
	archivePrefix := filepath.Base(parentDir)
	if archivePrefix == "" || archivePrefix == "/" {
		archivePrefix = "root"
	} else if archivePrefix == "." {
		archivePrefix = "pwd"
	}

	archivePath := filepath.Join(dir, fmt.Sprintf("%s-archive.tar.gz", archivePrefix))

	opts := ExecOptions {
		Pwd: "",
		Quiet: true, // set to false for debugging info
		Cleanenv: true,
		Env: DefaultEnvOptions(),
	}

	if !opts.Quiet {
		log.Println(archivePath)
	}

	// Create archive
	cmd := []string{"tar", "-C", parentDir, "-czf", archivePath, file}
	_, _, _, err := i.Execute(cmd, &opts)
	if err != nil {
		// Always return archivePath so the archive can still be deleted in the case of an error
		return archivePath, nil, err
	}

	// Create reader for archive
	b, err := ioutil.ReadFile(archivePath)

	if err != nil {
		log.Printf("Could not read file %s", err)
		return archivePath, nil, err
	}
	gzr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		log.Println("READ ERROR")
		return archivePath, nil, err
	}

	
	return archivePath, tar.NewReader(gzr), nil
}
