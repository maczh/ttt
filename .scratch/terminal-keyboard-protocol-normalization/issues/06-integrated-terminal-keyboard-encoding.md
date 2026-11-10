# 06: Independent integrated-terminal keyboard encoding

**What to build:** Keep ttt shortcut normalization independent from the byte encoding sent to programs inside the integrated terminal, with separate protocol configuration and capability-aware forwarding.

**Blocked by:** 02: Complete keyboard protocol normalization; 05: Integrate normalized events with the application event loop.

**Status:** resolved

- [x] Editor input protocol and integrated-terminal keyboard protocol are configured independently through `terminal.keyboardProtocol`.
- [x] Terminal forwarding selects legacy, xterm-compatible, or Kitty-compatible encoding.
- [x] Shells and nested TUIs receive stable sequences for supported navigation, modifier, and Unicode key combinations.
- [x] Existing terminal copy/paste, raw-key, mouse, and force-key behavior remains intact.
- [x] Encoding tests prove forwarding behavior without coupling it to editor shortcut matching.

The default remains legacy for child PTYs. The outer terminal's capability session is not written to the child shell; nested programs can negotiate their own protocol, while ttt's forwarding mode is explicitly configurable.
