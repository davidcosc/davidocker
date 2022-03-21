package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
)

// CreateNetworkNamespace runs a new process with unshared network namespace.
// After the spawned networkNamespaceCreated process finishes, program flow contiunues
// in this parent process.
var CreateNetworkNamespace = func() error {
	fmt.Println("* Run new process in new network namespace....")
	cmd := exec.Command("/proc/self/exe", "networkNamespaceCreated")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNET,
	}
	return cmd.Run()
}

// FinalizeNetworkNamespace bind mounts the network namespace to a file inside the
// container directory. This way the network namespace will persist even after the
// process terminates. This allows for veth interface tunnels to be setup from
// within the host network namespace, linking the host and bind mounted network namespace.
// The mount is flagged MS_SHARED, to allow the container to remove it globally
// after joining the network namespace later on. Once this is done, the lifetime
// of the network namespace ist bound to the lifetime of the containerized process.
var FinalizeNetworkNamespace = func(dir string) error {
	fmt.Println("FinalizeNetworkNamespace......................")
	netFD, err := os.Create(path.Join(dir, "net"))
	defer netFD.Close()
	if err != nil {
		return err
	}
	return syscall.Mount("/proc/self/ns/net", path.Join(dir, "net"), "bind", syscall.MS_BIND|syscall.MS_SHARED, "")
}

// joinNetworkNamespace is intended to make the containerized process join the previously
// set up network namespace. It also attempts to cleanup the now obsolete network bind mount.
// After cleanup the lifetime of the network namespace is bound to the lifetime of the
// containerized process.
var joinNetworkNamespace = func(dir string) error {
	fmt.Println("* Opening network namespace mount.............")
	netFD, err := syscall.Open(path.Join(dir, "net"), syscall.O_RDONLY, 0644)
	if err != nil {
		err = syscall.Unmount(path.Join(dir, "net"), 0)
		return err
	}
	fmt.Println("* Joining network namespace...................")
	// 308 is trap code for setns syscall
	_, _, errNo := syscall.RawSyscall(308, uintptr(netFD), 0, 0)
	if errNo != 0 {
		err = syscall.Close(netFD)
		err = syscall.Unmount(path.Join(dir, "net"), 0)
		return errNo
	}
	err = syscall.Close(netFD)
	err = syscall.Unmount(path.Join(dir, "net"), 0)
	fmt.Println("* Removing network namespace bind mount.......")
	err = os.Remove(path.Join(dir, "net"))
	return err
}
