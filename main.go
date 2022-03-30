/*
Package main contains functionality to call all functions required for
creating a single binary basic container. It manages call order an priority
based on commandline arguments passed.
*/
package main

import (
	"github.com/davidcosc/davidocker/container"
	"os"
)

// main calls required functions to setup a container.
// The run case defines the initial entry point to container setup.
// The remaining cases are intended for configurations that require
// a namespaced process.
func main() {
	switch os.Args[1] {
	case "run":
		err := container.CreateNetworkNamespace()
		if err != nil {
			panic(err)
		}
		err = container.CreateVethInterface(container.HOST_VETH, container.CONTAINER_VETH)
		if err != nil {
			panic(err)
		}
		err = container.MoveContainerVethToNetworkNamespace(container.CONTAINER_VETH, container.CONTAINER_NET_FILE)
		if err != nil {
			panic(err)
		}
		err = container.CreateContainerNamespaces(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "networkNamespaceCreated":
		err := container.FinalizeNetworkNamespace(container.CONTAINER_NET_FILE)
		if err != nil {
			panic(err)
		}
	case "containerNamespacesCreated":
		err := container.FinalizeContainer(os.Args[2:])
		if err != nil {
			panic(err)
		}
	}
}
