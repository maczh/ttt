package keyboard

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"
)

type KeyCode uint8

const (
	KeyUnknown KeyCode = iota
	KeyRune
	KeyEscape
	KeyEnter
	KeyTab
	KeyBackspace
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyInsert
	KeyDelete
	KeyFunction
)

type Modifiers uint8

const (
	ModShift Modifiers = 1 << iota
	ModAlt
	ModCtrl
	ModMeta
)

type KeyEvent struct {
	Code      KeyCode
	Rune      rune
	Function  int
	Modifiers Modifiers
}

var errUnsupportedKeySequence = errors.New("unsupported key sequence")

func ParseKeySequence(seq []byte) (KeyEvent, error) {
	if len(seq) == 0 {
		return KeyEvent{}, errUnsupportedKeySequence
	}
	if seq[0] != '\x1b' {
		return parseLegacyByte(seq)
	}
	if len(seq) == 1 {
		return KeyEvent{Code: KeyEscape}, nil
	}
	if seq[1] == '[' {
		return parseCSI(seq[2:])
	}
	if seq[1] == 'O' {
		return parseSS3(seq[2:])
	}
	if len(seq) > 1 {
		r, size := utf8.DecodeRune(seq[1:])
		if r != utf8.RuneError && size == len(seq)-1 && r >= 0x20 {
			return KeyEvent{Code: KeyRune, Rune: r, Modifiers: ModAlt}, nil
		}
	}
	return KeyEvent{}, errUnsupportedKeySequence
}

func parseLegacyByte(seq []byte) (KeyEvent, error) {
	if len(seq) != 1 {
		return KeyEvent{}, errUnsupportedKeySequence
	}
	switch seq[0] {
	case 0:
		return KeyEvent{Code: KeyRune, Rune: ' ', Modifiers: ModCtrl}, nil
	case '\t':
		return KeyEvent{Code: KeyTab}, nil
	case '\r':
		return KeyEvent{Code: KeyEnter}, nil
	case '\x7f', '\b':
		return KeyEvent{Code: KeyBackspace}, nil
	case 0x1f:
		return KeyEvent{Code: KeyRune, Rune: '/', Modifiers: ModCtrl}, nil
	}
	if seq[0] >= 1 && seq[0] <= 26 {
		return KeyEvent{Code: KeyRune, Rune: rune('a' + seq[0] - 1), Modifiers: ModCtrl}, nil
	}
	if seq[0] >= 0x20 {
		return KeyEvent{Code: KeyRune, Rune: rune(seq[0])}, nil
	}
	return KeyEvent{}, errUnsupportedKeySequence
}

func parseCSI(body []byte) (KeyEvent, error) {
	if len(body) == 0 {
		return KeyEvent{}, errUnsupportedKeySequence
	}
	final := body[len(body)-1]
	if final == 'u' {
		return parseKittyCSIU(body[:len(body)-1])
	}
	params, ok := parseNumericParams(body[:len(body)-1])
	if !ok {
		return KeyEvent{}, errUnsupportedKeySequence
	}
	if final == '~' && len(params) == 3 && params[0] == 27 {
		mods, ok := modifiersFromVTParam(params[1])
		if !ok || params[2] < 0 || params[2] > 0x10ffff {
			return KeyEvent{}, errUnsupportedKeySequence
		}
		return KeyEvent{Code: KeyRune, Rune: rune(params[2]), Modifiers: mods}, nil
	}
	mods, ok := modifiersFromVTParam(modifierParam(params))
	if !ok {
		return KeyEvent{}, errUnsupportedKeySequence
	}

	switch final {
	case 'A', 'B', 'C', 'D', 'H', 'F':
		if len(params) > 2 || (len(params) == 2 && params[0] != 1) {
			return KeyEvent{}, errUnsupportedKeySequence
		}
		code := map[byte]KeyCode{'A': KeyUp, 'B': KeyDown, 'C': KeyRight, 'D': KeyLeft, 'H': KeyHome, 'F': KeyEnd}[final]
		return KeyEvent{Code: code, Modifiers: mods}, nil
	case '~':
		if len(params) == 0 {
			return KeyEvent{}, errUnsupportedKeySequence
		}
		code, ok := tildeKeyCode(params[0])
		if !ok {
			return KeyEvent{}, errUnsupportedKeySequence
		}
		return KeyEvent{Code: code, Function: tildeFunction(params[0]), Modifiers: mods}, nil
	default:
		return KeyEvent{}, errUnsupportedKeySequence
	}
}

func parseSS3(body []byte) (KeyEvent, error) {
	if len(body) != 1 {
		return KeyEvent{}, errUnsupportedKeySequence
	}
	function := map[byte]int{'P': 1, 'Q': 2, 'R': 3, 'S': 4}[body[0]]
	if function == 0 {
		return KeyEvent{}, errUnsupportedKeySequence
	}
	return KeyEvent{Code: KeyFunction, Function: function}, nil
}

func parseKittyCSIU(body []byte) (KeyEvent, error) {
	params, ok := parseNumericParams(body)
	if !ok || len(params) < 1 || len(params) > 2 || params[0] < 0 || params[0] > 0x10ffff {
		return KeyEvent{}, errUnsupportedKeySequence
	}
	mods, ok := modifiersFromVTParam(modifierParam(params))
	if !ok {
		return KeyEvent{}, errUnsupportedKeySequence
	}
	switch params[0] {
	case 8, 127:
		return KeyEvent{Code: KeyBackspace, Modifiers: mods}, nil
	case 9:
		return KeyEvent{Code: KeyTab, Modifiers: mods}, nil
	case 13:
		return KeyEvent{Code: KeyEnter, Modifiers: mods}, nil
	case 27:
		return KeyEvent{Code: KeyEscape, Modifiers: mods}, nil
	}
	return KeyEvent{Code: KeyRune, Rune: rune(params[0]), Modifiers: mods}, nil
}

func parseNumericParams(body []byte) ([]int, bool) {
	if len(body) == 0 {
		return nil, true
	}
	parts := strings.Split(string(body), ";")
	params := make([]int, len(parts))
	for i, part := range parts {
		if part == "" {
			return nil, false
		}
		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, false
		}
		params[i] = value
	}
	return params, true
}

func modifierParam(params []int) int {
	if len(params) < 2 {
		return 1
	}
	return params[len(params)-1]
}

func tildeKeyCode(number int) (KeyCode, bool) {
	if number >= 15 && number <= 24 {
		return KeyFunction, true
	}
	code, ok := map[int]KeyCode{
		2: KeyInsert, 3: KeyDelete, 5: KeyPageUp, 6: KeyPageDown,
		7: KeyHome, 8: KeyEnd,
	}[number]
	return code, ok
}

func tildeFunction(number int) int {
	switch number {
	case 15:
		return 5
	case 17:
		return 6
	case 18:
		return 7
	case 19:
		return 8
	case 20:
		return 9
	case 21:
		return 10
	case 23:
		return 11
	case 24:
		return 12
	default:
		return 0
	}
}

func modifiersFromVTParam(param int) (Modifiers, bool) {
	if param < 1 || param > 256 {
		return 0, false
	}
	param--
	var mods Modifiers
	if param&1 != 0 {
		mods |= ModShift
	}
	if param&2 != 0 {
		mods |= ModAlt
	}
	if param&4 != 0 {
		mods |= ModCtrl
	}
	if param&32 != 0 {
		mods |= ModMeta
	}
	return mods, true
}
