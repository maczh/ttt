package keyboard

import (
	"errors"
	"strconv"
	"strings"
)

type KeyCode uint8

const (
	KeyUnknown KeyCode = iota
	KeyLeft
	KeyRight
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
	Modifiers Modifiers
}

var errUnsupportedKeySequence = errors.New("unsupported key sequence")

func ParseKeySequence(seq []byte) (KeyEvent, error) {
	s := string(seq)
	if !strings.HasPrefix(s, "\x1b[") {
		return KeyEvent{}, errUnsupportedKeySequence
	}
	code := KeyUnknown
	switch {
	case strings.HasSuffix(s, "D"):
		code = KeyLeft
	case strings.HasSuffix(s, "C"):
		code = KeyRight
	default:
		return KeyEvent{}, errUnsupportedKeySequence
	}

	params := strings.Split(strings.TrimSuffix(strings.TrimPrefix(s, "\x1b["), s[len(s)-1:]), ";")
	if len(params) != 2 || params[0] != "1" {
		return KeyEvent{}, errUnsupportedKeySequence
	}
	modifier, err := strconv.Atoi(params[1])
	if err != nil {
		return KeyEvent{}, errUnsupportedKeySequence
	}
	mods, ok := modifiersFromVTParam(modifier)
	if !ok {
		return KeyEvent{}, errUnsupportedKeySequence
	}
	return KeyEvent{Code: code, Modifiers: mods}, nil
}

func modifiersFromVTParam(param int) (Modifiers, bool) {
	if param < 1 || param > 8 {
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
	return mods, true
}
