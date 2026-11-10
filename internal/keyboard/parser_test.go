package keyboard

import "testing"

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
	for _, seq := range []string{"\x1b[A", "\x1b[1;5D u"} {
		if _, err := ParseKeySequence([]byte(seq)); err == nil {
			t.Fatalf("ParseKeySequence(%q) error = nil, want unsupported sequence error", seq)
		}
	}
}
