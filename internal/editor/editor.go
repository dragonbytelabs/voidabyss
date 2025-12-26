package editor

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dragonbytelabs/voidabyss/core/buffer"
	"github.com/dragonbytelabs/voidabyss/internal/config"
	"github.com/gdamore/tcell/v2"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
	ModeVisual
	ModeSearch
)

// string returns the string representation of a Mode
func (m Mode) string() string {
	switch m {
	case ModeNormal:
		return "normal"
	case ModeInsert:
		return "insert"
	case ModeCommand:
		return "command"
	case ModeVisual:
		return "visual"
	case ModeSearch:
		return "search"
	default:
		return "unknown"
	}
}

type RegisterKind int

const (
	RegCharwise RegisterKind = iota
	RegLinewise
	RegBlockwise
)

type Register struct {
	kind RegisterKind
	text string
}

type Registers struct {
	unnamed  Register          // "
	numbered [10]Register      // 0-9
	named    map[rune]Register // a-z
	small    Register          // "- small delete
}

type VisualKind int

const (
	VisualChar VisualKind = iota
	VisualLine
)

type RepeatKind int

const (
	RepeatNone RepeatKind = iota
	RepeatOpMotion
	RepeatPasteAfter
	RepeatPasteBefore
	RepeatDeleteChar // x
	RepeatInsert     // i, a, A, o, O with text typed
)

type RepeatAction struct {
	kind   RepeatKind
	op     rune
	motion rune
	count  int
	reg    rune
	// text object (iw/aw/iW/aW)
	textObjPrefix rune // 'i' or 'a'
	textObjUnit   rune // 'w' or 'W'
	// insert mode tracking
	insertCmd  rune   // 'i', 'a', 'A', 'o', 'O'
	insertText []rune // text typed during insert mode
}

// Mark represents a position in the buffer
type Mark struct {
	line int
	col  int
}

// JumpListEntry represents a position in jump history
type JumpListEntry struct {
	line int
	col  int
}

type Editor struct {
	s tcell.Screen

	// multiple buffers
	buffers       []*BufferView
	currentBuffer int

	// Current buffer state (synced when switching buffers)
	buffer               *buffer.Buffer
	filename             string
	dirty                bool
	cx, cy               int
	rowOffset, colOffset int
	wantX                int
	marks                map[rune]Mark
	jumpList             []JumpListEntry
	jumpListIndex        int

	mode Mode

	// command mode
	cmdBuf            []rune
	cmdHistory        []string // command history
	cmdHistoryIdx     int      // current position in history (-1 = not browsing)
	cmdHistorySave    []rune   // saved current input when browsing history
	cmdCompletions    []string // available completions for current input
	cmdCompletionIdx  int      // current completion index (-1 = not completing)
	cmdCompletionSave []rune   // saved input before completion
	statusMsg         string

	// operator pending
	pendingCount   int
	pendingOp      rune // d/c/y/g
	pendingOpCount int  // captured count at time op was entered
	pendingTextObj rune // 'i' or 'a' when waiting for iw/aw etc

	// registers (shared across all buffers)
	regs             Registers
	awaitingRegister bool
	regOverride      rune
	regOverrideSet   bool

	// visual
	visualKind   VisualKind
	visualAnchor int // absolute rune pos
	visualActive bool

	// dot repeat
	last          RepeatAction
	insertCapture []rune // text typed during current insert session

	// search
	searchQuery   string // current search pattern
	searchForward bool   // true for /, false for ?
	searchBuf     []rune // input buffer while typing search
	searchMatches []int  // positions of matches in viewport (for highlighting)

	// character find (f/F/t/T)
	lastCharFind     rune // character to find
	lastCharFindKind rune // 'f', 'F', 't', or 'T'
	awaitingCharFind rune // waiting for character after f/F/t/T

	// completion
	completionActive     bool     // true when cycling through completions
	completionCandidates []string // all word candidates
	completionIndex      int      // current selection index
	completionPrefix     string   // the partial word being completed
	completionStartPos   int      // position where completion started

	// marks - awaitingMarkSet/Jump for current operation
	awaitingMarkSet  bool // waiting for mark name after 'm'
	awaitingMarkJump rune // waiting for mark name after ' or `

	// macros
	recordingMacro    bool           // true when recording a macro
	recordingRegister rune           // register being recorded to (a-z)
	macroKeys         []MacroKey     // keys being recorded
	macros            map[rune]Macro // stored macros (a-z)
	playingMacro      bool           // true when playing back a macro
	awaitingMacroPlay rune           // waiting for register after @

	// configuration
	indentWidth int            // number of spaces for indentation
	config      *config.Config // user configuration
	loader      *config.Loader // Lua runtime (for notifications, scheduled tasks, etc.)

	// notifications
	notifTimeout time.Time                // When current notification should clear
	notifLevel   config.NotificationLevel // Level of current notification

	// popup UI
	popupActive bool
	popupTitle  string
	popupLines  []string
	popupScroll int
	popupFixedH int

	// file tree
	fileTree       *FileTree
	treeOpen       bool
	treePanelWidth int
	focusTree      bool // true if tree has focus, false if buffer has focus

	// splits
	splits         []*Split // list of splits
	currentSplit   int      // index of focused split
	awaitingWindow bool     // waiting for window command after Ctrl+W

	// folding
	foldRanges map[int]*FoldRange // map of start line to fold range
	parser     *TreeSitterParser  // tree-sitter parser for current buffer
}

