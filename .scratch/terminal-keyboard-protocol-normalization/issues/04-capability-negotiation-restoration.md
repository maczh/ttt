# 04: Terminal capability negotiation and restoration

**What to build:** Detect, select, enable, downgrade, and restore terminal keyboard capabilities during the Terminal Session lifecycle.

**Blocked by:** 03: Stateful raw input stream parser.

**Status:** resolved

- [x] `auto`, `legacy`, `xterm`, and `kitty` selection modes are represented.
- [x] Selection precedence is explicit configuration, reliable environment information, active probing, then legacy fallback.
- [x] DA1/DA2-related capability responses are handled without becoming editor input.
- [x] Kitty keyboard protocol and xterm modifyOtherKeys activation failures fall back safely.
- [x] Every mode enabled by ttt is restored on normal and error-driven shutdown.
