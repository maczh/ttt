package term

import (
	"errors"

	"github.com/eugenioenko/ttt/internal/keyboard"
	"github.com/gdamore/tcell/v3"
)

var errUnsupportedKeyboardEvent = errors.New("unsupported keyboard event")

// TcellKeyToKeyboard crosses the real-terminal adapter boundary. tcell owns
// the single TTY reader; the application consumes the resulting event through
// the protocol-independent keyboard model from this point onward.
func TcellKeyToKeyboard(event *tcell.EventKey) (keyboard.KeyEvent, bool) {
	if event == nil {
		return keyboard.KeyEvent{}, false
	}
	mod := keyboard.Modifiers(0)
	if event.Modifiers()&tcell.ModShift != 0 {
		mod |= keyboard.ModShift
	}
	if event.Modifiers()&tcell.ModAlt != 0 {
		mod |= keyboard.ModAlt
	}
	if event.Modifiers()&tcell.ModCtrl != 0 {
		mod |= keyboard.ModCtrl
	}
	if event.Modifiers()&tcell.ModMeta != 0 {
		mod |= keyboard.ModMeta
	}
	if event.Key() == tcell.KeyRune {
		r := KeyRune(event)
		if r == 0 {
			return keyboard.KeyEvent{}, false
		}
		return keyboard.KeyEvent{Code: keyboard.KeyRune, Rune: r, Modifiers: mod}, true
	}
	for code, key := range map[keyboard.KeyCode]tcell.Key{
		keyboard.KeyEscape:    keyEscape,
		keyboard.KeyEnter:     keyEnter,
		keyboard.KeyTab:       keyTab,
		keyboard.KeyBackspace: keyBackspace,
		keyboard.KeyLeft:      keyLeft,
		keyboard.KeyRight:     keyRight,
		keyboard.KeyUp:        keyUp,
		keyboard.KeyDown:      keyDown,
		keyboard.KeyHome:      keyHome,
		keyboard.KeyEnd:       keyEnd,
		keyboard.KeyPageUp:    keyPageUp,
		keyboard.KeyPageDown:  keyPageDown,
		keyboard.KeyInsert:    keyInsert,
		keyboard.KeyDelete:    keyDelete,
	} {
		if event.Key() == key {
			return keyboard.KeyEvent{Code: code, Modifiers: mod}, true
		}
	}
	if event.Key() >= tcell.KeyF1 && event.Key() <= tcell.KeyF12 {
		return keyboard.KeyEvent{Code: keyboard.KeyFunction, Function: int(event.Key()-tcell.KeyF1) + 1, Modifiers: mod}, true
	}
	return keyboard.KeyEvent{}, false
}

const (
	keyEscape    = tcell.KeyEscape
	keyEnter     = tcell.KeyEnter
	keyTab       = tcell.KeyTab
	keyBackspace = tcell.KeyBackspace
	keyLeft      = tcell.KeyLeft
	keyRight     = tcell.KeyRight
	keyUp        = tcell.KeyUp
	keyDown      = tcell.KeyDown
	keyHome      = tcell.KeyHome
	keyEnd       = tcell.KeyEnd
	keyPageUp    = tcell.KeyPgUp
	keyPageDown  = tcell.KeyPgDn
	keyInsert    = tcell.KeyInsert
	keyDelete    = tcell.KeyDelete
)

// KeyboardEventToTcell adapts a normalized keyboard event to the existing
// tcell event seam used by Root and widgets.
func KeyboardEventToTcell(event keyboard.KeyEvent) (*tcell.EventKey, bool) {
	mod := tcell.ModMask(0)
	if event.Modifiers&keyboard.ModShift != 0 {
		mod |= tcell.ModShift
	}
	if event.Modifiers&keyboard.ModAlt != 0 {
		mod |= tcell.ModAlt
	}
	if event.Modifiers&keyboard.ModCtrl != 0 {
		mod |= tcell.ModCtrl
	}
	if event.Modifiers&keyboard.ModMeta != 0 {
		mod |= tcell.ModMeta
	}

	if event.Code == keyboard.KeyRune {
		return tcell.NewEventKey(tcell.KeyRune, string(event.Rune), mod), true
	}
	key, ok := tcellKey(event)
	if !ok {
		return nil, false
	}
	return tcell.NewEventKey(key, "", mod), true
}

func tcellKey(event keyboard.KeyEvent) (tcell.Key, bool) {
	keys := map[keyboard.KeyCode]tcell.Key{
		keyboard.KeyEscape:    tcell.KeyEscape,
		keyboard.KeyEnter:     tcell.KeyEnter,
		keyboard.KeyTab:       tcell.KeyTab,
		keyboard.KeyBackspace: tcell.KeyBackspace,
		keyboard.KeyLeft:      tcell.KeyLeft,
		keyboard.KeyRight:     tcell.KeyRight,
		keyboard.KeyUp:        tcell.KeyUp,
		keyboard.KeyDown:      tcell.KeyDown,
		keyboard.KeyHome:      tcell.KeyHome,
		keyboard.KeyEnd:       tcell.KeyEnd,
		keyboard.KeyPageUp:    tcell.KeyPgUp,
		keyboard.KeyPageDown:  tcell.KeyPgDn,
		keyboard.KeyInsert:    tcell.KeyInsert,
		keyboard.KeyDelete:    tcell.KeyDelete,
	}
	if key, ok := keys[event.Code]; ok {
		return key, true
	}
	if event.Code == keyboard.KeyFunction && event.Function >= 1 && event.Function <= 12 {
		return tcell.KeyF1 + tcell.Key(event.Function-1), true
	}
	return 0, false
}

func (t *TcellScreen) PostKeyboardEvent(event keyboard.KeyEvent) error {
	key, ok := KeyboardEventToTcell(event)
	if !ok {
		return errUnsupportedKeyboardEvent
	}
	return t.PostEvent(key)
}

func (t *TcellScreen) PostInputEvents(events []keyboard.InputEvent) error {
	for _, event := range events {
		switch value := event.(type) {
		case keyboard.KeyboardEvent:
			if err := t.PostKeyboardEvent(value.Key); err != nil {
				return err
			}
		case keyboard.PasteEvent:
			if err := t.PostEvent(tcell.NewEventPaste(true)); err != nil {
				return err
			}
			for _, r := range value.Text {
				if err := t.PostKeyboardEvent(keyboard.KeyEvent{Code: keyboard.KeyRune, Rune: r}); err != nil {
					return err
				}
			}
			if err := t.PostEvent(tcell.NewEventPaste(false)); err != nil {
				return err
			}
		case keyboard.MouseEvent:
			if err := t.PostEvent(tcell.NewEventMouse(value.X, value.Y, mouseButton(value.Button, value.Press), 0)); err != nil {
				return err
			}
		case keyboard.ResizeEvent:
			if err := t.PostEvent(tcell.NewEventResize(value.Width, value.Height)); err != nil {
				return err
			}
		default:
			return errUnsupportedKeyboardEvent
		}
	}
	return nil
}

func mouseButton(button int, press bool) tcell.ButtonMask {
	if !press {
		return tcell.ButtonNone
	}
	switch button & 3 {
	case 0:
		return tcell.Button1
	case 1:
		return tcell.Button3
	case 2:
		return tcell.Button2
	default:
		return tcell.ButtonNone
	}
}
