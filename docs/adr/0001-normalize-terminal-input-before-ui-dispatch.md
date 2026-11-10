# Normalize terminal input before UI dispatch

**Status:** proposed

ttt will introduce a project-owned input event seam between raw terminal input and application/UI dispatch. `internal/term` will own the terminal session, raw input ownership, protocol negotiation, and restoration of terminal modes; `internal/keyboard` will parse and normalize ANSI/VT100, xterm, Kitty CSI-u, DA1/DA2-related capability responses, and tmux transport wrapping into canonical input events. The UI and command layers will consume canonical events rather than terminal-specific encodings.

## Considered Options

### Continue relying on tcell for all input parsing

Rejected because tcell does not expose enough control for Kitty keyboard negotiation, raw CSI-u handling, DA1/DA2 responses, or tmux wrapping. Protocol-specific compatibility would continue to spread through UI code as special cases.

### Add a raw-input side channel beside tcell

Rejected because two consumers would compete for the same stdin bytes, creating ordering and ownership races around keyboard, mouse, paste, resize, and escape-sequence parsing.

### Replace the real-input path with a project-owned parser

Accepted because one input owner can define deterministic protocol handling and emit a stable event model. tcell remains available for rendering, simulated screens, tests, and a compatibility adapter during migration.

## Consequences

- Editor shortcuts can remain stable across legacy terminals, Kitty-compatible terminals, xterm variants, and tmux transport wrappers.
- Protocol negotiation can fall back to legacy input when capability detection or activation fails.
- Kitty CSI-u can distinguish keys that legacy encodings collapse, while legacy mode retains the real limitation.
- The parser must handle ESC ambiguity with a short timeout and must safely discard unknown control sequences.
- Terminal mode changes must be tracked and restored when ttt exits.
- The project must test both byte-level protocol behavior and real terminal matrices; CI should not depend on a particular terminal application.
