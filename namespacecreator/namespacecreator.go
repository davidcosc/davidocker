package namespacecreator

import (
	"os"
	"os/exec"
	"syscall"
)

type NamespaceCreator interface {
	CreateNamespaces() (int, string, error)
}

type NamespaceCreatorImpl struct{}

func (namespaceCreatorImpl *NamespaceCreatorImpl) CreateNamespaces() (int, string, error) {
	err := syscall.Unshare(syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS)
	if err != nil {
		return -1, "", err
	}
	// err = syscall.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		return -1, "", err
	}
	err = syscall.Sethostname([]byte("container"))
	if err != nil {
		return -1, "", err
	}
	hostname, err := os.Hostname()
	pid := os.Getpid()
	cmd := exec.Command("/lsns")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	return pid, hostname, err
}
