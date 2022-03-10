package containercreator

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Sets up and runs a new process with unshared namespaces and redirected stdio.
// Simply unsharing the namespaces for the current process is not enough.
// The PID and user namespaces require a new process to take effect.
var CreateContainerNamespaces = func(cmdArgs []string) error {
	cmd := exec.Command("/proc/self/exe", append([]string{"containerNamespacesCreated"}, cmdArgs...)...)
	fmt.Println("Creating container namespaces...........................................................")
	fmt.Println("* Preparing stdio descriptors...........................................................")
	var err error
	cmd, err = prepareStdioDescriptors(cmd)
	if err != nil {
		return err
	}
	cmd = prepareNamespaces(cmd)
	fmt.Println("* Restarting self in namespaces.........................................................")
	err = cmd.Start()
	return err
}

// Sets up file descriptors for redirecting stdio of the new process.
// Stdio will be redirected to respective files inside the containers rootfs.
var prepareStdioDescriptors = func(cmd *exec.Cmd) (*exec.Cmd, error) {
	var err error
	cmd.Stdin, err = os.Create("/root/container/stdin")
	cmd.Stdout, err = os.Create("/root/container/stdout")
	cmd.Stderr, err = os.Create("/root/container/stderr")
	return cmd, err
}

// Sets up namespaces for the new process.
var prepareNamespaces = func(cmd *exec.Cmd) *exec.Cmd {
	fmt.Println("* Preparing namespaces..................................................................")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS,
	}
	return cmd
}

// Sets up hostname, rootfs and mounts inside the new process.
// Also changes the command being run by the process to the actual command to be containerized.
// Changing the command by exec closes all open fds that are marked close-on-exec.
// Since this is not true for the stdio descriptors, they will be kept open.
var FinalizeContainer = func(cmdArgs []string) error {
	fmt.Println("Finalizing container....................................................................")
	err := createMounts()
	if err != nil {
		return err
	}
	err = changeRootFS()
	if err != nil {
		return err
	}
	fmt.Println("* Setting hostname......................................................................")
	err = syscall.Sethostname([]byte("container"))
	if err != nil {
		return err
	}
	content, err := os.ReadFile("/proc/self/mountinfo")
	fmt.Printf("Mount: %s\n", content)
	hostname, err := os.Hostname()
	fmt.Printf("Hostname: %s\n", hostname)
	return syscall.Exec(cmdArgs[0], cmdArgs, []string{})
}

// Setup mandatory container mounts and required directories.
// Intended for mount setup inside the new mount namespace.
// Only mounts flagged MS_PRIVATE prevent propagation to the host.
// Destroying the mountnamespace removes all mountnamespace specific private mounts.
// To ensure all mounts in the new mount namespace are flagged MS_PRIVATE we first
// need to recursively override the flags of the per default copied mounts of the parent
// mount namespace with MS_PRIVATE. This is done by only supplying the mount target /
// as well as the flags MS_REC and MS_PRIVATE to the mount command.
// Mounting the proc filesystem afterwards has the desired effect of correctly displaying
// PID 1 for the namespaced root process.
var createMounts = func() error {
	fmt.Println("* Override / mount with MS_REC / MS_PRIVATE to ensure all further mounts are private....")
	err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, "")
	if err != nil {
		return err
	}
	fmt.Println("* Creating /proc dir if not exist.......................................................")
	if _, err := os.Stat("/root/container/proc"); os.IsNotExist(err) {
		os.Mkdir("/root/container/proc", os.ModeDir)
	}
	fmt.Println("* Mounting proc.........................................................................")
	return syscall.Mount("proc", "/root/container/proc", "proc", 0, "")
}

// Chroots and cwds into the container directory.
// After chrooting the cwd still points to the old directory tree.
// To fix that we change the cwd.
var changeRootFS = func() error {
	fmt.Println("* Chrooting.............................................................................")
	err := syscall.Chroot("/root/container")
	if err != nil {
		return err
	}
	fmt.Println("* Changing working directory............................................................")
	return syscall.Chdir("/")
}
