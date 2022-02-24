package main

import (
	"fmt"
	"github.com/davidcosc/davidocker/namespacecreator"
)

func main() {
	namespaceCreator := &namespacecreator.NamespaceCreatorImpl{}
	pid, hostname, err := namespaceCreator.CreateNamespaces()
	if err != nil {
		fmt.Printf("PID: %d\n", pid)
		fmt.Println("Hostname: " + hostname)
	}
}
