package fuse

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
)

// errDirectMountNotPermitted means the caller may not mount(2) a fuse file
// system itself, or asked for an option only the helper implements, and
// should fall back to fusermount.
var errDirectMountNotPermitted = errors.New("direct fuse mount not permitted")

// directMountArgs translates conf into the mount(2) arguments the kernel
// expects, the way fusermount does. Options the kernel does not know and
// fusermount implements itself (allow_root) are not translated.
func directMountArgs(conf *mountConfig) (source, fstype string, flags uintptr, data []string, ok bool) {
	source, fstype = "fuse", "fuse"
	flags = syscall.MS_NOSUID | syscall.MS_NODEV
	for k, v := range conf.options {
		switch k {
		case "fsname":
			source = v
		case "subtype":
			fstype = "fuse." + v
		case "ro":
			flags |= syscall.MS_RDONLY
		case "dev":
			flags &^= syscall.MS_NODEV
		case "suid":
			flags &^= syscall.MS_NOSUID
		case "allow_other", "default_permissions":
			data = append(data, k)
		case "nonempty":
			// Enforced by mount before selecting the direct or helper path.
		default:
			return "", "", 0, nil, false
		}
	}
	return source, fstype, flags, data, true
}

func mountDirect(dir string, conf *mountConfig) (*os.File, error) {
	source, fstype, flags, opts, ok := directMountArgs(conf)
	if !ok {
		return nil, errDirectMountNotPermitted
	}

	var st syscall.Stat_t
	if err := syscall.Stat(dir, &st); err != nil {
		if os.IsNotExist(err) {
			return nil, &MountpointDoesNotExistError{Path: dir}
		}
		return nil, fmt.Errorf("stat mountpoint %s: %v", dir, err)
	}

	f, err := os.OpenFile("/dev/fuse", os.O_RDWR, 0)
	if err != nil {
		if os.IsPermission(err) {
			return nil, errDirectMountNotPermitted
		}
		return nil, fmt.Errorf("open /dev/fuse: %v", err)
	}

	data := append([]string{
		fmt.Sprintf("fd=%d", f.Fd()),
		fmt.Sprintf("rootmode=%o", st.Mode&syscall.S_IFMT),
		fmt.Sprintf("user_id=%d", os.Getuid()),
		fmt.Sprintf("group_id=%d", os.Getgid()),
	}, opts...)
	if err := syscall.Mount(source, dir, fstype, flags, strings.Join(data, ",")); err != nil {
		f.Close()
		if err == syscall.EPERM || err == syscall.EACCES {
			return nil, errDirectMountNotPermitted
		}
		return nil, fmt.Errorf("mount %s: %v", dir, err)
	}
	return f, nil
}
