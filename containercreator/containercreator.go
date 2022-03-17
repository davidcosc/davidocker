package containercreator

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// CreateContainerNamespaces runs a new process with unshared namespaces and redirected stdio.
// The PID and user namespaces require a new process to take effect.
// We can not unshare them for the current process.
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

// PrepareStdioDescriptors sets up stdio file descriptors of the new process.
// Stdio of the containerized process will be redirected to files inside the containers rootfs.
var prepareStdioDescriptors = func(cmd *exec.Cmd) (*exec.Cmd, error) {
	var err error
	cmd.Stdin, err = os.Create("/root/container/stdin")
	cmd.Stdout, err = os.Create("/root/container/stdout")
	cmd.Stderr, err = os.Create("/root/container/stderr")
	return cmd, err
}

// Prepare namespaces sets up namespaces for the new process.
var prepareNamespaces = func(cmd *exec.Cmd) *exec.Cmd {
	fmt.Println("* Preparing namespaces..................................................................")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS,
	}
	return cmd
}

// FinalizeContainer sets up hostname, rootfs and mounts inside the new process.
// It changes the command run by the process to the actual command to be containerized.
// This is done using the exec syscall. All open fds that are marked close-on-exec are closed.
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
	// Use to debug:
	// content, err := os.ReadFile("/proc/self/mountinfo")
	// fmt.Printf("Mount: %s\n", content)
	// hostname, err := os.Hostname()
	// fmt.Printf("Hostname: %s\n", hostname)
	return syscall.Exec(cmdArgs[0], cmdArgs, []string{})
}

// CreateMounts sets up mandatory container mounts and required directories.
// CreateMounts should only be called inside a new mountnamespace.
// A new mountnamespace is initialized with a copy of all the mount points of its parent.
// This also includes all flags of those mount points e.g. their propagation type.
// We need to take steps to cleanup and reconfigure these mount points according to our needs.
// Our goal is to set the mount points in such a way, that they are cleaned up automatically
// once the mountnamespace is destroyed. Also mount points of the container mountnamespace
// should not have any effect on the parent mountnamespace.
// Mount points, that are not bind mounted and do not propagate to the parent are destroyed
// once their respective mountnamespace is destroyed.
// Mounts flagged MS_PRIVATE prevent propagation to the host.
// To ensure all mounts in the new mount namespace are flagged MS_PRIVATE we first
// need to recursively override the propagation flags of all mount points that were copied
// from the parent. This is done by only supplying the mount target / and the flags MS_REC and
// MS_PRIVATE to the mount syscall. All further mounts will be flagged MS_PRIVATE by default.
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

// ChangeRootFS chroots into the container directory.
// After chrooting the cwd still points to the old directory tree.
// To fix that we change the the cwd to the new root dir.
var changeRootFS = func() error {
	fmt.Println("* Chrooting.............................................................................")
	err := syscall.Chroot("/root/container")
	if err != nil {
		return err
	}
	fmt.Println("* Changing working directory............................................................")
	return syscall.Chdir("/")
}
