package fuse

import (
	"reflect"
	"sort"
	"syscall"
	"testing"
)

func TestDirectMountArgs(t *testing.T) {
	for _, tc := range []struct {
		name    string
		options []MountOption
		source  string
		fstype  string
		flags   uintptr
		data    []string
		ok      bool
	}{
		{
			name:   "defaults",
			source: "fuse", fstype: "fuse",
			flags: syscall.MS_NOSUID | syscall.MS_NODEV,
			ok:    true,
		},
		{
			name:    "vnodefs",
			options: []MountOption{FSName("vnodefs"), AllowOther(), DefaultPermissions()},
			source:  "vnodefs", fstype: "fuse",
			flags: syscall.MS_NOSUID | syscall.MS_NODEV,
			data:  []string{"allow_other", "default_permissions"},
			ok:    true,
		},
		{
			name:    "subtype, read-only, dev and suid",
			options: []MountOption{Subtype("foo"), ReadOnly(), AllowDev(), AllowSUID()},
			source:  "fuse", fstype: "fuse.foo",
			flags: syscall.MS_RDONLY,
			ok:    true,
		},
		{
			name:    "allow_root needs the helper",
			options: []MountOption{AllowRoot()},
		},
		{
			name:    "nonempty is userspace-only",
			options: []MountOption{AllowNonEmptyMount()},
			source:  "fuse", fstype: "fuse",
			flags: syscall.MS_NOSUID | syscall.MS_NODEV,
			ok:    true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conf := mountConfig{options: make(map[string]string)}
			for _, option := range tc.options {
				if err := option(&conf); err != nil {
					t.Fatal(err)
				}
			}
			source, fstype, flags, data, ok := directMountArgs(&conf)
			sort.Strings(data)
			if ok != tc.ok || source != tc.source || fstype != tc.fstype || flags != tc.flags || !reflect.DeepEqual(data, tc.data) {
				t.Errorf("got %q %q %#x %q %v, want %q %q %#x %q %v",
					source, fstype, flags, data, ok, tc.source, tc.fstype, tc.flags, tc.data, tc.ok)
			}
		})
	}
}
