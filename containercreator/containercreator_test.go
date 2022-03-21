package containercreator

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

// TestCreateContainer checks if the functions for setting up and running a process
// inside a container work correctly.
// Containerizing the process requires to restart the current binary inside a new process.
// The new process will be namespaced. The current binary is pointed to by /proc/self/exe
// and called inside the CreateContainerNamespaces function. In context of this test it points
// to the compiled test binary.
// We must ensure, that test functions intended for use with the new test process can be
// differentiated from test functions intended for use in the initial test process.
// Therefore depending on the test process, some test functions will be skipped.
// Tests are skipped depending on the cmd args passed to the test binary call.
func TestCreateContainer(t *testing.T) {
	if len(os.Args) < 3 {
		t.Skip("Process is namespaced")
	}
	// given
	cmd := []string{}
	// when
	err := CreateNetworkNamespace()
	assert.Equal(t, nil, err, "should be equal")
	err = CreateContainerNamespaces(cmd)
	assert.Equal(t, nil, err, "should be equal")
	time.Sleep(1 * time.Second)
	actualContainerStdout, err := os.ReadFile("/root/container/stdout")
	// then
	assert.Equal(t, nil, err, "should be equal")
	expectedContainerStdout := `Finalizing container..........................
* Opening network namespace mount.............
* Joining network namespace...................
* Removing network namespace bind mount.......
* Setting hostname............................
* Override / mount with MS_REC / MS_PRIVATE...
* Creating /proc dir if not exist.............
* Mounting proc...............................
* Chrooting...................................
* Changing working directory..................
* Setting hostname............................
    PID TTY          TIME CMD
      1 ?        00:00:00 ps
`
	assert.Equal(t, expectedContainerStdout, string(actualContainerStdout), "should be equal")
}

func TestFinalizeNetworkNamespace(t *testing.T) {
	if len(os.Args) != 2 || os.Args[1] != "networkNamespaceCreated" {
		t.Skip("Process is not network namespaced")
	}
	// given
	// when
	err := FinalizeNetworkNamespace(CONTAINER_DIR)
	// then
	assert.Equal(t, nil, err, "should be equal")
}

// TestFinalizContainer checks the successful setup of the container after
// the test binary was called as a new process with namespaces isolation.
func TestFinalizeContainer(t *testing.T) {
	if len(os.Args) != 2 || os.Args[1] != "containerNamespacesCreated" {
		t.Skip("Process is not container namespaced")
	}
	// given
	cmd := []string{"/bin/ps"}
	// when
	err := FinalizeContainer(cmd)
	// then
	assert.Equal(t, nil, err, "should be equal")
}
