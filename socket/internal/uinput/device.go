// Package uinput creates one virtual keyboard on /dev/uinput.
// Access for this user is documented in socket/udev/99-e70kbd.rules.
package uinput

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	deviceName = "E70 Keyboard"

	uiDevCreate  = 0x5501
	uiDevDestroy = 0x5502
	uiDevSetup   = 0x405c5503 // _IOW('U', 3, struct uinput_setup)
	uiSetEvbit   = 0x40045564
	uiSetKeybit  = 0x40045565

	evSyn        = 0
	evKey        = 1
	synReport    = 0
	keyDown      = 1
	keyUp        = 0
	busVirtual   = 6
	keyLeftShift = 42

	setupSize = 92
	nameOff   = 8
	nameMax   = 80
)

// Device is one virtual keyboard. Close releases every advertised key.
type Device struct {
	f    *os.File
	keys []uint16
}

// Open creates "E70 Keyboard" and advertises keys. It does not emit a tap.
func Open(keys []uint16) (*Device, error) {
	f, err := os.OpenFile("/dev/uinput", os.O_WRONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	d := &Device{f: f, keys: append([]uint16(nil), keys...)}
	if err := d.setup(); err != nil {
		f.Close()
		return nil, err
	}
	return d, nil
}

func (d *Device) setup() error {
	if err := ioctl(d.f.Fd(), uiSetEvbit, uintptr(evKey)); err != nil {
		return err
	}
	seen := map[uint16]struct{}{}
	for _, code := range d.keys {
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		if err := ioctl(d.f.Fd(), uiSetKeybit, uintptr(code)); err != nil {
			return err
		}
	}
	buf := make([]byte, setupSize)
	binary.LittleEndian.PutUint16(buf[0:], busVirtual)
	binary.LittleEndian.PutUint16(buf[2:], 1)
	binary.LittleEndian.PutUint16(buf[4:], 1)
	binary.LittleEndian.PutUint16(buf[6:], 1)
	copy(buf[nameOff:nameOff+nameMax], deviceName)
	if err := ioctl(d.f.Fd(), uiDevSetup, uintptr(unsafe.Pointer(&buf[0]))); err != nil {
		return err
	}
	if err := ioctl(d.f.Fd(), uiDevCreate, 0); err != nil {
		return err
	}
	return waitEventNode()
}

// Tap emits key down, SYN, key up, SYN. Shift is pressed around a shifted key
// and released if a later write fails. On any error every advertised key is released.
func (d *Device) Tap(code uint16, shift bool) error {
	if shift {
		if err := d.writeKey(keyLeftShift, keyDown); err != nil {
			return d.fail(err)
		}
	}
	if err := d.writeKey(code, keyDown); err != nil {
		return d.fail(err)
	}
	if err := d.writeKey(code, keyUp); err != nil {
		return d.fail(err)
	}
	if shift {
		if err := d.writeKey(keyLeftShift, keyUp); err != nil {
			return d.fail(err)
		}
	}
	return nil
}

func (d *Device) fail(err error) error {
	_ = d.ReleaseAll()
	return err
}

// ReleaseAll writes key-up for Shift and every advertised key.
func (d *Device) ReleaseAll() error {
	var first error
	release := func(code uint16) {
		if err := d.writeKey(code, keyUp); err != nil && first == nil {
			first = err
		}
	}
	release(keyLeftShift)
	seen := map[uint16]struct{}{keyLeftShift: {}}
	for _, code := range d.keys {
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		release(code)
	}
	return first
}

// Close releases every advertised key and destroys the virtual device.
func (d *Device) Close() error {
	err := d.ReleaseAll()
	if e := ioctl(d.f.Fd(), uiDevDestroy, 0); e != nil && err == nil {
		err = e
	}
	if e := d.f.Close(); e != nil && err == nil {
		err = e
	}
	return err
}

func (d *Device) writeKey(code uint16, value int32) error {
	if err := d.writeEvent(evKey, code, value); err != nil {
		return err
	}
	return d.writeEvent(evSyn, synReport, 0)
}

func (d *Device) writeEvent(typ, code uint16, value int32) error {
	var buf [24]byte
	binary.LittleEndian.PutUint16(buf[16:], typ)
	binary.LittleEndian.PutUint16(buf[18:], code)
	binary.LittleEndian.PutUint32(buf[20:], uint32(value))
	_, err := d.f.Write(buf[:])
	return err
}

func ioctl(fd, req, arg uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, arg)
	if errno != 0 {
		return errno
	}
	return nil
}

func waitEventNode() error {
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if path, err := eventPath(); err == nil && path != "" {
			// Let a reader such as evtest open the node before the tap.
			time.Sleep(200 * time.Millisecond)
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return errors.New("E70 Keyboard event node did not appear")
}

func eventPath() (string, error) {
	matches, err := filepath.Glob("/sys/class/input/event*/device/name")
	if err != nil {
		return "", err
	}
	for _, namePath := range matches {
		b, err := os.ReadFile(namePath)
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(b)) != deviceName {
			continue
		}
		event := filepath.Base(filepath.Dir(filepath.Dir(namePath)))
		return "/dev/input/" + event, nil
	}
	return "", nil
}
