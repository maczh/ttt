package term

import (
	"github.com/eugenioenko/ttt/internal/keyboard"
	"github.com/gdamore/tcell/v3"
	"testing"
)

func TestKeyboardEventToTcellPreservesShortcutSemantics(t *testing.T) {
	got, ok := KeyboardEventToTcell(keyboard.KeyEvent{Code: keyboard.KeyLeft, Modifiers: keyboard.ModCtrl | keyboard.ModShift})
	if !ok || got.Key() != tcell.KeyLeft || got.Modifiers() != tcell.ModCtrl|tcell.ModShift {
		t.Fatalf("event = %#v, ok=%v", got, ok)
	}
}

func TestTcellKeyToKeyboardRoundTripsCanonicalShortcut(t *testing.T) {
	got, ok := TcellKeyToKeyboard(tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModCtrl|tcell.ModShift))
	if !ok {
		t.Fatal("TcellKeyToKeyboard rejected a supported key")
	}
	want := keyboard.KeyEvent{Code: keyboard.KeyLeft, Modifiers: keyboard.ModCtrl | keyboard.ModShift}
	if got != want {
		t.Fatalf("normalized key = %#v, want %#v", got, want)
	}
}

func TestTcellKeyToKeyboardPreservesUnicodeRune(t *testing.T) {
	got, ok := TcellKeyToKeyboard(tcell.NewEventKey(tcell.KeyRune, "界", tcell.ModAlt))
	if !ok || got.Code != keyboard.KeyRune || got.Rune != '界' || got.Modifiers != keyboard.ModAlt {
		t.Fatalf("normalized unicode key = %#v, ok=%v", got, ok)
	}
}

func TestKeyboardEventToTcellPreservesRuneAndFunction(t *testing.T) {
	runeEvent, ok := KeyboardEventToTcell(keyboard.KeyEvent{Code: keyboard.KeyRune, Rune: '界', Modifiers: keyboard.ModAlt})
	if !ok || runeEvent.Key() != tcell.KeyRune || runeEvent.Str() != "界" || runeEvent.Modifiers() != tcell.ModAlt {
		t.Fatalf("rune event = %#v, ok=%v", runeEvent, ok)
	}
	functionEvent, ok := KeyboardEventToTcell(keyboard.KeyEvent{Code: keyboard.KeyFunction, Function: 12})
	if !ok || functionEvent.Key() != tcell.KeyF12 {
		t.Fatalf("function event = %#v, ok=%v", functionEvent, ok)
	}
}

func TestKeyboardEventToTcellRejectsUnknownEvents(t *testing.T) {
	if _, ok := KeyboardEventToTcell(keyboard.KeyEvent{Code: keyboard.KeyUnknown}); ok {
		t.Fatal("unknown event was accepted")
	}
}

func TestTcellScreenPostsNormalizedInputEventCategories(t *testing.T) {
	screen := NewTcellScreenFrom(NewSimScreen())
	err := screen.PostInputEvents([]keyboard.InputEvent{
		keyboard.PasteEvent{Text: "x"},
		keyboard.MouseEvent{Button: 0, X: 2, Y: 3, Press: true},
		keyboard.ResizeEvent{Width: 80, Height: 24},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := screen.PollEvent().(*tcell.EventPaste); !ok {
		t.Fatal("first event was not paste start")
	}
	if event, ok := screen.PollEvent().(*tcell.EventKey); !ok || event.Str() != "x" {
		t.Fatalf("paste key = %#v, ok=%v", event, ok)
	}
	if _, ok := screen.PollEvent().(*tcell.EventPaste); !ok {
		t.Fatal("third event was not paste end")
	}
	event, ok := screen.PollEvent().(*tcell.EventMouse)
	if !ok {
		t.Fatalf("mouse event = %#v, ok=%v", event, ok)
	}
	x, y := event.Position()
	if x != 2 || y != 3 {
		t.Fatalf("mouse position = (%d, %d)", x, y)
	}
	resize, ok := screen.PollEvent().(*tcell.EventResize)
	if !ok {
		t.Fatalf("resize event = %#v, ok=%v", resize, ok)
	}
	w, h := resize.Size()
	if w != 80 || h != 24 {
		t.Fatalf("resize size = (%d, %d)", w, h)
	}
}
