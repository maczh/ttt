# 01: Canonical keyboard events for legacy and Kitty navigation

**What to build:** Make basic navigation shortcuts from legacy CSI and Kitty keyboard encodings produce the same Canonical Key Event, providing the first usable protocol-normalization path.

**Blocked by:** None (can start immediately).

**Status:** resolved

- [x] Legacy CSI and Kitty navigation encodings normalize to the same event.
- [x] Ctrl/Alt/Shift modifier values are represented independently of terminal-specific bytes.
- [x] Unsupported basic sequences fail safely without producing a misleading event.
- [x] Tests cover at least modified Left and Right navigation and preserve the existing partial implementation behavior.
