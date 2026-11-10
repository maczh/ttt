package keyboard

import (
	"bytes"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const escapeTimeout = 50 * time.Millisecond

type InputEvent interface{ inputEvent() }

type KeyboardEvent struct{ Key KeyEvent }

func (KeyboardEvent) inputEvent() {}

type PasteEvent struct{ Text string }

func (PasteEvent) inputEvent() {}

type MouseEvent struct {
	Button int
	X      int
	Y      int
	Press  bool
}

func (MouseEvent) inputEvent() {}

type ResizeEvent struct {
	Width  int
	Height int
}

func (ResizeEvent) inputEvent() {}

type StreamParser struct {
	buffer       []byte
	paste        []byte
	readingPaste bool
	escapeAt     time.Time
	now          func() time.Time
}

func NewStreamParser() *StreamParser {
	return &StreamParser{now: time.Now}
}

func (p *StreamParser) SetNow(now func() time.Time) {
	if now == nil {
		p.now = time.Now
		return
	}
	p.now = now
}

func (p *StreamParser) Feed(data []byte) []InputEvent {
	var events []InputEvent
	for _, b := range data {
		if p.readingPaste {
			p.paste = append(p.paste, b)
			const end = "\x1b[201~"
			if bytes.HasSuffix(p.paste, []byte(end)) {
				p.paste = p.paste[:len(p.paste)-len(end)]
				events = append(events, PasteEvent{Text: string(p.paste)})
				p.paste = nil
				p.readingPaste = false
			}
			continue
		}

		if len(p.buffer) == 0 {
			if b == '\x1b' {
				p.buffer = append(p.buffer, b)
				p.escapeAt = p.now()
				continue
			}
			if b == '\x00' || b == '\t' || b == '\r' || b == '\b' || b == '\x7f' || b == 0x1f || (b >= 1 && b <= 26) || b >= 0x20 {
				if event, err := ParseKeySequence([]byte{b}); err == nil {
					events = append(events, KeyboardEvent{Key: event})
				}
			}
			continue
		}

		p.buffer = append(p.buffer, b)
		if string(p.buffer) == "\x1b[200~" {
			p.buffer = nil
			p.readingPaste = true
			continue
		}
		if !sequenceComplete(p.buffer) {
			continue
		}
		if mouse, ok := parseSGRMouse(p.buffer); ok {
			events = append(events, mouse)
			p.buffer = nil
			continue
		}
		if event, err := ParseKeySequence(p.buffer); err == nil {
			events = append(events, KeyboardEvent{Key: event})
		}
		p.buffer = nil
	}
	return events
}

func (p *StreamParser) Flush(now time.Time) []InputEvent {
	if len(p.buffer) == 0 || now.Sub(p.escapeAt) < escapeTimeout {
		return nil
	}
	if len(p.buffer) > 1 {
		p.buffer = nil
		return nil
	}
	p.buffer = nil
	return []InputEvent{KeyboardEvent{Key: KeyEvent{Code: KeyEscape}}}
}

func (p *StreamParser) Resize(width, height int) []InputEvent {
	return []InputEvent{ResizeEvent{Width: width, Height: height}}
}

func sequenceComplete(seq []byte) bool {
	if len(seq) < 2 || seq[0] != '\x1b' {
		return false
	}
	if seq[1] == 'O' {
		return len(seq) == 3
	}
	if seq[1] != '[' {
		_, size := utf8.DecodeRune(seq[1:])
		return size == len(seq)-1 && size > 0
	}
	if len(seq) < 3 {
		return false
	}
	last := seq[len(seq)-1]
	return last >= 0x40 && last <= 0x7e
}

func parseSGRMouse(seq []byte) (MouseEvent, bool) {
	s := string(seq)
	if !strings.HasPrefix(s, "\x1b[<") || len(s) < 6 {
		return MouseEvent{}, false
	}
	final := s[len(s)-1]
	if final != 'M' && final != 'm' {
		return MouseEvent{}, false
	}
	parts := strings.Split(s[3:len(s)-1], ";")
	if len(parts) != 3 {
		return MouseEvent{}, false
	}
	button, err1 := strconv.Atoi(parts[0])
	x, err2 := strconv.Atoi(parts[1])
	y, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return MouseEvent{}, false
	}
	return MouseEvent{Button: button, X: x, Y: y, Press: final == 'M'}, true
}
