package editor

import (
	"os"
	"path/filepath"
)

func (e *Editor) exec(cmd string) bool {
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
	case "reg", "registers":
		e.popupFixedH = 10
		e.openPopup("REGISTERS", e.formatRegisters())
		return false
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
