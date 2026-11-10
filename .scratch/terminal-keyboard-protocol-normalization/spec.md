# Normalize terminal keyboard protocols

Status: ready-for-agent

## Problem Statement

Users expect the same ttt shortcut to work in different terminal environments, but terminal applications express the same physical key through different Terminal Input Protocols. Legacy Encoding, xterm variants, Kitty CSI-u, and Transport Wrappers such as tmux can produce different byte sequences or collapse distinct keys into the same sequence.

ttt currently receives most input through tcell events and contains compatibility logic close to UI key matching and terminal forwarding. This makes protocol behavior difficult to extend consistently. It also prevents ttt from reliably negotiating Kitty keyboard capabilities, handling DA1/DA2-related responses, or distinguishing protocol parsing from terminal-session lifecycle management.

## Solution

Introduce a project-owned input pipeline. `internal/term` owns the Terminal Session, raw input ownership, Protocol Negotiation, terminal mode restoration, and event delivery. `internal/keyboard` owns parsing and normalization of Terminal Input Protocols into canonical Input Events. The application and presentation layers consume canonical events instead of terminal-specific encodings.

The real terminal input path becomes a single-owner parser for keyboard, mouse, paste, and resize input. tcell remains available for rendering, simulated screens, tests, and a compatibility adapter during migration, but it is not a second consumer of the real stdin stream.

## User Stories

1. As a ttt user in a legacy VT/ANSI terminal, I want standard shortcuts to keep working, so that upgrading ttt does not regress my workflow.
2. As a Kitty user, I want Kitty keyboard protocol sequences to map to the same shortcuts as equivalent legacy sequences, so that my terminal choice does not change ttt behavior.
3. As a Ghostty or WezTerm user, I want supported keyboard capabilities to be recognized consistently, so that enhanced key combinations work without per-terminal exceptions.
4. As a user inside tmux, I want wrapped input sequences to be unwrapped and interpreted as the same Canonical Key Events, so that tmux does not require a separate shortcut vocabulary.
5. As an SSH user, I want protocol negotiation to degrade safely when the remote terminal or transport filters responses, so that ttt remains usable over imperfect connections.
6. As a user, I want Ctrl+letter shortcuts to remain stable across input encodings, so that common commands do not depend on terminal implementation details.
7. As a user, I want Ctrl+Space and Ctrl+Backtick to be distinguished when the Terminal Input Protocol supports it, so that I can assign separate shortcuts where the terminal provides enough information.
8. As a user on a legacy encoding, I want Ctrl+Space and Ctrl+Backtick to retain their real limitation when both produce NUL, so that ttt does not invent an incorrect distinction.
9. As a user, I want Ctrl+/, Ctrl+_, and Ctrl+? representations to normalize consistently, so that punctuation shortcuts work across keyboard layouts and terminal encodings.
10. As a user, I want modified arrows, Home/End, page keys, function keys, Tab, Enter, and Backspace to normalize consistently, so that navigation shortcuts remain portable.
11. As a user, I want Alt and Meta combinations to preserve their intended modifiers, so that terminal-specific ESC-prefixed forms do not become ordinary text.
12. As a user, I want Shift+Enter and other enhanced combinations to work when the active protocol can represent them, so that modern terminal features are usable without breaking fallback behavior.
13. As a user, I want unsupported protocol negotiation to fall back automatically, so that ttt does not become unusable because a terminal ignored a capability query.
14. As a user, I want ttt to restore terminal modes on exit, so that Kitty keyboard mode, modifyOtherKeys, mouse reporting, paste mode, and cursor state do not leak into my shell.
15. As a user, I want unknown control sequences to be ignored safely, so that terminal metadata or unsupported extensions cannot insert unwanted text into the editor.
16. As a user, I want a lone Escape key to remain usable, so that ESC-prefixed protocol sequences do not make Escape feel stuck or delayed beyond an acceptable short timeout.
17. As a user, I want paste events to remain intact while raw input is parsed, so that bracketed paste is not confused with ordinary keyboard input.
18. As a user, I want mouse events to remain intact while raw input is parsed, so that editor selection and panel interaction continue to work.
19. As a user, I want resize events to remain reliable, so that changing the terminal size still updates layout and integrated terminals.
20. As a user, I want the integrated terminal to receive correctly encoded key sequences, so that shell programs and nested TUIs continue to behave normally.
21. As a user, I want editor shortcut configuration and integrated-terminal forwarding configuration to be independent, so that changing one does not unexpectedly alter the other.
22. As a maintainer, I want protocol-specific parsing centralized, so that adding a new terminal protocol does not require edits across widgets and command handlers.
23. As a maintainer, I want UI code to consume Canonical Key Events, so that shortcut matching can be tested without a real terminal application.
24. As a maintainer, I want protocol capability state to be explicit, so that fallback behavior and terminal diagnostics are explainable.
25. As a maintainer, I want the parser to have deterministic timeout control in tests, so that ESC/CSI ambiguity tests do not rely on sleeps.
26. As a maintainer, I want byte-level parser tests in CI, so that protocol regressions are caught without requiring Kitty, Ghostty, WezTerm, or tmux to be installed.
27. As a maintainer, I want a real-terminal test matrix outside ordinary CI, so that implementation behavior is validated against actual terminal differences.
28. As a plugin author, I want key events exposed after normalization, so that plugin key handling does not need to understand terminal-specific byte sequences.
29. As a user, I want `auto` protocol selection to prefer explicit configuration, then reliable environment information, then active probing, and finally legacy fallback, so that predictable overrides win without sacrificing compatibility.
30. As a user, I want to force `legacy`, `xterm`, or `kitty` behavior separately for editor input and integrated-terminal forwarding, so that unusual terminal setups remain recoverable.

