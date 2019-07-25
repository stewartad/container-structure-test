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

// CopyTarball creates a Tar archive of a directory or file and places it in /tmp.
// It returns the path to the archive, and a reader for the archive
func (i *Instance) CopyTarball(path string) (string, *tar.Reader, error) {
	// Make directory for archive and set up filepath
	path = filepath.Clean(path)
	file := filepath.Base(path)
	parentDir := filepath.Dir(path)
	// fmt.Printf("%s\t%s\t%s\n", path, file, parentDir)
	dir := filepath.Join(os.TempDir(), i.Name)
	// dir2 := fmt.Sprintf("/tmp/%s", i.Name)
	// log.Println(dir, dir2)
	Mkdirp(dir)
	
	archivePath := filepath.Join(dir, fmt.Sprintf("%s-archive.tar.gz", filepath.Base(parentDir)))
	// archivePath := fmt.Sprintf("%s/%s-archive.tar.gz", dir, filepath.Base(parentDir))

	opts := ExecOptions {
		Pwd: "",
		Quiet: false,
		Cleanenv: true,
		Env: DefaultEnvOptions(),
	}

	log.Println(archivePath)

	// Create archive
	cmd := []string{"tar", "-C", parentDir, "-czvf", archivePath, file}
	_, _, code, err := i.Execute(cmd, &opts, i.Sudo)
	if err != nil || code != 0 {
		log.Println("Houston we have a problem")
		return "", nil, err
	}

	// Create reader for archive
	b, err := ioutil.ReadFile(archivePath)
	tarFile, tarErr := os.Open(archivePath)
	if tarErr != nil {
		log.Println("Error reading tar")
	}
	fi, _ := tarFile.Stat()
	fmt.Println(fi.Mode)
	if err != nil {
		panic(fmt.Sprintf("Could not read file %s", err))
	}
	gzr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		panic("READ ERROR")
	}

	
	return archivePath, tar.NewReader(gzr), nil
}
