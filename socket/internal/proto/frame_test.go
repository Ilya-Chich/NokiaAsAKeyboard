package proto

import (
	"bytes"
	"errors"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	tapA, err := TapText("a")
	if err != nil {
		t.Fatal(err)
	}
	frames := []Frame{
		Hello(),
		tapA,
		TapEnter(),
		TapBackspace(),
		Ping(),
		Bye(),
	}
	for _, want := range frames {
		raw, err := Encode(want)
		if err != nil {
			t.Fatalf("type %d: encode: %v", want.Type, err)
		}
		got, err := Decode(bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("type %d: decode: %v", want.Type, err)
		}
		if got.Type != want.Type || string(got.Payload) != string(want.Payload) {
			t.Fatalf("type %d: got %v, want %v", want.Type, got, want)
		}
	}
}

func TestShortRead(t *testing.T) {
	raw, err := Encode(Ping())
	if err != nil {
		t.Fatal(err)
	}
	_, err = Decode(bytes.NewReader(raw[:3]))
	if !errors.Is(err, ErrShort) {
		t.Fatalf("got %v, want short read", err)
	}
}

func TestBadMagic(t *testing.T) {
	raw, err := Encode(Ping())
	if err != nil {
		t.Fatal(err)
	}
	raw[0] = 0
	_, err = Decode(bytes.NewReader(raw))
	if !errors.Is(err, ErrBadMagic) {
		t.Fatalf("got %v, want bad magic", err)
	}
}

func TestLengthOver32(t *testing.T) {
	raw := []byte{magic0, magic1, version, TypePing, 33}
	raw = append(raw, bytes.Repeat([]byte{0}, 33)...)
	_, err := Decode(bytes.NewReader(raw))
	if !errors.Is(err, ErrLength) {
		t.Fatalf("got %v, want length over 32", err)
	}
}