## Implementation Decisions

- Add a project-owned input event seam from Raw Input Stream to Input Event.
- Keep `internal/term` responsible for terminal-session lifecycle, raw input ownership, Protocol Negotiation, terminal mode restoration, and delivery into the application event loop.
- Add or evolve `internal/keyboard` as the deep module responsible for keyboard protocol parsing, normalization, modifier handling, ESC disambiguation, and capability-aware interpretation.
- Represent keyboard, mouse, paste, and resize as distinct normalized Input Event categories. Do not compress them into one opaque event model.
- The real stdin stream has one owner. Do not add a raw-input side channel beside tcell.
- Keep tcell for rendering, `SimScreen`, tests, and a `TcellAdapter` compatibility path during migration.
- Parse legacy VT/ANSI, VT100, xterm, Kitty CSI-u, xterm modifyOtherKeys, DA1/DA2-related capability responses, and tmux Transport Wrapper sequences through the project-owned pipeline.
- Treat tmux as a Transport Wrapper, not as a separate keyboard semantic protocol. Unwrap it before passing the contained sequence to the common parser.
- Use a short parser timeout to distinguish a lone Escape from an ESC-prefixed sequence. The timeout must be injectable or clock-controlled in tests.
- Drop unknown or malformed control sequences safely at a recoverable byte boundary and emit debug diagnostics rather than injecting them as editor text.
- Track every terminal mode enabled by ttt and restore it on shutdown. Do not use a broad terminal reset as a substitute for state-aware restoration.
- In `auto` mode, choose capabilities in this order: explicit configuration, reliable environment information, active probing, then legacy fallback.
- Provide separate configuration for ttt shortcut input and integrated-terminal forwarding, conceptually `keyboard.protocol` and `terminal.keyboardProtocol`.
- Use `auto`, `legacy`, `xterm`, and `kitty` as the supported protocol-selection concepts, with legacy fallback when automatic negotiation fails.
- Preserve the semantic limitation of Legacy Encoding: Ctrl+Space and Ctrl+Backtick are distinct only when the active protocol carries enough information.
- Keep shortcut matching, command dispatch, and widget logic independent of terminal brands and raw escape sequences.
- Keep protocol negotiation in the terminal-session layer rather than in the keyboard semantic layer.
- Preserve the existing main-thread rule: normalized events may be produced by input machinery, but widget state changes occur through the application event loop.

## Testing Decisions

- The primary test seam is external behavior at Raw Input Stream → Input Event: given protocol bytes and capability state, assert the normalized event stream.
- Test the parser without terminal applications, using deterministic input bytes and an injectable clock for ESC ambiguity.
- Cover legacy control bytes, ANSI/VT100 CSI and SS3 sequences, xterm modified keys, Kitty CSI-u, DA1/DA2 responses, tmux wrapping, unknown sequences, malformed sequences, and fallback behavior.
- Cover the complete shortcut contract: Ctrl+letters, Ctrl+Space, Ctrl+Backtick, Ctrl+/, Alt combinations, Shift+Tab, modified arrows, modified navigation keys, Shift+Enter, and Backspace/DEL.
- Assert that Ctrl+Space and Ctrl+Backtick remain merged in Legacy Encoding and are distinct only in a protocol that represents the distinction.
- Test mouse, bracketed paste, and resize parsing as separate Input Event categories so the raw-input owner cannot regress non-keyboard behavior.
- Test Protocol Negotiation as an external state transition: capability query, response, activation, timeout fallback, and shutdown restoration.
- Keep `TcellAdapter` tests for existing simulated-screen and migration behavior, but do not make tcell internals the primary contract.
- Add application-level tests proving canonical events still trigger existing Root, editor, command, and force-key behavior.
- Add integrated-terminal tests proving terminal forwarding uses the configured terminal keyboard protocol independently of editor shortcut normalization.
- Use the existing E2E harness and `term.SimScreen` for composed application behavior where no real terminal is required.
- Use the existing integration/PTY test style for terminal mode lifecycle, raw input ownership, and real process boundaries.
- Add optional real-terminal matrix checks for legacy xterm-style terminals, Kitty, Ghostty, WezTerm, tmux, and SSH-like transport conditions; these are not required for ordinary CI.
- Verify no terminal mode remains enabled after normal exit and error-driven shutdown.

## Out of Scope

- Replacing the terminal emulator used to render integrated PTY output.
- Implementing a general-purpose terminal emulator or a full parser for every possible terminal output control sequence.
- Guaranteeing distinctions that the active terminal protocol does not encode.
- Treating tmux as an independent keyboard semantic protocol.
- Requiring all real terminal applications in CI.
- Changing the user-facing default keybindings solely because the new parser exists.
- Redesigning plugin command semantics beyond passing normalized key events through the existing event model.
- Supporting arbitrary terminal graphics protocols as part of keyboard normalization.

## Further Notes

The existing root and editor keyboard changes in the worktree are unrelated user changes and must be preserved. The implementation should first characterize current shortcut behavior before replacing the real-input path. The existing ADR on terminal input normalization is the governing architectural decision for this work.
