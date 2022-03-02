package main

import (
	"github.com/davidcosc/davidocker/containercreator"
	"os"
)

func main() {
	containerCreator := &containercreator.ContainerCreatorImpl{}
	switch os.Args[1] {
	case "run":
		err := containerCreator.CreateContainerNamespaces(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "containerNamespacesCreated":
		err := containerCreator.FinalizeContainer(os.Args[2:])
		if err != nil {
			panic(err)
		}
	}
}
