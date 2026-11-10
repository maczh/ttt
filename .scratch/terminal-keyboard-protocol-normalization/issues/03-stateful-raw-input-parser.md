# 03: Stateful raw input stream parser

**What to build:** Parse a real Raw Input Stream across read boundaries into keyboard, mouse, paste, and resize Input Events, while safely handling escape ambiguity and unknown control sequences.

**Blocked by:** 02: Complete keyboard protocol normalization.

**Status:** resolved

- [x] Sequences split across reads are parsed as one event.
- [x] A lone Escape is emitted after a short deterministic timeout.
- [x] CSI, SS3, bracketed paste, mouse, and resize inputs produce their corresponding event categories.
- [x] Unknown and malformed control sequences are discarded at recoverable boundaries and never injected as editor text.
- [x] Tests use an injectable or deterministic clock and do not rely on sleeps.
