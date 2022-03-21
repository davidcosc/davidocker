package containercreator

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
)

const CONTAINER_DIR = "/root/container"

// CreateNetworkNamespace runs a new process with unshared network namespace.
// After the spawned networkNamespaceCreated process finishes, program flow contiunues
// in this parent process.
var CreateNetworkNamespace = func() error {
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
var FinalizeNetworkNamespace = func() error {
	fmt.Println("FinalizeNetworkNamespace................................................................")
	netFD, err := os.Create(path.Join(CONTAINER_DIR, "net"))
	defer netFD.Close()
	if err != nil {
		return err
	}
	return syscall.Mount("/proc/self/ns/net", path.Join(CONTAINER_DIR, "net"), "bind", syscall.MS_BIND|syscall.MS_SHARED, "")
}

// CreateContainerNamespaces runs a new process with unshared namespaces and redirected stdio.
// The PID and user namespaces require a new process to take effect.
// We can not unshare them for the current process.
var CreateContainerNamespaces = func(cmdArgs []string) error {
	cmd := exec.Command("/proc/self/exe", append([]string{"containerNamespacesCreated"}, cmdArgs...)...)
	fmt.Println("Creating container namespaces...........................................................")
	var err error
	cmd.Stdin, cmd.Stdout, cmd.Stderr, err = prepareContainerStdioDescriptors()
	if err != nil {
		return err
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS,
	}
	fmt.Println("* Restarting self in namespaces.........................................................")
	err = cmd.Start()
	return err
}

// PrepareContainerStdioDescriptors sets up file descriptors inside the container rootfs.
// Stdio of the containerized process will be redirected to these later on.
var prepareContainerStdioDescriptors = func() (*os.File, *os.File, *os.File, error) {
	fmt.Println("* Preparing stdio descriptors...........................................................")
	stdin, err := os.Create(path.Join(CONTAINER_DIR, "stdin"))
	if err != nil {
		return nil, nil, nil, err
	}
	stdout, err := os.Create(path.Join(CONTAINER_DIR, "stdout"))
	if err != nil {
		return nil, nil, nil, err
	}
	stderr, err := os.Create(path.Join(CONTAINER_DIR, "stderr"))
	if err != nil {
		return nil, nil, nil, err
	}
	return stdin, stdout, stderr, err
}

// FinalizeContainer sets up hostname, rootfs and mounts inside the new process.
// It changes the command run by the process to the actual command to be containerized.
// This is done using the exec syscall. All open fds that are marked close-on-exec are closed.
// Since this is not true for the stdio descriptors, they will be kept open.
var FinalizeContainer = func(cmdArgs []string) error {
	fmt.Println("Finalizing container....................................................................")
	err := joinNetworkNamespace()
	if err != nil {
		return err
	}
	err = createMounts()
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

// joinNetworkNamespace is intended to make the containerized process join the previously
// set up network namespace. It also attempts to cleanup the now obsolete network bind mount.
// After cleanup the lifetime of the network namespace is bound to the lifetime of the
// containerized process.
var joinNetworkNamespace = func() error {
	netFD, err := syscall.Open(path.Join(CONTAINER_DIR, "net"), syscall.O_RDONLY, 0644)
	if err != nil {
		err = syscall.Unmount(path.Join(CONTAINER_DIR, "net"), 0)
		return err
	}
	// 308 is trap code for setns
	_, _, errNo := syscall.RawSyscall(308, uintptr(netFD), 0, 0)
	if errNo != 0 {
		err = syscall.Close(netFD)
		err = syscall.Unmount(path.Join(CONTAINER_DIR, "net"), 0)
		return errNo
	}
	err = syscall.Close(netFD)
	err = syscall.Unmount(path.Join(CONTAINER_DIR, "net"), 0)
	err = os.Remove(path.Join(CONTAINER_DIR, "net"))
	return err
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
	if _, err := os.Stat(path.Join(CONTAINER_DIR, "proc")); os.IsNotExist(err) {
		os.Mkdir(path.Join(CONTAINER_DIR, "proc"), os.ModeDir)
	}
	fmt.Println("* Mounting proc.........................................................................")
	return syscall.Mount("proc", path.Join(CONTAINER_DIR, "proc"), "proc", 0, "")
}

// ChangeRootFS chroots into the container directory.
// After chrooting the cwd still points to the old directory tree.
// To fix that we change the the cwd to the new root dir.
var changeRootFS = func() error {
	fmt.Println("* Chrooting.............................................................................")
	err := syscall.Chroot(CONTAINER_DIR)
	if err != nil {
		return err
	}
	fmt.Println("* Changing working directory............................................................")
	return syscall.Chdir("/")
}
