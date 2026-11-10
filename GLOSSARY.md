# ttt

本词汇表定义 ttt 终端输入与键盘协议领域的共享语言，避免把终端传输、协议能力和用户按键语义混为一谈。

## Terminal Input

**Terminal Input Protocol**:
终端将用户操作表达为输入字节序列的规则集合，例如传统 VT/ANSI、xterm、Kitty keyboard protocol 和 CSI-u。不要将它与终端输出控制序列或 PTY 本身混称为“终端协议”。
_Avoid_: terminal protocol（过于宽泛）、keyboard driver

**Canonical Key Event**:
与具体终端编码无关、描述用户按键语义的统一事件，例如“Ctrl+Left”或“Shift+Enter”。它是 ttt 快捷键匹配所依赖的语义表示。
_Avoid_: raw key、terminal key、tcell key

**Raw Input Stream**:
终端会话实际产生的、尚未按协议解释的输入字节流。
_Avoid_: key event、terminal event

**Legacy Encoding**:
传统终端使用的单字节控制字符、ESC/CSI 和 SS3 序列编码方式；它无法区分某些语义不同但字节相同的按键，例如部分终端中的 Ctrl+Space 与 Ctrl+Backtick。
_Avoid_: old protocol、ANSI protocol（ANSI 只是其中一部分）

**Keyboard Protocol Capability**:
当前终端会话能够可靠支持的键盘协议能力，例如 CSI-u、Kitty keyboard protocol 或 xterm modifyOtherKeys。
_Avoid_: terminal type、terminal brand

**Protocol Negotiation**:
ttt 查询、启用或恢复终端输入协议能力的会话过程，包括能力探测、失败降级和退出时恢复状态。
_Avoid_: protocol detection（只表示识别，不包含启用和恢复）

**Transport Wrapper**:
包裹、转发或改写终端输入序列的中间层，例如 tmux。它改变字节传输方式，但不定义新的用户按键语义。
_Avoid_: tmux keyboard protocol、terminal protocol

## Terminal Events

**Input Event**:
终端输入经过协议解释后交给应用的事件类别，包括键盘、鼠标、粘贴和尺寸变化事件。
_Avoid_: tcell event（具体实现名称）

**Keyboard Event**:
描述一次按键语义的 Input Event，通常包含按键类别、字符或功能键编号以及修饰键。
_Avoid_: key byte、escape sequence

**Terminal Session**:
ttt 与一个真实终端输入输出环境之间的生命周期，包括输入接管、协议状态、PTY 交互和退出恢复。
_Avoid_: terminal window（可能只是宿主程序的窗口）
