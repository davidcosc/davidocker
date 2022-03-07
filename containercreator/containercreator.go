package containercreator

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type ContainerCreator interface {
	CreateContainerNamespaces(cmdArgs []string) error
	FinalizeContainer() error
}

type ContainerCreatorImpl struct{}

// Sets up and runs a new process with unshared namespaces and redirected stdio.
// Simply unsharing the namespaces for the current process is not enough.
// The PID and user namespaces require a new process to take effect.
func (containerCreatorImpl *ContainerCreatorImpl) CreateContainerNamespaces(cmdArgs []string) error {
	cmd := exec.Command("/proc/self/exe", append([]string{"containerNamespacesCreated"}, cmdArgs...)...)
	fmt.Println("Creating container namespaces.....")
	fmt.Println("* Preparing stdio descriptors.....")
	var err error
	cmd, err = prepareStdioDescriptors(cmd)
	if err != nil {
		return err
	}
	cmd = prepareNamespaces(cmd)
	fmt.Println("* Restarting self in namespaces...")
	err = cmd.Start()
	return err
}

// Sets up file descriptors for redirecting stdio of the new process.
// Stdio will be redirected to respective files inside the containers rootfs.
func prepareStdioDescriptors(cmd *exec.Cmd) (*exec.Cmd, error) {
	var err error
	cmd.Stdin, err = os.Create("/root/container/stdin")
	cmd.Stdout, err = os.Create("/root/container/stdout")
	cmd.Stderr, err = os.Create("/root/container/stderr")
	return cmd, err
}

// Sets up namespaces for the new process.
func prepareNamespaces(cmd *exec.Cmd) *exec.Cmd {
	// Sets up the namespaces to be created for the new process.
	fmt.Println("* Preparing namespaces............")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
	}
	return cmd
}

// Sets up hostname, rootfs and mounts inside the new process.
// Also changes the command being run by the process to the actual command to be containerized.
// Changing the command by exec closes all open fds that are marked close-on-exec.
// Since this is not true for the stdio descriptors, they will be kept open.
func (containerCreatorImpl *ContainerCreatorImpl) FinalizeContainer(cmdArgs []string) error {
	fmt.Println("Finalizing container..............")
	fmt.Println("* Setting hostname................")
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
	return syscall.Exec(cmdArgs[0], cmdArgs, []string{})
}

// Chroots and cwds into the container directory.
// After chrooting the cwd still points to the old directory tree.
// To fix that we change the cwd.
func changeRootFS() error {
	fmt.Println("* Chrooting.......................")
	err := syscall.Chroot("/root/container")
	if err != nil {
		return err
	}
	fmt.Println("* Changing working directory......")
	return syscall.Chdir("/")
}

// Setup mandatory container mounts and required directories.
// Intended for mount setup inside the new mount namespace.
// Mounts are flagged MS_PRIVATE to prevent propagation to the host.
// Destroying the mountnamespace removes all mountnamespace specific private mounts.
// We therefore do not have to do any umount cleanup.
func createMounts() error {
	fmt.Println("* Creating /proc dir if not exist.")
	if _, err := os.Stat("/proc"); os.IsNotExist(err) {
		os.Mkdir("/proc", os.ModeDir)
	}
	fmt.Println("* Mounting proc...................")
	return syscall.Mount("proc", "/proc", "proc", syscall.MS_PRIVATE, "")
}
