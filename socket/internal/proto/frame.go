// Package proto is version 1 of the E70 keyboard frame codec.
// It does not open Bluetooth or uinput.
package proto

import (
	"errors"
	"io"
	"unicode/utf8"
)

const (
	magic0  = 0x4B // 'K'
	magic1  = 0x42 // 'B'
	version = 1

	headerLen  = 5
	maxPayload = 32

	TypeHello byte = 1
	TypeTap   byte = 2
	TypePing  byte = 3
	TypeBye   byte = 4

	TapChar    byte = 0
	TapSpecial byte = 1

	SpecialEnter     byte = 1
	SpecialBackspace byte = 2
)

var (
	ErrBadMagic   = errors.New("bad magic")
	ErrBadVersion = errors.New("bad version")
	ErrLength     = errors.New("length over 32")
	ErrShort      = errors.New("short read")
	ErrPayload    = errors.New("bad payload")
)

// Frame is one version-1 message. Payload is owned by the caller after Encode
// and by the Frame after Decode.
type Frame struct {
	Type    byte
	Payload []byte
}

func Hello() Frame {
	return Frame{Type: TypeHello, Payload: []byte{version}}
}

func TapText(ch string) (Frame, error) {
	if !oneRune(ch) {
		return Frame{}, ErrPayload
	}
	payload := make([]byte, 1+len(ch))
	payload[0] = TapChar
	copy(payload[1:], ch)
	return Frame{Type: TypeTap, Payload: payload}, nil
}

func TapEnter() Frame {
	return Frame{Type: TypeTap, Payload: []byte{TapSpecial, SpecialEnter}}
}

func TapBackspace() Frame {
	return Frame{Type: TypeTap, Payload: []byte{TapSpecial, SpecialBackspace}}
}

func Ping() Frame {
	return Frame{Type: TypePing}
}

func Bye() Frame {
	return Frame{Type: TypeBye}
}

// Encode writes a big-endian version-1 frame. Every field is one byte.
func Encode(f Frame) ([]byte, error) {
	if err := validate(f); err != nil {
		return nil, err
	}
	out := make([]byte, headerLen+len(f.Payload))
	out[0] = magic0
	out[1] = magic1
	out[2] = version
	out[3] = f.Type
	out[4] = byte(len(f.Payload))
	copy(out[5:], f.Payload)
	return out, nil
}

// Decode reads one frame. A bad magic, a bad version, a length over 32, or a
// short read is an error. It does not scan ahead for another magic.
func Decode(r io.Reader) (Frame, error) {
	hdr := make([]byte, headerLen)
	if err := readFull(r, hdr); err != nil {
		return Frame{}, err
	}
	if hdr[0] != magic0 || hdr[1] != magic1 {
		return Frame{}, ErrBadMagic
	}
	if hdr[2] != version {
		return Frame{}, ErrBadVersion
	}
	n := int(hdr[4])
	if n > maxPayload {
		return Frame{}, ErrLength
	}
	payload := make([]byte, n)
	if err := readFull(r, payload); err != nil {
		return Frame{}, err
	}
	f := Frame{Type: hdr[3], Payload: payload}
	if err := validate(f); err != nil {
		return Frame{}, err
	}
	return f, nil
}

func validate(f Frame) error {
	if len(f.Payload) > maxPayload {
		return ErrLength
	}
	switch f.Type {
	case TypeHello:
		if len(f.Payload) != 1 || f.Payload[0] != version {
			return ErrPayload
		}
	case TypeTap:
		if len(f.Payload) < 1 {
			return ErrPayload
		}
		switch f.Payload[0] {
		case TapChar:
			if !oneRune(string(f.Payload[1:])) {
				return ErrPayload
			}
		case TapSpecial:
			if len(f.Payload) != 2 || (f.Payload[1] != SpecialEnter && f.Payload[1] != SpecialBackspace) {
				return ErrPayload
			}
		default:
			return ErrPayload
		}
	case TypePing, TypeBye:
		if len(f.Payload) != 0 {
			return ErrPayload
		}
	default:
		return ErrPayload
	}
	return nil
}

func oneRune(s string) bool {
	r, size := utf8.DecodeRuneInString(s)
	return r != utf8.RuneError && size == len(s)
}

func readFull(r io.Reader, buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	_, err := io.ReadFull(r, buf)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return ErrShort
	}
	return err
}
