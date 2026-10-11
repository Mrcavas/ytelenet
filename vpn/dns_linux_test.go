package vpn

import (
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvConfFile(t *testing.T) {
	dir := t.TempDir()
	path, backup := filepath.Join(dir, "resolv.conf"), filepath.Join(dir, "backup")
	original := "nameserver 192.168.1.1\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	restore, err := setResolvConfDNS(path, backup, netip.MustParseAddr("1.1.1.1"))
	if err != nil {
		t.Fatal(err)
	}
	if content, _ := os.ReadFile(path); !strings.Contains(string(content), "nameserver 1.1.1.1\n") {
		t.Fatalf("resolv.conf wasn't replaced: %q", content)
	}

	restore()
	if content, _ := os.ReadFile(path); string(content) != original {
		t.Fatalf("resolv.conf wasn't restored: %q", content)
	}
	if _, err := os.Lstat(backup); err == nil {
		t.Fatal("backup is left over")
	}
}

func TestResolvConfSymlink(t *testing.T) {
	dir := t.TempDir()
	path, backup := filepath.Join(dir, "resolv.conf"), filepath.Join(dir, "backup")
	target := filepath.Join(dir, "stub-resolv.conf")
	if err := os.WriteFile(target, []byte("nameserver 127.0.0.53\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}

	restore, err := setResolvConfDNS(path, backup, netip.MustParseAddr("1.1.1.1"))
	if err != nil {
		t.Fatal(err)
	}
	if info, _ := os.Lstat(path); info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("resolv.conf is still a symlink")
	}

	restore()
	if link, err := os.Readlink(path); err != nil || link != target {
		t.Fatalf("symlink wasn't restored: %q, %v", link, err)
	}
}

func TestResolvConfLeftoverBackup(t *testing.T) {
	dir := t.TempDir()
	path, backup := filepath.Join(dir, "resolv.conf"), filepath.Join(dir, "backup")
	original := "nameserver 192.168.1.1\n"

	// a killed ytelenet left its resolv.conf and the backup
	if err := os.WriteFile(backup, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("nameserver 1.1.1.1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	restore, err := setResolvConfDNS(path, backup, netip.MustParseAddr("8.8.8.8"))
	if err != nil {
		t.Fatal(err)
	}
	restore()

	if content, _ := os.ReadFile(path); string(content) != original {
		t.Fatalf("the original resolv.conf is lost: %q", content)
	}
}
