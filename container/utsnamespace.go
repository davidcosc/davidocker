package container

import (
	"fmt"
	"syscall"
)

func setHostname(hostname string) error {
	fmt.Println("* Setting hostname............................")
	return syscall.Sethostname([]byte(hostname))
}
