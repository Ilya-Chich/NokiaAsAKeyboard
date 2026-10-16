package bt

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/godbus/dbus/v5"

	"e70kbd/internal/mapkeys"
	"e70kbd/internal/uinput"
)

const (
	// UUID is the custom service the phone will search for.
	UUID = "8b0d1a3e-70e7-4c1a-9b5e-6f0a0e70c0de"
	// Channel is the fixed RFCOMM channel requested from BlueZ.
	Channel = 22
	name    = "E70 Keyboard"
	path    = dbus.ObjectPath("/org/e70kbd/profile")

	solBluetooth = 274
	btSecurity   = 4
	secHigh      = 3 // authenticated link key and encryption
)

// Listen registers Profile1 and blocks until ctx is cancelled.
func Listen(ctx context.Context) error {
	keys, err := uinput.Open(mapkeys.Codes())
	if err != nil {
		return err
	}
	defer keys.Close()

	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	s := &server{conn: conn, keys: keys}
	if err := conn.Export(s, path, "org.bluez.Profile1"); err != nil {
		return err
	}
	opts := map[string]dbus.Variant{
		"Name":                  dbus.MakeVariant(name),
		"Service":               dbus.MakeVariant(UUID),
		"Role":                  dbus.MakeVariant("server"),
		"Channel":               dbus.MakeVariant(uint16(Channel)),
		"RequireAuthentication": dbus.MakeVariant(true),
		"RequireAuthorization":  dbus.MakeVariant(false),
		"AutoConnect":           dbus.MakeVariant(false),
	}
	obj := conn.Object("org.bluez", "/org/bluez")
	call := obj.Call("org.bluez.ProfileManager1.RegisterProfile", 0, path, UUID, opts)
	if call.Err != nil {
		return call.Err
	}
	fmt.Fprintf(os.Stderr, "rfcomm channel %d\n", Channel)
	defer func() {
		obj.Call("org.bluez.ProfileManager1.UnregisterProfile", 0, path)
		s.closeCurrent()
	}()

	<-ctx.Done()
	return nil
}

type server struct {
	conn *dbus.Conn
	mu   sync.Mutex
	cur  *os.File
	dev  dbus.ObjectPath
	keys Keyboard
}

func (s *server) Release() *dbus.Error {
	s.closeCurrent()
	return nil
}

func (s *server) RequestDisconnection(device dbus.ObjectPath) *dbus.Error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dev == device {
		s.closeLocked()
	}
	return nil
}

func (s *server) NewConnection(device dbus.ObjectPath, fd dbus.UnixFD, _ map[string]dbus.Variant) *dbus.Error {
	f := os.NewFile(uintptr(fd), "rfcomm")
	paired, err := devicePaired(s.conn, device)
	if err != nil || !paired {
		fmt.Fprintf(os.Stderr, "refused unpaired %s\n", device)
		f.Close()
		if err == nil {
			err = errors.New("unpaired")
		}
		return dbus.MakeFailedError(err)
	}
	if err := requireEncrypted(int(f.Fd())); err != nil {
		fmt.Fprintf(os.Stderr, "refused insecure %s: %v\n", device, err)
		f.Close()
		return dbus.MakeFailedError(err)
	}
	_ = syscall.SetNonblock(int(f.Fd()), false)
	s.mu.Lock()
	if s.cur != nil {
		fmt.Fprintf(os.Stderr, "closed previous %s\n", s.dev)
		s.closeLocked()
	}
	s.cur = f
	s.dev = device
	s.mu.Unlock()
	fmt.Fprintf(os.Stderr, "connected %s\n", device)
	go s.read(f)
	return nil
}

func (s *server) read(f *os.File) {
	err := serve(f, s.keys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "disconnect %v\n", err)
	}
	if s.keys != nil {
		if relErr := s.keys.ReleaseAll(); relErr != nil {
			fmt.Fprintf(os.Stderr, "release keys: %v\n", relErr)
		}
	}
	f.Close()
}

func (s *server) closeCurrent() {
	s.mu.Lock()
	s.closeLocked()
	s.mu.Unlock()
}

func (s *server) closeLocked() {
	if s.cur != nil {
		s.cur.Close()
		s.cur = nil
	}
	s.dev = ""
}

func devicePaired(conn *dbus.Conn, device dbus.ObjectPath) (bool, error) {
	v, err := conn.Object("org.bluez", device).GetProperty("org.bluez.Device1.Paired")
	if err != nil {
		return false, err
	}
	paired, ok := v.Value().(bool)
	if !ok {
		return false, errors.New("Paired is not a bool")
	}
	return paired, nil
}

func requireEncrypted(fd int) error {
	level := [2]byte{secHigh, 0}
	_, _, errno := syscall.Syscall6(syscall.SYS_SETSOCKOPT, uintptr(fd), solBluetooth, btSecurity,
		uintptr(unsafe.Pointer(&level[0])), 2, 0)
	if errno != 0 {
		return errno
	}
	return nil
}
