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

func (containerCreatorImpl *ContainerCreatorImpl) CreateContainerNamespaces(cmdArgs []string) error {
	// unsharing the PID namespace does not move the calling process to the new PID namespace
	// but instead moves the next child process
	// used with clone however, will move the cloned Process immediately
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

func prepareStdioDescriptors(cmd *exec.Cmd) (*exec.Cmd, error) {
	var err error
	cmd.Stdin, err = os.Create("/root/container/stdin")
	cmd.Stdout, err = os.Create("/root/container/stdout")
	cmd.Stderr, err = os.Create("/root/container/stderr")
	return cmd, err
}

func prepareNamespaces(cmd *exec.Cmd) *exec.Cmd {
	fmt.Println("* Preparing namespaces............")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
	}
	return cmd
}

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

func changeRootFS() error {
	fmt.Println("* Chrooting.......................")
	err := syscall.Chroot("/root/container")
	if err != nil {
		return err
	}
	fmt.Println("* Changing working directory......")
	return syscall.Chdir("/")
}

func createMounts() error {
	//mounting with MS_PRIVATE Flag prevents the mount from propagating to the host
	//we therefore do not have to do any umount clean up
	//destroying the mountnamespace removes all mountnamespace specific private mounts
	fmt.Println("* Creating /proc dir if not exist.")
	if _, err := os.Stat("/proc"); os.IsNotExist(err) {
		os.Mkdir("/proc", os.ModeDir)
	}
	fmt.Println("* Mounting proc...................")
	return syscall.Mount("proc", "/proc", "proc", syscall.MS_PRIVATE, "")
}