func newEditorFromFile(path string, cfg *config.Config, loader *config.Loader) (*Editor, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	txt := ""
	data, readErr := os.ReadFile(abs)
	if readErr == nil {
		txt = strings.ReplaceAll(string(data), "\r\n", "\n")
	} else {
		txt = ""
	}

	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.Sync()

	bufView := NewBufferView(txt, abs)

	// Apply tab width from config
	indentWidth := cfg.TabWidth
	if indentWidth < 1 {
		indentWidth = 4
	}

	ed := &Editor{
		s:             s,
		buffers:       []*BufferView{bufView},
		currentBuffer: 0,
		mode:          ModeNormal,
		indentWidth:   indentWidth,
		config:        cfg,
		loader:        loader,
		foldRanges:    make(map[int]*FoldRange),
	}
	ed.regs.named = make(map[rune]Register)
	ed.macros = make(map[rune]Macro)
	ed.syncFromBuffer()

	// Initialize splits
	ed.initSplits()

	// Apply filetype-specific options (this initializes tree-sitter parser)
	ed.setFiletypeOptions()

	// Sync parser from buffer view after filetype init
	if bv := ed.buf(); bv != nil && bv.parser != nil {
		ed.parser = bv.parser
		ed.UpdateFoldStates()
	}

	// Register editor as context for Lua buffer operations
	ed.RegisterWithLoader()

	if readErr != nil && !os.IsNotExist(readErr) {
		ed.statusMsg = "read failed: " + readErr.Error()
	} else if readErr != nil && os.IsNotExist(readErr) {
		ed.statusMsg = "new file"
	}

	// Fire startup events
	ed.FireVimEnter()
	ed.FireEditorReady()
	ed.FireBufRead()
	if ft := ed.getFiletype(); ft != nil {
		ed.FireFileType(ft.Name)
	}
	ed.FireBufEnter()

	return ed, nil
}

func newEditorFromProject(path string, cfg *config.Config, loader *config.Loader) (*Editor, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.Sync()

	// Apply tab width from config
	indentWidth := cfg.TabWidth
	if indentWidth < 1 {
		indentWidth = 4
	}

	ed := &Editor{
		s:           s,
		buffer:      buffer.NewFromString(""),
		mode:        ModeNormal,
		filename:    abs,
		indentWidth: indentWidth,
		config:      cfg,
		loader:      loader,
	}
	ed.regs.named = make(map[rune]Register)
	ed.marks = make(map[rune]Mark)
	ed.macros = make(map[rune]Macro)
	ed.cmdHistory = make([]string, 0, 100)
	ed.cmdHistoryIdx = -1
	ed.cmdCompletionIdx = -1
	ed.jumpList = make([]JumpListEntry, 0, 100)
	ed.jumpListIndex = -1

	// Initialize splits
	ed.initSplits()

	// Register editor as context for Lua buffer operations
	ed.RegisterWithLoader()

	// Fire startup events
	ed.FireVimEnter()
	ed.FireEditorReady()

	return ed, nil
}

func (e *Editor) run() error {
	defer e.s.Fini()
	defer e.FireVimLeave()

	for {
		// Process notifications from Lua
		e.processNotifications()

		e.ensureCursorValid()
		e.ensureCursorVisible()
		e.draw()

		ev := e.s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlQ {
				return nil
			}
			if e.handleKey(ev) {
				return nil
			}
		case *tcell.EventResize:
			e.s.Sync()
		}
	}
}

// buf returns the current BufferView
func (e *Editor) buf() *BufferView {
	if len(e.buffers) == 0 || e.currentBuffer < 0 || e.currentBuffer >= len(e.buffers) {
		return nil
	}
	return e.buffers[e.currentBuffer]
}

// syncToBuffer copies editor state to current buffer before switching
func (e *Editor) syncToBuffer() {
	if b := e.buf(); b != nil {
		b.buffer = e.buffer
		b.filename = e.filename
		b.dirty = e.dirty
		b.cx = e.cx
		b.cy = e.cy
		b.rowOffset = e.rowOffset
		b.colOffset = e.colOffset
		b.wantX = e.wantX
		b.marks = e.marks
		b.jumpList = e.jumpList
		b.jumpListIndex = e.jumpListIndex
	}
}

// syncFromBuffer copies current buffer state to editor after switching
func (e *Editor) syncFromBuffer() {
	if b := e.buf(); b != nil {
		e.buffer = b.buffer
		e.filename = b.filename
		e.dirty = b.dirty
		e.cx = b.cx
		e.cy = b.cy
		e.rowOffset = b.rowOffset
		e.colOffset = b.colOffset
		e.wantX = b.wantX
		e.marks = b.marks
		e.jumpList = b.jumpList
		e.jumpListIndex = b.jumpListIndex
		e.parser = b.parser

		// Initialize fold ranges if parser exists
		if e.parser != nil {
			if e.foldRanges == nil {
				e.foldRanges = make(map[int]*FoldRange)
			}
			e.UpdateFoldStates()
		}
	}
}
