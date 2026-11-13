package fuse

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tempMountpoint(t *testing.T) string {
	t.Helper()
	dir, err := ioutil.TempDir("", "fuse-mountpoint-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestMountNonEmptyPolicyAppliesBeforeMethodSelection(t *testing.T) {
	dir := tempMountpoint(t)
	defer os.RemoveAll(dir)
	if err := ioutil.WriteFile(filepath.Join(dir, "existing"), []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}

	oldPath, pathWasSet := os.LookupEnv("PATH")
	if err := os.Setenv("PATH", ""); err != nil {
		t.Fatal(err)
	}
	defer func() {
		var err error
		if pathWasSet {
			err = os.Setenv("PATH", oldPath)
		} else {
			err = os.Unsetenv("PATH")
		}
		if err != nil {
			t.Errorf("restore PATH: %v", err)
		}
	}()

	for _, tc := range []struct {
		name    string
		options []MountOption
		wantErr string
	}{
		{
			name:    "helper-only option cannot bypass rejection",
			options: []MountOption{AllowRoot()},
			wantErr: "mountpoint is not empty",
		},
		{
			name:    "explicit opt-in reaches helper",
			options: []MountOption{AllowRoot(), AllowNonEmptyMount()},
			wantErr: "fusermount",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conf := mountConfig{options: make(map[string]string)}
			for _, option := range tc.options {
				if err := option(&conf); err != nil {
					t.Fatal(err)
				}
			}

			_, err := mount(dir, &conf, make(chan struct{}), nil)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestCheckMountpointEmpty(t *testing.T) {
	dir := tempMountpoint(t)
	defer os.RemoveAll(dir)

	if err := checkMountpointEmpty(dir); err != nil {
		t.Fatalf("empty mountpoint: %v", err)
	}

	existing := filepath.Join(dir, "existing")
	if err := ioutil.WriteFile(existing, []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := checkMountpointEmpty(dir); err == nil || !strings.Contains(err.Error(), "mountpoint is not empty") {
		t.Fatalf("non-empty mountpoint: want a descriptive error, got %v", err)
	}

	if err := checkMountpointEmpty(existing); err == nil || !strings.Contains(err.Error(), "open mountpoint") {
		t.Fatalf("non-directory mountpoint: want an open error, got %v", err)
	}

	missing := filepath.Join(dir, "missing")
	err := checkMountpointEmpty(missing)
	var target *MountpointDoesNotExistError
	if !errors.As(err, &target) || target.Path != missing {
		t.Fatalf("missing mountpoint: want MountpointDoesNotExistError for %q, got %v", missing, err)
	}
}
