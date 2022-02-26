package namespacecreator

import (
	"fmt"
	"os"
	//"os/exec"
	"syscall"
)

type NamespaceCreator interface {
	CreateNamespaces() (int, string, error)
}

type NamespaceCreatorImpl struct{}

func (namespaceCreatorImpl *NamespaceCreatorImpl) CreateNamespaces() (int, string, error) {
	// there is no point in unsharing the  PID namespace without forking
	// unsharing the PID namespace does not move the calling process to the new PID namespace
	// but instead moves the next child process
	// to make full use of the PID namespace clone with NEWNS clone and unshare flag
	fmt.Println("Creating namespaces")
	err := syscall.Unshare(syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID)
	if err != nil {
		return -1, "", err
	}
	fmt.Println("Setting hostname")
	err = syscall.Sethostname([]byte("container"))
	if err != nil {
		return -1, "", err
	}
	fmt.Println("Chrooting")
	err = syscall.Chroot("/root/container")
	if err != nil {
		return -1, "", err
	}
	fmt.Println("Changing working directory")
	err = syscall.Chdir("/")
	if err != nil {
		return -1, "", err
	}
	fmt.Println("Creating /proc dir if not exist")
	if _, err := os.Stat("/proc"); os.IsNotExist(err) {
		os.Mkdir("/proc", os.ModeDir)
	}
	fmt.Println("Mounting proc")
	err = syscall.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		return -1, "", err
	}
	fmt.Println("Cleanup mounts")
	err = syscall.Unmount("/proc", 0)
	if err != nil {
		return -1, "", err
	}
	hostname, err := os.Hostname()
	pid := os.Getpid()
	return pid, hostname, err
}
