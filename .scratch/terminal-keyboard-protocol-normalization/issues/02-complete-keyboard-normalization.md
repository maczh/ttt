# 02: Complete keyboard protocol normalization

**What to build:** Normalize the complete shortcut contract across Legacy Encoding, xterm variants, and Kitty keyboard input so users can use the same ttt shortcuts across terminals.

**Blocked by:** 01: Canonical keyboard events for legacy and Kitty navigation.

**Status:** resolved

- [x] Ctrl, Alt, Shift, and Meta combinations normalize consistently.
- [x] Ctrl+Space and Ctrl+Backtick are distinguished only when the active encoding carries that distinction.
- [x] Ctrl+/, Ctrl+_, Ctrl+?, Tab, Enter, Backspace/DEL, Home/End, page keys, and function keys are covered.
- [x] Modified navigation keys normalize to stable Canonical Key Events.
- [x] Tests use literal protocol inputs and do not depend on parser implementation details.
