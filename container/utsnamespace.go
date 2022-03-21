package container

import (
	"fmt"
	"syscall"
)

var setHostname = func(hostname string) error {
	fmt.Println("* Setting hostname............................")
	return syscall.Sethostname([]byte(hostname))
}
