package editor

import (
	"os"
	"path/filepath"
)

func (e *Editor) exec(cmd string) bool {
	// Handle :e filename
	if len(cmd) > 2 && cmd[0:2] == "e " {
		filename := cmd[2:]
		e.openFile(filename)
		return false
	}

	switch cmd {
	case "q":
		if e.dirty {
			e.statusMsg = "No write since last change"
			return false
		}
		return true
	case "q!":
		return true
	case "w":
		e.save()
	case "wq":
		e.save()
		return true
	case "bn", "bnext":
		e.nextBuffer()
	case "bp", "bprev", "bprevious":
		e.prevBuffer()
	case "ls", "buffers":
		e.listBuffers()
	case "reg", "registers":
		e.popupFixedH = 10
		e.openPopup("REGISTERS", e.formatRegisters())
		return false
	case "noh", "nohlsearch":
		e.searchMatches = nil
		e.statusMsg = "search highlight cleared"
	default:
		e.statusMsg = "Not a command: " + cmd
	}
	return false
}

func (e *Editor) save() {
	outPath := e.filename
	if info, err := os.Stat(outPath); err == nil && info.IsDir() {
		outPath = filepath.Join(outPath, "out.txt")
	}
	_ = os.WriteFile(outPath, []byte(e.buffer.String()), 0644)
	e.dirty = false
	e.statusMsg = "written"
}
