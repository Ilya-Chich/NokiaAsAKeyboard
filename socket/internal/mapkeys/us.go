// Package mapkeys maps one Unicode character to a US QWERTY evdev key.
// Uppercase and shifted punctuation set Shift. The phone does not send Shift.
package mapkeys

// Linux evdev codes used by the US table. Values are from input-event-codes.h.
const (
	KeyEnter      uint16 = 28
	KeyBackspace  uint16 = 14
	KeyLeftShift  uint16 = 42
	KeySpace      uint16 = 57
	Key1          uint16 = 2
	Key0          uint16 = 11
	KeyMinus      uint16 = 12
	KeyEqual      uint16 = 13
	KeyQ          uint16 = 16
	KeyW          uint16 = 17
	KeyE          uint16 = 18
	KeyR          uint16 = 19
	KeyT          uint16 = 20
	KeyY          uint16 = 21
	KeyU          uint16 = 22
	KeyI          uint16 = 23
	KeyO          uint16 = 24
	KeyP          uint16 = 25
	KeyLeftBrace  uint16 = 26
	KeyRightBrace uint16 = 27
	KeyA          uint16 = 30
	KeyS          uint16 = 31
	KeyD          uint16 = 32
	KeyF          uint16 = 33
	KeyG          uint16 = 34
	KeyH          uint16 = 35
	KeyJ          uint16 = 36
	KeyK          uint16 = 37
	KeyL          uint16 = 38
	KeySemicolon  uint16 = 39
	KeyApostrophe uint16 = 40
	KeyGrave      uint16 = 41
	KeyBackslash  uint16 = 43
	KeyZ          uint16 = 44
	KeyX          uint16 = 45
	KeyC          uint16 = 46
	KeyV          uint16 = 47
	KeyB          uint16 = 48
	KeyN          uint16 = 49
	KeyM          uint16 = 50
	KeyComma      uint16 = 51
	KeyDot        uint16 = 52
	KeySlash      uint16 = 53
)

// Key is one evdev key, with Shift if the US layout needs it.
type Key struct {
	Code  uint16
	Shift bool
}

func Enter() Key     { return Key{Code: KeyEnter} }
func Backspace() Key { return Key{Code: KeyBackspace} }

// Lookup returns false for a character this build does not type.
func Lookup(r rune) (Key, bool) {
	k, ok := table[r]
	return k, ok
}

// Codes is every key the virtual device must advertise, including Shift.
func Codes() []uint16 {
	seen := map[uint16]struct{}{KeyLeftShift: {}, KeyEnter: {}, KeyBackspace: {}}
	out := []uint16{KeyLeftShift, KeyEnter, KeyBackspace}
	for _, k := range table {
		if _, ok := seen[k.Code]; ok {
			continue
		}
		seen[k.Code] = struct{}{}
		out = append(out, k.Code)
	}
	return out
}

func key(code uint16, shift bool) Key { return Key{Code: code, Shift: shift} }

var table = map[rune]Key{
	'a': key(KeyA, false), 'A': key(KeyA, true),
	'b': key(KeyB, false), 'B': key(KeyB, true),
	'c': key(KeyC, false), 'C': key(KeyC, true),
	'd': key(KeyD, false), 'D': key(KeyD, true),
	'e': key(KeyE, false), 'E': key(KeyE, true),
	'f': key(KeyF, false), 'F': key(KeyF, true),
	'g': key(KeyG, false), 'G': key(KeyG, true),
	'h': key(KeyH, false), 'H': key(KeyH, true),
	'i': key(KeyI, false), 'I': key(KeyI, true),
	'j': key(KeyJ, false), 'J': key(KeyJ, true),
	'k': key(KeyK, false), 'K': key(KeyK, true),
	'l': key(KeyL, false), 'L': key(KeyL, true),
	'm': key(KeyM, false), 'M': key(KeyM, true),
	'n': key(KeyN, false), 'N': key(KeyN, true),
	'o': key(KeyO, false), 'O': key(KeyO, true),
	'p': key(KeyP, false), 'P': key(KeyP, true),
	'q': key(KeyQ, false), 'Q': key(KeyQ, true),
	'r': key(KeyR, false), 'R': key(KeyR, true),
	's': key(KeyS, false), 'S': key(KeyS, true),
	't': key(KeyT, false), 'T': key(KeyT, true),
	'u': key(KeyU, false), 'U': key(KeyU, true),
	'v': key(KeyV, false), 'V': key(KeyV, true),
	'w': key(KeyW, false), 'W': key(KeyW, true),
	'x': key(KeyX, false), 'X': key(KeyX, true),
	'y': key(KeyY, false), 'Y': key(KeyY, true),
	'z': key(KeyZ, false), 'Z': key(KeyZ, true),

	'1': key(Key1, false), '!': key(Key1, true),
	'2': key(Key1+1, false), '@': key(Key1+1, true),
	'3': key(Key1+2, false), '#': key(Key1+2, true),
	'4': key(Key1+3, false), '$': key(Key1+3, true),
	'5': key(Key1+4, false), '%': key(Key1+4, true),
	'6': key(Key1+5, false), '^': key(Key1+5, true),
	'7': key(Key1+6, false), '&': key(Key1+6, true),
	'8': key(Key1+7, false), '*': key(Key1+7, true),
	'9': key(Key1+8, false), '(': key(Key1+8, true),
	'0': key(Key0, false), ')': key(Key0, true),

	' ': key(KeySpace, false),
	'-': key(KeyMinus, false), '_': key(KeyMinus, true),
	'=': key(KeyEqual, false), '+': key(KeyEqual, true),
	'[': key(KeyLeftBrace, false), '{': key(KeyLeftBrace, true),
	']': key(KeyRightBrace, false), '}': key(KeyRightBrace, true),
	'\\': key(KeyBackslash, false), '|': key(KeyBackslash, true),
	';': key(KeySemicolon, false), ':': key(KeySemicolon, true),
	'\'': key(KeyApostrophe, false), '"': key(KeyApostrophe, true),
	'`': key(KeyGrave, false), '~': key(KeyGrave, true),
	',': key(KeyComma, false), '<': key(KeyComma, true),
	'.': key(KeyDot, false), '>': key(KeyDot, true),
	'/': key(KeySlash, false), '?': key(KeySlash, true),
}
