package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unicode/utf8"

	"e70kbd/internal/bt"
	"e70kbd/internal/mapkeys"
	"e70kbd/internal/uinput"
)

func main() {
	tap := flagTap()
	if tap == "" {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		defer stop()
		if err := bt.Listen(ctx); err != nil {
			fmt.Fprintln(os.Stderr, "e70kbd:", err)
			os.Exit(1)
		}
		return
	}
	r, size := utf8.DecodeRuneInString(tap)
	if r == utf8.RuneError || size != len(tap) {
		fmt.Fprintln(os.Stderr, "e70kbd: -tap needs one character")
		os.Exit(1)
	}
	key, ok := mapkeys.Lookup(r)
	if !ok {
		fmt.Fprintf(os.Stderr, "e70kbd: no US key for %q\n", tap)
		os.Exit(1)
	}
	dev, err := uinput.Open(mapkeys.Codes())
	if err != nil {
		fmt.Fprintln(os.Stderr, "e70kbd:", err)
		os.Exit(1)
	}
	err = dev.Tap(key.Code, key.Shift)
	closeErr := dev.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, "e70kbd:", err)
		os.Exit(1)
	}
	if closeErr != nil {
		fmt.Fprintln(os.Stderr, "e70kbd:", closeErr)
		os.Exit(1)
	}
}

func flagTap() string {
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "-tap" && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
		if len(arg) > 5 && arg[:5] == "-tap=" {
			return arg[5:]
		}
	}
	return ""
}
