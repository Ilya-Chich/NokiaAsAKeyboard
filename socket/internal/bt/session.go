package bt

import (
	"fmt"
	"io"
	"os"
	"time"
	"unicode/utf8"

	"e70kbd/internal/mapkeys"
	"e70kbd/internal/proto"
)

const idle = 15 * time.Second

type deadlineReader interface {
	io.Reader
	SetReadDeadline(time.Time) error
}

// Keyboard receives one decoded tap. ReleaseAll must not leave a key down.
type Keyboard interface {
	Tap(code uint16, shift bool) error
	ReleaseAll() error
}

// serve decodes version-1 frames, answers PING, and injects TAP.
func serve(rw io.ReadWriter, keys Keyboard) error {
	for {
		if d, ok := rw.(deadlineReader); ok {
			if err := d.SetReadDeadline(time.Now().Add(idle)); err != nil {
				return err
			}
		}
		frame, err := proto.Decode(rw)
		if err != nil {
			return err
		}
		switch frame.Type {
		case proto.TypeHello:
			fmt.Fprintln(os.Stderr, "hello")
		case proto.TypePing:
			fmt.Fprintln(os.Stderr, "ping")
			raw, err := proto.Encode(proto.Ping())
			if err != nil {
				return err
			}
			if _, err := rw.Write(raw); err != nil {
				return err
			}
		case proto.TypeTap:
			if err := inject(keys, frame); err != nil {
				return err
			}
		case proto.TypeBye:
			fmt.Fprintln(os.Stderr, "bye")
			return nil
		default:
			return proto.ErrPayload
		}
	}
}

func inject(keys Keyboard, frame proto.Frame) error {
	if len(frame.Payload) < 1 {
		return proto.ErrPayload
	}
	var key mapkeys.Key
	var label string
	switch frame.Payload[0] {
	case proto.TapChar:
		r, _ := utf8.DecodeRune(frame.Payload[1:])
		got, ok := mapkeys.Lookup(r)
		if !ok {
			fmt.Fprintf(os.Stderr, "dropped %q\n", string(r))
			return nil
		}
		key = got
		label = string(r)
	case proto.TapSpecial:
		if len(frame.Payload) != 2 {
			return proto.ErrPayload
		}
		if frame.Payload[1] == proto.SpecialEnter {
			key = mapkeys.Enter()
			label = "enter"
		} else if frame.Payload[1] == proto.SpecialBackspace {
			key = mapkeys.Backspace()
			label = "backspace"
		} else {
			return proto.ErrPayload
		}
	default:
		return proto.ErrPayload
	}
	fmt.Fprintf(os.Stderr, "tap %s keycode %d shift %v\n", label, key.Code, key.Shift)
	if keys == nil {
		return nil
	}
	return keys.Tap(key.Code, key.Shift)
}
