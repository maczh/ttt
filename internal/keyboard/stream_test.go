package keyboard

import (
	"testing"
	"time"
)

func TestStreamParserJoinsSequencesAcrossChunks(t *testing.T) {
	p := NewStreamParser()
	if got := p.Feed([]byte("\x1b[1;5")); len(got) != 0 {
		t.Fatalf("first chunk emitted %v, want no events", got)
	}
	got := p.Feed([]byte("D"))
	if len(got) != 1 {
		t.Fatalf("second chunk emitted %d events, want 1", len(got))
	}
	want := KeyboardEvent{Key: KeyEvent{Code: KeyLeft, Modifiers: ModCtrl}}
	if got[0] != want {
		t.Fatalf("event = %#v, want %#v", got[0], want)
	}
}

func TestStreamParserParsesPasteAndMouseEvents(t *testing.T) {
	p := NewStreamParser()
	got := p.Feed([]byte("\x1b[200~hello\n世界\x1b[201~\x1b[<0;12;7M"))
	if len(got) != 2 {
		t.Fatalf("events = %#v, want paste and mouse", got)
	}
	if got[0] != (PasteEvent{Text: "hello\n世界"}) {
		t.Fatalf("paste event = %#v", got[0])
	}
	wantMouse := MouseEvent{Button: 0, X: 12, Y: 7, Press: true}
	if got[1] != (wantMouse) {
		t.Fatalf("mouse event = %#v, want %#v", got[1], wantMouse)
	}
}

func TestStreamParserFlushesLoneEscapeWithInjectedTime(t *testing.T) {
	p := NewStreamParser()
	now := time.Unix(100, 0)
	p.SetNow(func() time.Time { return now })
	if got := p.Feed([]byte{'\x1b'}); len(got) != 0 {
		t.Fatalf("escape emitted early: %#v", got)
	}
	if got := p.Flush(now.Add(escapeTimeout - time.Nanosecond)); len(got) != 0 {
		t.Fatalf("escape emitted before timeout: %#v", got)
	}
	got := p.Flush(now.Add(escapeTimeout))
	if len(got) != 1 || got[0] != (KeyboardEvent{Key: KeyEvent{Code: KeyEscape}}) {
		t.Fatalf("flush = %#v, want escape event", got)
	}
}

func TestStreamParserDropsIncompleteSequenceAfterTimeout(t *testing.T) {
	p := NewStreamParser()
	now := time.Unix(100, 0)
	p.SetNow(func() time.Time { return now })
	p.Feed([]byte("\x1b["))
	if got := p.Flush(now.Add(escapeTimeout)); len(got) != 0 {
		t.Fatalf("flush = %#v, want no event", got)
	}
}

func TestStreamParserDropsUnknownSequenceAndEmitsFollowingRune(t *testing.T) {
	p := NewStreamParser()
	got := p.Feed([]byte("\x1b[999zX"))
	if len(got) != 1 || got[0] != (KeyboardEvent{Key: KeyEvent{Code: KeyRune, Rune: 'X'}}) {
		t.Fatalf("events = %#v, want only X", got)
	}
}

func TestStreamParserResizeEvent(t *testing.T) {
	p := NewStreamParser()
	got := p.Resize(120, 40)
	if len(got) != 1 || got[0] != (ResizeEvent{Width: 120, Height: 40}) {
		t.Fatalf("resize events = %#v", got)
	}
}
