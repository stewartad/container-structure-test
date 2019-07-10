package client

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// CopyTarball creates a Tar archive of a directory or file and places it in /tmp.
// It returns the path to the archive, and a reader for the archive
func (c *Client) CopyTarball(instance string, path string) (string, *tar.Reader, error) {
	// Make directory for archive and set up filepath
	path = filepath.Clean(path)
	file := filepath.Base(path)
	parentDir := filepath.Dir(path)
	// fmt.Printf("%s\t%s\t%s\n", path, file, parentDir)
	dir := fmt.Sprintf("/tmp/%s", instance)
	Mkdirp(dir)
	archivePath := fmt.Sprintf("%s/%s-archive.tar.gz", dir, filepath.Base(parentDir))

	// Create archive
	cmd := []string{"tar", "-C", parentDir, "-czvf", archivePath, file}
	_, _, code, err := c.Execute(instance, cmd, DefaultExecOptions())
	if err != nil || code != 0 {
		return "", nil, err
	}

	// Create reader for archive
	b, err := ioutil.ReadFile(archivePath)
	if err != nil {
		panic(fmt.Sprintf("Could not read file %s", err))
	}
	gzr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		panic("READ ERROR")
	}

	return archivePath, tar.NewReader(gzr), nil
}
