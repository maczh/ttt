package ui

import (
	"github.com/eugenioenko/ttt/internal/keyboard"
	"github.com/eugenioenko/ttt/internal/term"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestKeyToVTModifiers(t *testing.T) {
	tests := []struct {
		name string
		key  tcell.Key
		mod  tcell.ModMask
		want string
	}{
		{"enter", tcell.KeyEnter, tcell.ModNone, "\r"},
		{"shift+enter", tcell.KeyEnter, tcell.ModShift, "\x1b\r"},
		{"alt+enter", tcell.KeyEnter, tcell.ModAlt, "\x1b\r"},
		{"ctrl+enter", tcell.KeyEnter, tcell.ModCtrl, "\r"},
		{"left", tcell.KeyLeft, tcell.ModNone, "\x1b[D"},
		{"ctrl+left", tcell.KeyLeft, tcell.ModCtrl, "\x1b[1;5D"},
		{"shift+up", tcell.KeyUp, tcell.ModShift, "\x1b[1;2A"},
		{"ctrl+shift+right", tcell.KeyRight, tcell.ModCtrl | tcell.ModShift, "\x1b[1;6C"},
		{"ctrl+delete", tcell.KeyDelete, tcell.ModCtrl, "\x1b[3;5~"},
		{"backspace", tcell.KeyBackspace2, tcell.ModNone, "\x7f"},
		{"alt+backspace", tcell.KeyBackspace2, tcell.ModAlt, "\x1b\x7f"},
		{"ctrl+backspace", tcell.KeyBackspace2, tcell.ModCtrl, "\x7f"},
		{"shift+tab", tcell.KeyBacktab, tcell.ModShift, "\x1b[Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keyToVT(tcell.NewEventKey(tt.key, "", tt.mod))
			if got != tt.want {
				t.Errorf("keyToVT(%s) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestKeyToVTKittyIsIndependentFromLegacyEncoding(t *testing.T) {
	tests := []struct {
		name  string
		event keyboard.KeyEvent
		want  string
	}{
		{"ctrl-left", keyboard.KeyEvent{Code: keyboard.KeyLeft, Modifiers: keyboard.ModCtrl}, "\x1b[1;5D"},
		{"alt-unicode", keyboard.KeyEvent{Code: keyboard.KeyRune, Rune: '界', Modifiers: keyboard.ModAlt}, "\x1b[30028;3u"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keyboardEventToKitty(tt.event)
			if got != tt.want {
				t.Fatalf("kitty encoding = %q, want %q", got, tt.want)
			}
		})
	}
	if got := keyToVTForProtocol(tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModCtrl), term.KeyboardLegacy); got != "\x1b[1;5D" {
		t.Fatalf("legacy encoding changed = %q", got)
	}
}

func TestTerminalWidgetKeyboardProtocolConfiguration(t *testing.T) {
	tw := &TerminalWidget{}
	tw.SetKeyboardProtocol("kitty")
	if tw.KeyboardProtocol != term.KeyboardKitty {
		t.Fatalf("protocol = %q, want kitty", tw.KeyboardProtocol)
	}
	tw.SetKeyboardProtocol("unknown")
	if tw.KeyboardProtocol != term.KeyboardLegacy {
		t.Fatalf("unknown protocol = %q, want legacy", tw.KeyboardProtocol)
	}
}
