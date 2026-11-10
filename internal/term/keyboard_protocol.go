package term

import (
	"strconv"
	"strings"
)

type KeyboardProtocol string

const (
	KeyboardAuto   KeyboardProtocol = "auto"
	KeyboardLegacy KeyboardProtocol = "legacy"
	KeyboardXterm  KeyboardProtocol = "xterm"
	KeyboardKitty  KeyboardProtocol = "kitty"
)

type KeyboardEnvironment struct {
	TERM        string
	TermProgram string
	TMUX        string
}

type KeyboardCapabilities struct {
	KittyKeyboard    bool
	KittyFlags       int
	DeviceAttributes bool
}

type KeyboardProtocolSession struct {
	protocol KeyboardProtocol
	enabled  bool
	probing  bool
	closed   bool
	write    func([]byte) error
}

func NewKeyboardProtocolSession(requested KeyboardProtocol, env KeyboardEnvironment, write func([]byte) error) *KeyboardProtocolSession {
	protocol := requested
	if protocol == KeyboardAuto {
		protocol = protocolFromEnvironment(env)
		if protocol == KeyboardAuto {
			return &KeyboardProtocolSession{protocol: KeyboardAuto, probing: true, write: write}
		}
	}
	return &KeyboardProtocolSession{protocol: protocol, write: write}
}

func protocolFromEnvironment(env KeyboardEnvironment) KeyboardProtocol {
	termName := strings.ToLower(env.TERM)
	if env.TMUX != "" || strings.HasPrefix(termName, "screen") || strings.HasPrefix(termName, "tmux") {
		return KeyboardLegacy
	}
	if strings.Contains(termName, "kitty") || strings.EqualFold(env.TermProgram, "kitty") || strings.EqualFold(env.TermProgram, "ghostty") || strings.EqualFold(env.TermProgram, "wezterm") {
		return KeyboardKitty
	}
	return KeyboardAuto
}

func (s *KeyboardProtocolSession) Protocol() KeyboardProtocol { return s.protocol }

func (s *KeyboardProtocolSession) Start() error {
	if s.probing {
		return s.writeBytes([]byte("\x1b[?u\x1b[c"))
	}
	return s.enable()
}

func (s *KeyboardProtocolSession) enable() error {
	sequence := s.enableSequence()
	if len(sequence) == 0 {
		return nil
	}
	if err := s.writeBytes(sequence); err != nil {
		s.protocol = KeyboardLegacy
		s.enabled = false
		return nil
	}
	s.enabled = true
	return nil
}

func (s *KeyboardProtocolSession) HandleResponse(response []byte) (KeyboardCapabilities, bool) {
	caps, ok := parseKeyboardCapabilities(response)
	if !ok {
		return KeyboardCapabilities{}, false
	}
	if s.probing && caps.KittyKeyboard {
		s.protocol = KeyboardKitty
		s.probing = false
		_ = s.enable()
	}
	return caps, true
}

func (s *KeyboardProtocolSession) ProbeTimeout() {
	if s.probing {
		s.probing = false
		s.protocol = KeyboardLegacy
	}
}

func (s *KeyboardProtocolSession) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	if !s.enabled {
		return nil
	}
	return s.writeBytes(s.restoreSequence())
}

func (s *KeyboardProtocolSession) writeBytes(sequence []byte) error {
	if len(sequence) == 0 || s.write == nil {
		return nil
	}
	return s.write(sequence)
}

func (s *KeyboardProtocolSession) enableSequence() []byte {
	switch s.protocol {
	case KeyboardKitty:
		return []byte("\x1b[>1u")
	case KeyboardXterm:
		return []byte("\x1b[>4;1m")
	default:
		return nil
	}
}

func (s *KeyboardProtocolSession) restoreSequence() []byte {
	switch s.protocol {
	case KeyboardKitty:
		return []byte("\x1b[<u")
	case KeyboardXterm:
		return []byte("\x1b[>4m")
	default:
		return nil
	}
}

func parseKeyboardCapabilities(response []byte) (KeyboardCapabilities, bool) {
	s := string(response)
	if strings.HasPrefix(s, "\x1b[?") && strings.HasSuffix(s, "u") {
		flags, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(s, "\x1b[?"), "u"))
		if err != nil {
			return KeyboardCapabilities{}, false
		}
		return KeyboardCapabilities{KittyKeyboard: true, KittyFlags: flags}, true
	}
	if (strings.HasPrefix(s, "\x1b[?") || strings.HasPrefix(s, "\x1b[>")) && strings.HasSuffix(s, "c") {
		return KeyboardCapabilities{DeviceAttributes: true}, true
	}
	return KeyboardCapabilities{}, false
}
