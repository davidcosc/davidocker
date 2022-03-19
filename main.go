package main

import (
	"github.com/davidcosc/davidocker/containercreator"
	"os"
)

func main() {
	switch os.Args[1] {
	case "run":
		//err := containercreator.CreateNetworkNamespace()
		//if err != nil {
		//	panic(err)
		//}
		err := containercreator.CreateContainerNamespaces(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "networkNamespaceCreated":
		err := containercreator.FinalizeNetworkNamespace()
		if err != nil {
			panic(err)
		}
	case "containerNamespacesCreated":
		err := containercreator.FinalizeContainer(os.Args[2:])
		if err != nil {
			panic(err)
		}
	}
}
