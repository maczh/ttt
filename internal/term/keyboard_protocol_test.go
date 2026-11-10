package term

import (
	"errors"
	"reflect"
	"testing"
)

func TestKeyboardProtocolSessionUsesExplicitEnvironmentAndProbeOrder(t *testing.T) {
	var writes [][]byte
	s := NewKeyboardProtocolSession(KeyboardKitty, KeyboardEnvironment{TERM: "xterm-256color"}, func(b []byte) error {
		writes = append(writes, append([]byte(nil), b...))
		return nil
	})
	if err := s.Start(); err != nil || s.Protocol() != KeyboardKitty {
		t.Fatalf("explicit Kitty session = %q, %v", s.Protocol(), err)
	}
	if string(writes[0]) != "\x1b[>1u" {
		t.Fatalf("enable = %q", writes[0])
	}

	writes = nil
	s = NewKeyboardProtocolSession(KeyboardAuto, KeyboardEnvironment{}, func(b []byte) error {
		writes = append(writes, append([]byte(nil), b...))
		return nil
	})
	if err := s.Start(); err != nil || string(writes[0]) != "\x1b[?u\x1b[c" {
		t.Fatalf("probe = %q, %v", writes, err)
	}
	if caps, ok := s.HandleResponse([]byte("\x1b[?31u")); !ok || !caps.KittyKeyboard || s.Protocol() != KeyboardKitty {
		t.Fatalf("Kitty response = %#v, %v, protocol %q", caps, ok, s.Protocol())
	}
}

func TestKeyboardProtocolSessionFallsBackAndRestoresIdempotently(t *testing.T) {
	var writes [][]byte
	s := NewKeyboardProtocolSession(KeyboardKitty, KeyboardEnvironment{}, func([]byte) error {
		return errors.New("write failed")
	})
	if err := s.Start(); err != nil || s.Protocol() != KeyboardLegacy {
		t.Fatalf("failed activation = %q, %v", s.Protocol(), err)
	}

	writes = nil
	s = NewKeyboardProtocolSession(KeyboardKitty, KeyboardEnvironment{}, func(b []byte) error {
		writes = append(writes, append([]byte(nil), b...))
		return nil
	})
	_ = s.Start()
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil || len(writes) != 2 || string(writes[1]) != "\x1b[<u" {
		t.Fatalf("close writes = %#v, err = %v", writes, err)
	}
}

func TestKeyboardProtocolSessionFiltersCapabilityResponses(t *testing.T) {
	s := NewKeyboardProtocolSession(KeyboardLegacy, KeyboardEnvironment{}, nil)
	tests := []struct {
		seq  string
		want KeyboardCapabilities
	}{
		{seq: "\x1b[?31u", want: KeyboardCapabilities{KittyKeyboard: true, KittyFlags: 31}},
		{seq: "\x1b[?1;2c", want: KeyboardCapabilities{DeviceAttributes: true}},
		{seq: "\x1b[>1;4000;123c", want: KeyboardCapabilities{DeviceAttributes: true}},
	}
	for _, tt := range tests {
		got, ok := s.HandleResponse([]byte(tt.seq))
		if !ok || !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("response %q = %#v, %v; want %#v, true", tt.seq, got, ok, tt.want)
		}
	}
	if _, ok := s.HandleResponse([]byte("\x1b[1;5D")); ok {
		t.Fatal("keyboard sequence classified as capability response")
	}
}

func TestKeyboardProtocolSessionProbeTimeoutFallsBack(t *testing.T) {
	s := NewKeyboardProtocolSession(KeyboardAuto, KeyboardEnvironment{}, nil)
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	s.ProbeTimeout()
	if s.Protocol() != KeyboardLegacy {
		t.Fatalf("timeout protocol = %q, want legacy", s.Protocol())
	}
}

func TestKeyboardProtocolSessionUsesLegacyInMultiplexerEnvironments(t *testing.T) {
	for _, env := range []KeyboardEnvironment{
		{TERM: "xterm-256color", TMUX: "/tmp/tmux-1/default"},
		{TERM: "screen", TermProgram: "screen"},
	} {
		s := NewKeyboardProtocolSession(KeyboardAuto, env, nil)
		if err := s.Start(); err != nil {
			t.Fatal(err)
		}
		if got := s.Protocol(); got != KeyboardLegacy {
			t.Fatalf("environment %+v selected %q, want legacy", env, got)
		}
	}
}
