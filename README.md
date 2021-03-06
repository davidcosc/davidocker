<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# davidocker

```go
import "github.com/davidcosc/davidocker"
```

Package davidocker contains functionality required for creating a single binary basic container\. It manages call order and priority based on commandline arguments passed\. It sets up an isolated\, containerized process and changes the command executed inside that process to the specified binary\. The process will be isolated in its own pid\, mount\, uts and network namespace\. Setup of all necessary mounts and file systems required for these namespaces to work correctly is included\. The namespace setup is split into separate files for each namespace\.

## Index

- [Constants](<#constants>)
- [func CreateContainerNamespaces(cmdArgs []string) error](<#func-createcontainernamespaces>)
- [func CreateNetworkNamespace() error](<#func-createnetworknamespace>)
- [func CreateVethInterface(hostVeth, containerVeth string) error](<#func-createvethinterface>)
- [func FinalizeContainer(cmdArgs []string) error](<#func-finalizecontainer>)
- [func FinalizeNetworkNamespace(containerNetFile string) error](<#func-finalizenetworknamespace>)
- [func MoveContainerVethToNetworkNamespace(containerVeth, containerNetFile string) error](<#func-movecontainervethtonetworknamespace>)
- [func changeRootFS(dir string) error](<#func-changerootfs>)
- [func configureLink(ipCIDR string, link netlink.Link) error](<#func-configurelink>)
- [func createMounts(dir string) error](<#func-createmounts>)
- [func joinNetworkNamespace(containerNetFile, containerVeth string) error](<#func-joinnetworknamespace>)
- [func main()](<#func-main>)
- [func prepareCustomStdioDescriptors(dir string) (*os.File, *os.File, *os.File, error)](<#func-preparecustomstdiodescriptors>)
- [func setHostname(hostname string) error](<#func-sethostname>)


## Constants

```go
const CONTAINER_DIR = "/root/" + CONTAINER_ID
```

```go
const CONTAINER_ID = "container"
```

```go
const CONTAINER_NET_FILE = CONTAINER_DIR + "/net"
```

```go
const CONTAINER_VETH = "veth1_" + CONTAINER_ID
```

```go
const HOST_VETH = "veth0_" + CONTAINER_ID
```

## func CreateContainerNamespaces

```go
func CreateContainerNamespaces(cmdArgs []string) error
```

CreateContainerNamespaces runs a new process with unshared namespaces and redirected stdio\. The PID and user namespaces require a new process to take effect\. We can not unshare them for the current process\.

## func CreateNetworkNamespace

```go
func CreateNetworkNamespace() error
```

CreateNetworkNamespace runs a new process with unshared network namespace\. After the spawned networkNamespaceCreated process finishes\, program flow contiunues in this parent process\.

## func CreateVethInterface

```go
func CreateVethInterface(hostVeth, containerVeth string) error
```

CreateVethInterface sets up a veth tunnel interface inside the host network namespace\. It is intended to be used for linking the container network to the host later on\. It sets the host link of the veth pair to up\.

## func FinalizeContainer

```go
func FinalizeContainer(cmdArgs []string) error
```

FinalizeContainer sets up hostname\, rootfs and mounts inside the new process\. It changes the command run by the process to the actual command to be containerized\. This is done using the exec syscall\. All open fds that are marked close\-on\-exec are closed\. Since this is not true for the stdio descriptors\, they will be kept open\.

## func FinalizeNetworkNamespace

```go
func FinalizeNetworkNamespace(containerNetFile string) error
```

FinalizeNetworkNamespace bind mounts the network namespace to a file inside the container directory\. This way the network namespace will persist even after the process terminates\. This allows for veth interface tunnels to be setup from within the host network namespace\, linking the host and bind mounted network namespace\. The mount is flagged MS\_SHARED\, to allow the container to remove it globally after joining the network namespace later on\. Once this is done\, the lifetime of the network namespace and related interfaces is bound to the lifetime of the containerized process\.

## func MoveContainerVethToNetworkNamespace

```go
func MoveContainerVethToNetworkNamespace(containerVeth, containerNetFile string) error
```

MoveContainerVethToNetworkNamespace moves one part of the veth interface pair to the bind mounted network namespace\.

## func changeRootFS

```go
func changeRootFS(dir string) error
```

changeRootFS chroots into the specified directory\. After chrooting the cwd still points to the old directory tree\. To fix that we change the the cwd to the new root dir\.

## func configureLink

```go
func configureLink(ipCIDR string, link netlink.Link) error
```

configureLink sets the ip for the specified link\. The ip set is intended to be statically hard coded\. This implementation focuses on exploring container basics\. A full network setup including dhcp\, container bridge etc\. will not be provided\. This results in only one container being able to run at a time\.

## func createMounts

```go
func createMounts(dir string) error
```

createMounts sets up mandatory container mounts and required directories\. createMounts should only be called inside a new mountnamespace\. A new mountnamespace is initialized with a copy of all the mount points of its parent\. This also includes all flags of those mount points e\.g\. their propagation type\. We need to take steps to cleanup and reconfigure these mount points according to our needs\. Our goal is to set the mount points in such a way\, that they are cleaned up automatically once the mountnamespace is destroyed\. Also mount points of the container mountnamespace should not have any effect on the parent mountnamespace\. Mount points\, that are not bind mounted and do not propagate to the parent are destroyed once their respective mountnamespace is destroyed\. Mounts flagged MS\_PRIVATE prevent propagation to the host\. To ensure all mounts in the new mount namespace are flagged MS\_PRIVATE we first need to recursively override the propagation flags of all mount points that were copied from the parent\. This is done by only supplying the mount target / and the flags MS\_REC and MS\_PRIVATE to the mount syscall\. All further mounts will be flagged MS\_PRIVATE by default\. Mounting the proc filesystem afterwards has the desired effect of correctly displaying PID 1 for the namespaced root process\.

## func joinNetworkNamespace

```go
func joinNetworkNamespace(containerNetFile, containerVeth string) error
```

joinNetworkNamespace is intended to make the containerized process join the previously set up network namespace\. It also attempts to cleanup the now obsolete network bind mount\. After cleanup the lifetime of the network namespace is bound to the lifetime of the containerized process\. After joining is complete remaining network interface configurations are completed\.

## func main

```go
func main()
```

main calls required functions to setup a container\. The run case defines the initial entry point to container setup\. The remaining cases are intended for configurations that require a namespaced process\.

## func prepareCustomStdioDescriptors

```go
func prepareCustomStdioDescriptors(dir string) (*os.File, *os.File, *os.File, error)
```

prepareCustomStdioDescriptors sets up file descriptors inside the specified directory\. Stdio of the containerized process will be redirected to these later on\.

## func setHostname

```go
func setHostname(hostname string) error
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
