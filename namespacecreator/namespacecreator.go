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
	// unsharing the PID namespace does not move the calling process to the new PID namespace
	// but instead moves the next child process
	// used with clone however, will move the cloned Process immediately
	if len(cmdArgs) > 1 {
		return namespaceCreatorImpl.FinalizeNamespaces()
	}
	fmt.Println("Creating namespaces")
	cmd := exec.Command("/proc/self/exe", "afterNamespaced")
	cmd = prepareStdioDescriptors(cmd, os.Stdin, os.Stdout, os.Stderr)
	cmd = prepareNamespaces(cmd)
	err := cmd.Start()
	return err
}

func prepareStdioDescriptors(cmd *exec.Cmd, stdin, stdout, stderr *os.File) *exec.Cmd {
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd
}

func prepareNamespaces(cmd *exec.Cmd) *exec.Cmd {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
	}
	return cmd
}

func (namespaceCreatorImpl *NamespaceCreatorImpl) FinalizeNamespaces() error {
	fmt.Println("Finalizing namespaces.............")
	fmt.Println("Setting hostname..................")
	err := syscall.Sethostname([]byte("container"))
	if err != nil {
		return err
	}
	err = changeRootFS()
	if err != nil {
		return err
	}
	createMounts()
	if err != nil {
		return err
	}
	hostname, err := os.Hostname()
	fmt.Printf("PID: %d Hostname: %s\n", os.Getpid(), hostname)
	//fmt.Println("Cleanup mounts....................")
	//err = syscall.Unmount("/proc", 0)
	return syscall.Exec("/bin/sleep", []string{"/bin/sleep", "100s"}, []string{})
}

func changeRootFS() error {
	fmt.Println("Chrooting.........................")
	err := syscall.Chroot("/root/container")
	if err != nil {
		return err
	}
	fmt.Println("Changing working directory........")
	return syscall.Chdir("/")
}

func createMounts() error {
	fmt.Println("Creating /proc dir if not exist...")
	if _, err := os.Stat("/proc"); os.IsNotExist(err) {
		os.Mkdir("/proc", os.ModeDir)
	}
	fmt.Println("Mounting proc.....................")
	return syscall.Mount("proc", "/proc", "proc", 0, "")
}
