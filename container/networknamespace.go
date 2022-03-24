package container

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"

	"github.com/vishvananda/netlink"
)

// CreateNetworkNamespace runs a new process with unshared network namespace.
// After the spawned networkNamespaceCreated process finishes, program flow contiunues
// in this parent process.
var CreateNetworkNamespace = func() error {
	fmt.Println("Creating network namespace....................")
	fmt.Println("* Run new process in new network namespace....")
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
var FinalizeNetworkNamespace = func(containerNetFile string) error {
	fmt.Println("FinalizeNetworkNamespace......................")
	netFD, err := os.Create(containerNetFile)
	defer netFD.Close()
	if err != nil {
		return err
	}
	fmt.Println("* Bind mounting network namespace.............")
	return syscall.Mount("/proc/self/ns/net", containerNetFile, "bind", syscall.MS_BIND|syscall.MS_SHARED, "")
}

// CreateVethInterface sets up a veth tunnel interface inside the host network
// namespace. It is intended to be used for linking the container network to the
// host later on. It only sets the ip for the host side of the veth.
// The ip is statically hard coded, since this implementation is focused on
// exploring container basics. A full network setup including dhcp, container bridge
// etc. will not be provided. This results in only one container being able to run
// at a time.
var CreateVethInterface = func(hostVeth, containerVeth string) error {
	fmt.Println("Creating veth interface.......................")
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = hostVeth
	vethStruct := &netlink.Veth{
		LinkAttrs: linkAttrs,
		PeerName:  containerVeth,
	}
	fmt.Println("* Adding veth link............................")
	err := netlink.LinkAdd(vethStruct)
	if err != nil {
		return err
	}
	return configureLink("10.0.0.1/24", vethStruct)

}

// configureLink brings up the specified link and sets the specified ip address
var configureLink = func(ipCIDR string, link netlink.Link) error {
	fmt.Println("* Bringing up link............................")
	err := netlink.LinkSetUp(link)
	if err != nil {
		return err
	}
	ip, netMask, err := net.ParseCIDR(ipCIDR)
	if err != nil {
		return err
	}
	ipNet := &net.IPNet{IP: ip, Mask: netMask.Mask}
	addr := &netlink.Addr{IPNet: ipNet, Label: ""}
	fmt.Println("* Adding ip address...........................")
	return netlink.AddrAdd(link, addr)
}

// MoveContainerVethToNetworkNamespace moves one part of the veth interface pair to the
// bind mounted network namespace.
var MoveContainerVethToNetworkNamespace = func(containerVeth, containerNetFile string) error {
	fmt.Println("*Moving Veth to network namespace.............")
	containerVethLink, err := netlink.LinkByName(containerVeth)
	if err != nil {
		return err
	}
	netFD, err := syscall.Open(containerNetFile, syscall.O_RDONLY, 0644)
	defer syscall.Close(netFD)
	if err != nil {
		return err
	}
	err = netlink.LinkSetNsFd(containerVethLink, netFD)
	if err != nil {
		return err
	}
	return nil
}

// joinNetworkNamespace is intended to make the containerized process join the previously
// set up network namespace. It also attempts to cleanup the now obsolete network bind mount.
// After cleanup the lifetime of the network namespace is bound to the lifetime of the
// containerized process.
var joinNetworkNamespace = func(containerNetFile, containerVeth string) error {
	fmt.Println("* Opening network namespace mount.............")
	netFD, err := syscall.Open(containerNetFile, syscall.O_RDONLY, 0644)
	if err != nil {
		err = syscall.Unmount(containerNetFile, 0)
		return err
	}
	fmt.Println("* Joining network namespace...................")
	// 308 is trap code for setns syscall
	_, _, errNo := syscall.RawSyscall(308, uintptr(netFD), 0, 0)
	if errNo != 0 {
		err = syscall.Close(netFD)
		err = syscall.Unmount(containerNetFile, 0)
		return errNo
	}
	err = syscall.Close(netFD)
	err = syscall.Unmount(containerNetFile, 0)
	fmt.Println("* Removing network namespace bind mount.......")
	err = os.Remove(containerNetFile)
	if err != nil {
		return err
	}
	fmt.Println("* Bringing up lo..............................")
	loLink, err := netlink.LinkByName("lo")
	if err != nil || loLink == nil {
		return err
	}
	err = netlink.LinkSetUp(loLink)
	if err != nil {
		return err
	}
	return err
	//fmt.Println("* Getting container veth link.................")
	//containerVethLink, err := netlink.LinkByName("veth1_container")
	//if err != nil || containerVethLink == nil {
	//	return err
	//}
	//return configureLink("10.0.0.2/24", containerVethLink)
}