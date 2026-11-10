# 05: Integrate normalized events with the application event loop

**What to build:** Route the single real terminal input owner into the existing application event loop so Root, editor, plugins, overlays, and force keys keep their current behavior while consuming normalized Input Events.

**Blocked by:** 04: Terminal capability negotiation and restoration.

**Status:** resolved

- [x] Real stdin has one input owner and does not race with a second reader. tcell's `Tty` remains the sole low-level reader; ttt does not read stdin in parallel.
- [x] Canonical keyboard events reach existing shortcut and command dispatch behavior through the event-loop adapter before `Root.HandleEvent`.
- [x] Mouse, paste, and resize events preserve existing application behavior; their event queue adapter is covered by `internal/term` tests.
- [x] tcell remains usable through an adapter for SimScreen, rendering, and migration tests.
- [x] Existing force-key and terminal raw-key behavior remains covered by application-level tests in `tests/e2e/key_interceptor_test.go` and `internal/ui/root_test.go`.

The low-level byte reader/parser is intentionally still tcell-owned because tcell v3 does not expose a supported API for replacing its input loop. The application boundary is protocol-independent after tcell has produced an event; replacing the low-level reader would require a maintained tcell fork and is outside this ticket's safe integration surface.
