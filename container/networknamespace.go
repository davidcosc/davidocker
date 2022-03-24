package container

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
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
var FinalizeNetworkNamespace = func(dir string) error {
	fmt.Println("FinalizeNetworkNamespace......................")
	netFileName := path.Join(dir, "net")
	netFD, err := os.Create(netFileName)
	defer netFD.Close()
	if err != nil {
		return err
	}
	return syscall.Mount("/proc/self/ns/net", netFileName, "bind", syscall.MS_BIND|syscall.MS_SHARED, "")
}

// CreateVethInterface sets up a veth tunnel interface inside the host network
// namespace. It is intended to be used for linking the container network to the
// host later on. It only sets the ip for the host side of the veth.
// The ip is statically hard coded, since this implementation is focused on
// exploring container basics. A full network setup including dhcp, container bridge
// etc. will not be provided. This results in only one container being able to run
// at a time.
var CreateVethInterface = func(name string) error {
	fmt.Println("Creating veth interface.......................")
	veth0 := "veth0_" + name
	veth1 := "veth1_" + name
	vethExists, err := checkLinkExists(veth0)
	if err != nil || vethExists {
		fmt.Println("* Veth already exists.........................")
		return err
	}
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = veth0
	veth0Struct := &netlink.Veth{
		LinkAttrs: linkAttrs,
		PeerName:  veth1,
	}
	fmt.Println("* Adding veth link............................")
	err = netlink.LinkAdd(veth0Struct)
	if err != nil {
		return err
	}
	err = netlink.LinkSetUp(veth0Struct)
	if err != nil {
		return err
	}
	ip, netMask, err := net.ParseCIDR("10.0.0.1/24")
	if err != nil {
		return err
	}
	ipNet := &net.IPNet{IP: ip, Mask: netMask.Mask}
	fmt.Println("* Adding ip address...........................")
	addr := &netlink.Addr{IPNet: ipNet, Label: ""}
	return netlink.AddrAdd(veth0Struct, addr)
}

// checkLinkExists iterates through the list of existing interfaces.
// It returns true if the interface with the specified name already exists.
var checkLinkExists = func(name string) (bool, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return false, err
	}
	for _, link := range links {
		if link.Attrs().Name == name {
			return true, err
		}
	}
	return false, err
}

// MoveVeth1ToNetworkNamespace moves one part of the veth interface pair to the
// bind mounted network namespace.
var MoveVeth1ToNetworkNamespace = func(name, dir string) error {
	fmt.Println("*Moving Veth to network namespace.............")
	veth1 := "veth1_" + name
	fmt.Println("veth1: " + veth1)
	veth1Link, err := netlink.LinkByName(veth1)
	if err != nil {
		return err
	}
	netFD, err := syscall.Open(path.Join(dir, "net"), syscall.O_RDONLY, 0666)
	defer syscall.Close(netFD)
	if err != nil {
		return err
	}
	err = netlink.LinkSetNsFd(veth1Link, netFD)
	if err != nil {
		return err
	}
	return nil
}

// joinNetworkNamespace is intended to make the containerized process join the previously
// set up network namespace. It also attempts to cleanup the now obsolete network bind mount.
// After cleanup the lifetime of the network namespace is bound to the lifetime of the
// containerized process.
var joinNetworkNamespace = func(dir string) error {
	fmt.Println("* Opening network namespace mount.............")
	netFD, err := syscall.Open(path.Join(dir, "net"), syscall.O_RDONLY, 0644)
	if err != nil {
		err = syscall.Unmount(path.Join(dir, "net"), 0)
		return err
	}
	fmt.Println("* Joining network namespace...................")
	// 308 is trap code for setns syscall
	_, _, errNo := syscall.RawSyscall(308, uintptr(netFD), 0, 0)
	if errNo != 0 {
		err = syscall.Close(netFD)
		err = syscall.Unmount(path.Join(dir, "net"), 0)
		return errNo
	}
	err = syscall.Close(netFD)
	err = syscall.Unmount(path.Join(dir, "net"), 0)
	fmt.Println("* Removing network namespace bind mount.......")
	err = os.Remove(path.Join(dir, "net"))
	return err
}
