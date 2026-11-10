package keyboard

import "testing"

func TestProtocolMatrixNormalizesSameShortcut(t *testing.T) {
	tests := []struct {
		name string
		seq  string
	}{
		{name: "legacy control byte", seq: "\x03"},
		{name: "xterm modified key", seq: "\x1b[27;5;99~"},
		{name: "kitty CSI u", seq: "\x1b[99;5u"},
	}
	want := KeyEvent{Code: KeyRune, Rune: 'c', Modifiers: ModCtrl}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeySequence([]byte(tt.seq))
			if err != nil {
				t.Fatalf("parse error = %v", err)
			}
			if got != want {
				t.Fatalf("normalized = %#v, want %#v", got, want)
			}
		})
	}
}

func TestParseKeySequenceNormalizesCtrlLeft(t *testing.T) {
	tests := []struct {
		name string
		seq  string
		want KeyEvent
	}{
		{name: "legacy csi", seq: "\x1b[1;5D", want: KeyEvent{Code: KeyLeft, Modifiers: ModCtrl}},
		{name: "kitty functional key", seq: "\x1b[1;5D", want: KeyEvent{Code: KeyLeft, Modifiers: ModCtrl}},
		{name: "ctrl right", seq: "\x1b[1;5C", want: KeyEvent{Code: KeyRight, Modifiers: ModCtrl}},
		{name: "shift left", seq: "\x1b[1;2D", want: KeyEvent{Code: KeyLeft, Modifiers: ModShift}},
		{name: "alt left", seq: "\x1b[1;3D", want: KeyEvent{Code: KeyLeft, Modifiers: ModAlt}},
		{name: "ctrl shift left", seq: "\x1b[1;6D", want: KeyEvent{Code: KeyLeft, Modifiers: ModCtrl | ModShift}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeySequence([]byte(tt.seq))
			if err != nil {
				t.Fatalf("ParseKeySequence() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseKeySequence() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseKeySequenceRejectsUnsupportedInput(t *testing.T) {
	for _, seq := range []string{"\x1b[1;5D u", "\x1b[99;5~", "\x1b[1;257D"} {
		if _, err := ParseKeySequence([]byte(seq)); err == nil {
			t.Fatalf("ParseKeySequence(%q) error = nil, want unsupported sequence error", seq)
		}
	}
}

func TestParseKeySequenceNormalizesNavigationAndFunctionKeys(t *testing.T) {
	tests := []struct {
		name string
		seq  string
		want KeyEvent
	}{
		{name: "home", seq: "\x1b[1;1H", want: KeyEvent{Code: KeyHome}},
		{name: "end", seq: "\x1b[1;5F", want: KeyEvent{Code: KeyEnd, Modifiers: ModCtrl}},
		{name: "page up", seq: "\x1b[5;2~", want: KeyEvent{Code: KeyPageUp, Modifiers: ModShift}},
		{name: "delete", seq: "\x1b[3~", want: KeyEvent{Code: KeyDelete}},
		{name: "f1 ss3", seq: "\x1bOP", want: KeyEvent{Code: KeyFunction, Function: 1}},
		{name: "f4 ss3", seq: "\x1bOS", want: KeyEvent{Code: KeyFunction, Function: 4}},
		{name: "f5 csi", seq: "\x1b[15~", want: KeyEvent{Code: KeyFunction, Function: 5}},
		{name: "f12 csi", seq: "\x1b[24~", want: KeyEvent{Code: KeyFunction, Function: 12}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeySequence([]byte(tt.seq))
			if err != nil {
				t.Fatalf("ParseKeySequence() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseKeySequence() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseKeySequenceNormalizesKittyControlKeys(t *testing.T) {
	tests := []struct {
		seq  string
		want KeyEvent
	}{
		{seq: "\x1b[13;1u", want: KeyEvent{Code: KeyEnter}},
		{seq: "\x1b[9;1u", want: KeyEvent{Code: KeyTab}},
		{seq: "\x1b[127;1u", want: KeyEvent{Code: KeyBackspace}},
		{seq: "\x1b[32;5u", want: KeyEvent{Code: KeyRune, Rune: ' ', Modifiers: ModCtrl}},
		{seq: "\x1b[96;5u", want: KeyEvent{Code: KeyRune, Rune: '`', Modifiers: ModCtrl}},
		{seq: "\x1b[47;5u", want: KeyEvent{Code: KeyRune, Rune: '/', Modifiers: ModCtrl}},
		{seq: "\x1b[95;5u", want: KeyEvent{Code: KeyRune, Rune: '_', Modifiers: ModCtrl}},
		{seq: "\x1b[63;5u", want: KeyEvent{Code: KeyRune, Rune: '?', Modifiers: ModCtrl}},
	}
	for _, tt := range tests {
		got, err := ParseKeySequence([]byte(tt.seq))
		if err != nil {
			t.Fatalf("ParseKeySequence(%q) error = %v", tt.seq, err)
		}
		if got != tt.want {
			t.Fatalf("ParseKeySequence(%q) = %+v, want %+v", tt.seq, got, tt.want)
		}
	}
}

func TestParseKeySequenceNormalizesXtermModifyOtherKeys(t *testing.T) {
	got, err := ParseKeySequence([]byte("\x1b[27;5;65~"))
	if err != nil {
		t.Fatalf("ParseKeySequence() error = %v", err)
	}
	want := KeyEvent{Code: KeyRune, Rune: 'A', Modifiers: ModCtrl}
	if got != want {
		t.Fatalf("ParseKeySequence() = %+v, want %+v", got, want)
	}
}

func TestParseKeySequenceNormalizesKittyCSIU(t *testing.T) {
	tests := []struct {
		name string
		seq  string
		want KeyEvent
	}{
		{name: "ctrl shift", seq: "\x1b[105;6u", want: KeyEvent{Code: KeyRune, Rune: 'i', Modifiers: ModCtrl | ModShift}},
		{name: "meta", seq: "\x1b[105;33u", want: KeyEvent{Code: KeyRune, Rune: 'i', Modifiers: ModMeta}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeySequence([]byte(tt.seq))
			if err != nil {
				t.Fatalf("ParseKeySequence() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseKeySequence() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseKeySequenceNormalizesLegacyControlAndAltInput(t *testing.T) {
	tests := []struct {
		name string
		seq  string
		want KeyEvent
	}{
		{name: "ctrl c", seq: "\x03", want: KeyEvent{Code: KeyRune, Rune: 'c', Modifiers: ModCtrl}},
		{name: "ctrl slash", seq: "\x1f", want: KeyEvent{Code: KeyRune, Rune: '/', Modifiers: ModCtrl}},
		{name: "ctrl question", seq: "\x7f", want: KeyEvent{Code: KeyBackspace}},
		{name: "ctrl space", seq: "\x00", want: KeyEvent{Code: KeyRune, Rune: ' ', Modifiers: ModCtrl}},
		{name: "alt x", seq: "\x1bx", want: KeyEvent{Code: KeyRune, Rune: 'x', Modifiers: ModAlt}},
		{name: "alt unicode", seq: "\x1b界", want: KeyEvent{Code: KeyRune, Rune: '界', Modifiers: ModAlt}},
		{name: "tab", seq: "\t", want: KeyEvent{Code: KeyTab}},
		{name: "backspace", seq: "\x7f", want: KeyEvent{Code: KeyBackspace}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeySequence([]byte(tt.seq))
			if err != nil {
				t.Fatalf("ParseKeySequence() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseKeySequence() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
