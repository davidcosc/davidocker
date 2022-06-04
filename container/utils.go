package container

import (
	"fmt"
	"os"
	"path"
)

// prepareCustomStdioDescriptors sets up file descriptors inside the specified directory.
// Stdio of the containerized process will be redirected to these later on.
func prepareCustomStdioDescriptors(dir string) (*os.File, *os.File, *os.File, error) {
	fmt.Println("* Preparing stdio descriptors.................")
	stdin, err := os.Create(path.Join(dir, "stdin"))
	if err != nil {
		return nil, nil, nil, err
	}
	stdout, err := os.Create(path.Join(dir, "stdout"))
	if err != nil {
		return nil, nil, nil, err
	}
	stderr, err := os.Create(path.Join(dir, "stderr"))
	if err != nil {
		return nil, nil, nil, err
	}
	return stdin, stdout, stderr, err
}
