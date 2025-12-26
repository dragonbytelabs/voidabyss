package editor

import (
	"fmt"
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

	// Handle :colorscheme [name]
	if cmd == "colorscheme" || (len(cmd) > 12 && cmd[0:12] == "colorscheme ") {
		if cmd == "colorscheme" {
			// Show available schemes
			schemes := ListColorSchemes()
			lines := make([]string, 0, len(schemes))
			for _, name := range schemes {
				marker := " "
				if name == e.config.ColorScheme {
					marker = "*"
				}
				lines = append(lines, fmt.Sprintf("%s %s", marker, name))
			}
			e.popupFixedH = 10
			e.openPopup("COLOR SCHEMES", lines)
			return false
		}
		schemeName := cmd[12:]
		scheme := GetColorScheme(schemeName)
		e.ApplyColorScheme(scheme)
		return false
	}

	// Handle :set filetype=<type>
	if len(cmd) > 13 && cmd[0:13] == "set filetype=" {
		// This is a placeholder - filetype is auto-detected
		e.statusMsg = "filetype is auto-detected from file extension"
		return false
	}

	// Handle :help [topic]
	if cmd == "help" || (len(cmd) > 5 && cmd[0:5] == "help ") {
		topic := ""
		if len(cmd) > 5 {
			topic = cmd[5:]
		}
		e.showHelp(topic)
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
	case "bd", "bdelete":
		e.deleteBuffer()
	case "bd!", "bdelete!":
		e.deleteBufferForce()
	case "ls", "buffers":
		e.listBuffers()
	case "vsplit", "vs":
		e.vsplit()
	case "split", "sp":
		e.split()
	case "close":
		e.closeSplit()
	case "only":
		// Close all splits except current
		e.splits = []*Split{e.splits[e.currentSplit]}
		e.currentSplit = 0
		e.redistributeSplitSpace()
		e.statusMsg = "closed all other splits"
	case "Explore", "Ex":
		e.toggleFileTree()
	case "reg", "registers":
		e.popupFixedH = 10
		e.openPopup("REGISTERS", e.formatRegisters())
		return false
	case "noh", "nohlsearch":
		e.searchMatches = nil
		e.statusMsg = "search highlight cleared"
	case "fold":
		e.ToggleFold()
	case "foldopen", "fo":
		fold := e.findFoldAtLine(e.cy)
		if fold != nil {
			fold.folded = false
			e.statusMsg = "unfolded"
		} else {
			e.statusMsg = "no fold at cursor"
		}
	case "foldclose", "fc":
		fold := e.findFoldAtLine(e.cy)
		if fold != nil {
			fold.folded = true
			e.statusMsg = "folded"
		} else {
			e.statusMsg = "no fold at cursor"
		}
	case "foldall", "fca":
		e.FoldAll()
	case "unfoldall", "ufa":
		e.UnfoldAll()
	case "foldinfo":
		// Debug command to show fold information
		if e.parser == nil {
			e.statusMsg = "no parser available"
		} else if e.foldRanges == nil {
			e.statusMsg = "fold ranges not initialized"
		} else {
			e.statusMsg = fmt.Sprintf("parser: yes, folds: %d", len(e.foldRanges))
		}
		return false
	case "colorschemes":
		// List available color schemes
		schemes := ListColorSchemes()
		lines := make([]string, 0, len(schemes))
		for _, name := range schemes {
			marker := " "
			if name == e.config.ColorScheme {
				marker = "*"
			}
			lines = append(lines, fmt.Sprintf("%s %s", marker, name))
		}
		e.popupFixedH = 10
		e.openPopup("COLOR SCHEMES", lines)
		return false
	case "set":
		// Show current filetype and settings
		ft := e.getFiletype()
		if ft != nil {
			e.statusMsg = fmt.Sprintf("filetype=%s tabwidth=%d colorscheme=%s", ft.Name, e.indentWidth, e.config.ColorScheme)
		} else {
			e.statusMsg = fmt.Sprintf("filetype=none tabwidth=%d colorscheme=%s", e.indentWidth, e.config.ColorScheme)
		}
	default:
		e.statusMsg = "Not a command: " + cmd
	}
	return false
}

func (e *Editor) save() {
	// Fire BufWritePre event
	e.FireBufWritePre()

	outPath := e.filename
	if info, err := os.Stat(outPath); err == nil && info.IsDir() {
		outPath = filepath.Join(outPath, "out.txt")
	}
	_ = os.WriteFile(outPath, []byte(e.buffer.String()), 0644)
	e.dirty = false
	e.statusMsg = "written"

	// Fire BufWritePost event
	e.FireBufWritePost()
}
