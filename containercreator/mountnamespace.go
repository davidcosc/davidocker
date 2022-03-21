package containercreator

import (
	"fmt"
	"os"
	"path"
	"syscall"
)

// createMounts sets up mandatory container mounts and required directories.
// createMounts should only be called inside a new mountnamespace.
// A new mountnamespace is initialized with a copy of all the mount points of its parent.
// This also includes all flags of those mount points e.g. their propagation type.
// We need to take steps to cleanup and reconfigure these mount points according to our needs.
// Our goal is to set the mount points in such a way, that they are cleaned up automatically
// once the mountnamespace is destroyed. Also mount points of the container mountnamespace
// should not have any effect on the parent mountnamespace.
// Mount points, that are not bind mounted and do not propagate to the parent are destroyed
// once their respective mountnamespace is destroyed.
// Mounts flagged MS_PRIVATE prevent propagation to the host.
// To ensure all mounts in the new mount namespace are flagged MS_PRIVATE we first
// need to recursively override the propagation flags of all mount points that were copied
// from the parent. This is done by only supplying the mount target / and the flags MS_REC and
// MS_PRIVATE to the mount syscall. All further mounts will be flagged MS_PRIVATE by default.
// Mounting the proc filesystem afterwards has the desired effect of correctly displaying
// PID 1 for the namespaced root process.
var createMounts = func(dir string) error {
	fmt.Println("* Setting hostname............................")
	fmt.Println("* Override / mount with MS_REC / MS_PRIVATE...")
	err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, "")
	if err != nil {
		return err
	}
	fmt.Println("* Creating /proc dir if not exist.............")
	if _, err := os.Stat(path.Join(dir, "proc")); os.IsNotExist(err) {
		os.Mkdir(path.Join(dir, "proc"), os.ModeDir)
	}
	fmt.Println("* Mounting proc...............................")
	return syscall.Mount("proc", path.Join(CONTAINER_DIR, "proc"), "proc", 0, "")
}

// changeRootFS chroots into the specified directory.
// After chrooting the cwd still points to the old directory tree.
// To fix that we change the the cwd to the new root dir.
var changeRootFS = func(dir string) error {
	fmt.Println("* Chrooting...................................")
	err := syscall.Chroot(dir)
	if err != nil {
		return err
	}
	fmt.Println("* Changing working directory..................")
	return syscall.Chdir("/")
}
