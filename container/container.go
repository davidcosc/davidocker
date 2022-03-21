package container

import (
	"fmt"
	"os/exec"
	"syscall"
)

const CONTAINER_ID = "container"
const CONTAINER_DIR = "/root/" + CONTAINER_ID

// CreateContainerNamespaces runs a new process with unshared namespaces and redirected stdio.
// The PID and user namespaces require a new process to take effect.
// We can not unshare them for the current process.
var CreateContainerNamespaces = func(cmdArgs []string) error {
	cmd := exec.Command("/proc/self/exe", append([]string{"containerNamespacesCreated"}, cmdArgs...)...)
	fmt.Println("Creating container namespaces.................")
	var err error
	cmd.Stdin, cmd.Stdout, cmd.Stderr, err = prepareCustomStdioDescriptors(CONTAINER_DIR)
	if err != nil {
		return err
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS,
	}
	fmt.Println("* Restarting self in namespaces...............")
	err = cmd.Start()
	return err
}

// FinalizeContainer sets up hostname, rootfs and mounts inside the new process.
// It changes the command run by the process to the actual command to be containerized.
// This is done using the exec syscall. All open fds that are marked close-on-exec are closed.
// Since this is not true for the stdio descriptors, they will be kept open.
var FinalizeContainer = func(cmdArgs []string) error {
	fmt.Println("Finalizing container..........................")
	err := joinNetworkNamespace(CONTAINER_DIR)
	if err != nil {
		return err
	}
	err = createMounts(CONTAINER_DIR)
	if err != nil {
		return err
	}
	err = changeRootFS(CONTAINER_DIR)
	if err != nil {
		return err
	}
	err = setHostname(CONTAINER_ID)
	if err != nil {
		return err
	}
	return syscall.Exec(cmdArgs[0], cmdArgs, []string{})
}
