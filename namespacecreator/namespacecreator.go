package namespacecreator

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type NamespaceCreator interface {
	CreateNamespaces(cmdArgs []string) error
	FinalizeNamespaces() error
}

type NamespaceCreatorImpl struct{}

func (namespaceCreatorImpl *NamespaceCreatorImpl) CreateNamespaces(cmdArgs []string) error {
	// there is no point in unsharing the  PID namespace without forking
	// unsharing the PID namespace does not move the calling process to the new PID namespace
	// but instead moves the next child process
	// to make full use of the PID namespace clone with NEWNS clone flag
	if len(cmdArgs) > 1 {
		return namespaceCreatorImpl.FinalizeNamespaces()
	}
	fmt.Println("Creating namespaces")
	cmd := exec.Command("/proc/self/exe", "afterNamespaced")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
	}
	err := cmd.Run()
	return err
}

func (namespaceCreatorImpl *NamespaceCreatorImpl) FinalizeNamespaces() error {
	fmt.Println("Finalizing namespaces.............")
	fmt.Println("Setting hostname..................")
	err := syscall.Sethostname([]byte("container"))
	if err != nil {
		return err
	}
	fmt.Println("Chrooting.........................")
	err = syscall.Chroot("/root/container")
	if err != nil {
		return err
	}
	fmt.Println("Changing working directory........")
	err = syscall.Chdir("/")
	if err != nil {
		return err
	}
	fmt.Println("Creating /proc dir if not exist...")
	if _, err := os.Stat("/proc"); os.IsNotExist(err) {
		os.Mkdir("/proc", os.ModeDir)
	}
	fmt.Println("Mounting proc.....................")
	err = syscall.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		return err
	}
	hostname, err := os.Hostname()
	fmt.Printf("PID: %d Hostname: %s\n", os.Getpid(), hostname)
	fmt.Println("Cleanup mounts....................")
	cmd := exec.Command("lsns")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	err = syscall.Unmount("/proc", 0)
	if err != nil {
		return err
	}
	return nil
}
