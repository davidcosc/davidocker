package main

import (
	"github.com/davidcosc/davidocker/namespacecreator"
	"os"
)

func main() {
	namespaceCreator := &namespacecreator.NamespaceCreatorImpl{}
	err := namespaceCreator.CreateNamespaces(os.Args)
	if err != nil {
		panic(err)
	}
}
