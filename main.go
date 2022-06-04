/*
Package davidocker contains functionality required for creating a single binary basic container.
It manages call order and priority based on commandline arguments passed.
It sets up an isolated, containerized process and changes the command executed inside that process to the specified binary.
The process will be isolated in its own pid, mount, uts and network namespace.
Setup of all necessary mounts and file systems required for these namespaces to work correctly is included.
The namespace setup is split into separate files for each namespace.
*/
package main

import "os"

// main calls required functions to setup a container.
// The run case defines the initial entry point to container setup.
// The remaining cases are intended for configurations that require
// a namespaced process.
func main() {
	switch os.Args[1] {
	case "run":
		err := CreateNetworkNamespace()
		if err != nil {
			panic(err)
		}
		err = CreateVethInterface(HOST_VETH, CONTAINER_VETH)
		if err != nil {
			panic(err)
		}
		err = MoveContainerVethToNetworkNamespace(CONTAINER_VETH, CONTAINER_NET_FILE)
		if err != nil {
			panic(err)
		}
		err = CreateContainerNamespaces(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "networkNamespaceCreated":
		err := FinalizeNetworkNamespace(CONTAINER_NET_FILE)
		if err != nil {
			panic(err)
		}
	case "containerNamespacesCreated":
		err := FinalizeContainer(os.Args[2:])
		if err != nil {
			panic(err)
		}
	}
}
