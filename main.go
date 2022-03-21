package main

import (
	"github.com/davidcosc/davidocker/container"
	"os"
)

func main() {
	switch os.Args[1] {
	case "run":
		err := container.CreateVethInterface(container.CONTAINER_ID)
		if err != nil {
			panic(err)
		}
		err = container.CreateNetworkNamespace()
		if err != nil {
			panic(err)
		}
		err = container.CreateContainerNamespaces(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "networkNamespaceCreated":
		err := container.FinalizeNetworkNamespace(container.CONTAINER_DIR)
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
