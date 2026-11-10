package keyboard

import (
	"errors"
	"reflect"
	"testing"
)

func TestSelectProtocolUsesExplicitEnvironmentThenLegacy(t *testing.T) {
	tests := []struct {
		name    string
		setting Protocol
		env     Environment
		want    Protocol
	}{
		{name: "explicit legacy wins", setting: ProtocolLegacy, env: Environment{TERM: "xterm-kitty"}, want: ProtocolLegacy},
		{name: "kitty term", setting: ProtocolAuto, env: Environment{TERM: "xterm-kitty"}, want: ProtocolKitty},
		{name: "kitty program", setting: ProtocolAuto, env: Environment{TermProgram: "ghostty"}, want: ProtocolKitty},
		{name: "tmux falls back", setting: ProtocolAuto, env: Environment{TMUX: "1", TERM: "xterm-kitty"}, want: ProtocolLegacy},
		{name: "unknown falls back", setting: ProtocolAuto, env: Environment{TERM: "xterm-256color"}, want: ProtocolLegacy},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectProtocol(tt.setting, tt.env); got != tt.want {
				t.Fatalf("SelectProtocol() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNegotiatorPlansEnableAndRestore(t *testing.T) {
	n := NewNegotiator(ProtocolKitty)
	if got := n.EnableSequence(); string(got) != "\x1b[>1u" {
		t.Fatalf("Kitty enable = %q", got)
	}
	if got := n.Activate(func([]byte) error { return nil }); got != ProtocolKitty {
		t.Fatalf("Kitty activation = %q", got)
	}
	if got := n.RestoreSequence(); string(got) != "\x1b[<u" {
		t.Fatalf("Kitty restore = %q", got)
	}

	n = NewNegotiator(ProtocolXterm)
	if got := n.EnableSequence(); string(got) != "\x1b[>4;1m" {
		t.Fatalf("xterm enable = %q", got)
	}
	if got := n.Activate(func([]byte) error { return nil }); got != ProtocolXterm {
		t.Fatalf("xterm activation = %q", got)
	}
	if got := n.RestoreSequence(); string(got) != "\x1b[>4m" {
		t.Fatalf("xterm restore = %q", got)
	}
}

func TestNegotiatorWriteFailureFallsBackToLegacy(t *testing.T) {
	n := NewNegotiator(ProtocolKitty)
	got := n.Activate(func([]byte) error { return errors.New("unsupported") })
	if got != ProtocolLegacy || n.Protocol() != ProtocolLegacy {
		t.Fatalf("activation protocol = %q, want legacy", got)
	}
	if restore := n.RestoreSequence(); len(restore) != 0 {
		t.Fatalf("fallback restore = %q, want empty", restore)
	}
}

func TestParseCapabilityResponseDoesNotBecomeInput(t *testing.T) {
	tests := []struct {
		seq  string
		want Capabilities
	}{
		{seq: "\x1b[?31u", want: Capabilities{KittyKeyboard: true, KittyFlags: 31}},
		{seq: "\x1b[?1;2c", want: Capabilities{DeviceAttributes: true}},
		{seq: "\x1b[>1;4000;123c", want: Capabilities{DeviceAttributes: true}},
	}
	for _, tt := range tests {
		got, ok := ParseCapabilityResponse([]byte(tt.seq))
		if !ok || !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("ParseCapabilityResponse(%q) = %#v, %v; want %#v, true", tt.seq, got, ok, tt.want)
		}
	}
	if _, ok := ParseCapabilityResponse([]byte("\x1b[1;5D")); ok {
		t.Fatal("keyboard sequence was classified as capability response")
	}
}
