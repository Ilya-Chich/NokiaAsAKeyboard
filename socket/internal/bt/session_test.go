package bt

import (
	"bytes"
	"net"
	"testing"
	"time"

	"e70kbd/internal/proto"
)

func TestServeHelloPingIgnoresTap(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	keys := &fakeKeys{}
	errc := make(chan error, 1)
	go func() { errc <- serve(a, keys) }()

	hello, err := proto.Encode(proto.Hello())
	if err != nil {
		t.Fatal(err)
	}
	ping, err := proto.Encode(proto.Ping())
	if err != nil {
		t.Fatal(err)
	}
	tap, err := proto.Encode(proto.TapEnter())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.Write(hello); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Write(ping); err != nil {
		t.Fatal(err)
	}
	_ = b.SetReadDeadline(time.Now().Add(time.Second))
	reply := make([]byte, len(ping))
	if _, err := b.Read(reply); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(reply, ping) {
		t.Fatalf("reply %x, want ping %x", reply, ping)
	}
	if _, err := b.Write(tap); err != nil {
		t.Fatal(err)
	}
	_ = b.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
	extra := make([]byte, 8)
	n, _ := b.Read(extra)
	if n != 0 {
		t.Fatalf("tap produced %x", extra[:n])
	}
	if len(keys.taps) != 1 || keys.taps[0].code != 28 || keys.taps[0].shift {
		t.Fatalf("taps %+v", keys.taps)
	}
}

type fakeKeys struct {
	taps []tapCall
}

type tapCall struct {
	code  uint16
	shift bool
}

func (f *fakeKeys) Tap(code uint16, shift bool) error {
	f.taps = append(f.taps, tapCall{code, shift})
	return nil
}

func (f *fakeKeys) ReleaseAll() error { return nil }
