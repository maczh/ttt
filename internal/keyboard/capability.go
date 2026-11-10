package keyboard

import (
	"errors"
	"strconv"
	"strings"
)

type Protocol string

const (
	ProtocolAuto   Protocol = "auto"
	ProtocolLegacy Protocol = "legacy"
	ProtocolXterm  Protocol = "xterm"
	ProtocolKitty  Protocol = "kitty"
)

type Environment struct {
	TERM        string
	TermProgram string
	TMUX        string
}

type Capabilities struct {
	KittyKeyboard    bool
	KittyFlags       int
	DeviceAttributes bool
}

func SelectProtocol(setting Protocol, env Environment) Protocol {
	configured := setting
	if configured == ProtocolLegacy || configured == ProtocolXterm || configured == ProtocolKitty {
		return configured
	}
	term := strings.ToLower(env.TERM)
	if env.TMUX != "" || strings.HasPrefix(term, "screen") || strings.HasPrefix(term, "tmux") {
		return ProtocolLegacy
	}
	if strings.Contains(term, "kitty") || strings.EqualFold(env.TermProgram, "kitty") || strings.EqualFold(env.TermProgram, "ghostty") || strings.EqualFold(env.TermProgram, "wezterm") {
		return ProtocolKitty
	}
	return ProtocolLegacy
}

type Negotiator struct {
	protocol Protocol
	active   bool
}

func NewNegotiator(protocol Protocol) *Negotiator {
	return &Negotiator{protocol: protocol}
}

func (n *Negotiator) Protocol() Protocol { return n.protocol }

func (n *Negotiator) EnableSequence() []byte {
	switch n.protocol {
	case ProtocolKitty:
		return []byte("\x1b[>1u")
	case ProtocolXterm:
		return []byte("\x1b[>4;1m")
	default:
		return nil
	}
}

func (n *Negotiator) RestoreSequence() []byte {
	if !n.active {
		return nil
	}
	switch n.protocol {
	case ProtocolKitty:
		return []byte("\x1b[<u")
	case ProtocolXterm:
		return []byte("\x1b[>4m")
	default:
		return nil
	}
}

func (n *Negotiator) Activate(write func([]byte) error) Protocol {
	if n.protocol == ProtocolLegacy {
		return n.protocol
	}
	if err := write(n.EnableSequence()); err != nil {
		n.protocol = ProtocolLegacy
		n.active = false
		return n.protocol
	}
	n.active = true
	return n.protocol
}

func ParseCapabilityResponse(seq []byte) (Capabilities, bool) {
	s := string(seq)
	if strings.HasPrefix(s, "\x1b[?") && strings.HasSuffix(s, "u") {
		flags, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(s, "\x1b[?"), "u"))
		if err != nil {
			return Capabilities{}, false
		}
		return Capabilities{KittyKeyboard: true, KittyFlags: flags}, true
	}
	if strings.HasPrefix(s, "\x1b[?") && strings.HasSuffix(s, "c") {
		return Capabilities{DeviceAttributes: true}, true
	}
	if strings.HasPrefix(s, "\x1b[>") && strings.HasSuffix(s, "c") {
		return Capabilities{DeviceAttributes: true}, true
	}
	return Capabilities{}, false
}

var errNegotiationUnavailable = errors.New("keyboard protocol negotiation unavailable")
