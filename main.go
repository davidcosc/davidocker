package main

import (
	"github.com/davidcosc/davidocker/containercreator"
	"os"
)

func main() {
	switch os.Args[1] {
	case "run":
		err := containercreator.CreateContainerNamespaces(os.Args[2:])
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
