package singolang

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func load() {

}

func setenv() {

}

// GetFilename creates the filename of an image being pulled
func GetFilename(image string, ext string) string {
	if ext == "" {
		ext = "sif"
	}

	image = filepath.Base(image)
	match, err := regexp.Compile("^.*//")
	if err != nil {
		// TODO: Better error handling
		log.Fatalf("%s\n", "bad")
	}
	image = match.ReplaceAllString(image, "")

	if !strings.HasSuffix(image, ext) {
		if strings.HasPrefix(ext, ".") {
			image = strings.Join([]string{image, ext}, "")
		} else {
			image = strings.Join([]string{image, ext}, ".")
		}
	}

	return image
}

func getURI() {

}

// Mkdirp simulates mkdir -p UNIX command
func Mkdirp(path string) {
	err := os.MkdirAll(path, 0777)
	if err != nil {
		log.Fatalf("Error creating path %s, exiting.", path)
	}
}

// GetSingularityVersion gets installed Singularity version
func GetSingularityVersion() string {
	version, _, status, _ := runCommand([]string{"singularity", "--version"}, defaultRunCommandOptions())
	// log.Println(status)
	if status == 0 {
		return strings.TrimSpace(string(version.Bytes()))
	}
	return ""
}

// SplitURI splits the URI into protocol and path"
func SplitURI(container string) (string, string) {
	// Splits
	parts := strings.Split(container, "://")
	if len(parts) == 1 {
		return "", parts[0]
	}
	return parts[0], parts[1]
}

// RemoveURI strips the protocol and returns only the path
func RemoveURI(container string) string {
	_, path := SplitURI(container)
	return path
}
