# 07: Cross-terminal regression and compatibility matrix

**What to build:** Verify that normalized shortcuts and integrated-terminal forwarding remain compatible across the supported terminal matrix, while keeping ordinary CI independent of installed terminal applications.

**Blocked by:** 05: Integrate normalized events with the application event loop; 06: Independent integrated-terminal keyboard encoding.

**Status:** resolved

- [x] Byte-level protocol tests run in ordinary CI for legacy control bytes, xterm modifyOtherKeys, and Kitty CSI-u.
- [x] Composed E2E tests cover shortcut dispatch through the normalized event seam; force-key and raw-terminal paths remain covered by the existing E2E suite.
- [x] PTY-independent protocol-session tests cover capability probing, multiplexer fallback, activation failure, timeout, and restoration.
- [ ] Optional real-terminal checks cover legacy xterm-style terminals, Kitty, Ghostty, WezTerm, tmux, and SSH-like transport conditions. These remain an environment-dependent manual/CI job and are intentionally not required by ordinary unit CI.
- [x] Protocol tests report the input encoding and environment context in each table case, rather than only reporting a generic key mismatch.
