package containercreator

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestCreateContainer(t *testing.T) {
	if len(os.Args) > 1 && os.Args[1] == "containerNamespacesCreated" {
		t.Skip("Process is namespaced")
	}
	// given
	containerCreator := &ContainerCreatorImpl{}
	// when
	// /proc/self/exe within CreateContainerNamespaces points to the compiled test binary
	// since all tests to be run are compiled within no additional args need to be passed
	// the containerNamespacesCreated arg is appended as test binary arg, but has no effect
	err := containerCreator.CreateContainerNamespaces([]string{})
	assert.Equal(t, nil, err, "should be equal")
	time.Sleep(1 * time.Second)
	actualContainerStdout, err := os.ReadFile("/root/container/stdout")
	// then
	assert.Equal(t, nil, err, "should be equal")
	expectedContainerStdout := `Finalizing container..............
* Setting hostname................
* Chrooting.......................
* Changing working directory......
* Creating /proc dir if not exist.
* Mounting proc...................
PID: 1 Hostname: container
bin
lib
lib64
proc
stderr
stdin
stdout
`
	assert.Equal(t, expectedContainerStdout, string(actualContainerStdout), "should be equal")
}

func TestFinalizeContainer(t *testing.T) {
	if len(os.Args) != 2 || os.Args[1] != "containerNamespacesCreated" {
		t.Skip("Process is not namespaced")
	}
	// given
	containerCreator := &ContainerCreatorImpl{}
	// when
	containerCreator.FinalizeContainer([]string{"/bin/ls"})
}
