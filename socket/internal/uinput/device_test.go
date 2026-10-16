package uinput

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"e70kbd/internal/mapkeys"
)

func TestAdvertisedKeys(t *testing.T) {
	d, err := Open(mapkeys.Codes())
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	path, err := eventPath()
	if err != nil || path == "" {
		t.Fatal(path, err)
	}
	bits := keyBits(t, path)
	for _, code := range []uint16{mapkeys.KeyA, mapkeys.KeyLeftShift, mapkeys.KeyEnter, mapkeys.KeyBackspace} {
		if !bits[code] {
			t.Fatalf("key %d not advertised", code)
		}
	}
}

func TestTapReadable(t *testing.T) {
	d, err := Open(mapkeys.Codes())
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	path, err := eventPath()
	if err != nil || path == "" {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Skip("event node is not readable; evtest needs the input group or the udev rule")
	}
	f.Close()
	if err := d.Tap(mapkeys.KeyA, false); err != nil {
		t.Fatal(err)
	}
}

func keyBits(t *testing.T, eventPath string) map[uint16]bool {
	t.Helper()
	sys := "/sys/class/input/" + eventPath[len("/dev/input/"):] + "/device/capabilities/key"
	text, err := os.ReadFile(sys)
	if err != nil {
		t.Fatal(err)
	}
	fields := strings.Fields(string(text))
	bits := map[uint16]bool{}
	// amd64 sysfs words are 64-bit, highest word first.
	for i, field := range fields {
		word, err := strconv.ParseUint(field, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		shift := uint(len(fields)-1-i) * 64
		for bit := 0; bit < 64; bit++ {
			if word&(uint64(1)<<uint(bit)) != 0 {
				bits[uint16(shift+uint(bit))] = true
			}
		}
	}
	return bits
}
