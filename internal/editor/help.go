package editor

import "strings"

// helpTopics contains all help documentation
var helpTopics = map[string]string{
	"":                helpMain,
	"main":            helpMain,
	"lua":             helpLua,
	"keymap":          helpKeymap,
	"events":          helpEvents,
	"defaults":        helpDefaults,
	"vim-differences": helpVimDifferences,
	"completion":      helpCompletion,
	"undo":            helpUndo,
}

const helpMain = `VOIDABYSS HELP - A Vim-inspired modal text editor

QUICK START
  :help defaults          - Default keymaps and settings
  :help vim-differences   - How this differs from Vim
  :help lua               - Lua API reference
  :help keymap            - Custom keymaps
  :help events            - Event system
  :help completion        - Word completion
  :help undo              - Undo system

BASIC COMMANDS
  :e file     - Open file
  :w          - Save
  :q          - Quit
  :help       - This help
`

const helpLua = `LUA API REFERENCE

vb.opt            - Editor options (tabwidth, leader, etc.)
vb.keymap(...)    - Define custom keymaps
vb.command(...)   - Define custom commands
vb.on(...)        - Register event handlers
vb.notify(...)    - Show status message
vb.has(...)       - Check feature availability

See full documentation at :help lua
`

const helpKeymap = `KEYMAP SYSTEM

Define custom keymaps with vb.keymap(mode, lhs, rhs, opts)

Example:
  vb.keymap("n", "<leader>w", ":w<CR>")
  vb.keymap("i", "jj", "<Esc>")

Modes: "n" (normal), "i" (insert), "v" (visual)
`

const helpEvents = `EVENT SYSTEM

Register event handlers with vb.on(event, callback, opts)

Available events:
  - EditorReady
  - BufWritePost
  - ModeChanged

Example:
  vb.on("BufWritePost", function(ctx)
    vb.notify("Saved: " .. ctx.file)
  end)
`

const helpDefaults = `DEFAULT KEYMAPS

Normal Mode:
  h j k l     - Move cursor
  w b e       - Word motions
  i a A o O   - Enter insert mode
  dd yy p     - Delete, yank, paste
  u Ctrl-R    - Undo, redo
  / n N       - Search

Insert Mode:
  Esc         - Exit insert
  Ctrl-N/P    - Completion

Commands:
  :w :q :wq   - Save, quit
  :e file     - Open file
  :bn :bp     - Next/prev buffer
`

const helpVimDifferences = `DIFFERENCES FROM VIM

Similar to Vim:
  - Modal editing (normal/insert/visual)
  - hjkl navigation, w/b/e motions
  - Operators + text objects (dw, ciw, etc.)
  - Search, undo/redo, registers

Different from Vim:
  - Lua configuration (not Vimscript)
  - No splits/tabs/windows
  - No macros yet
  - Simpler plugin system
  - Built-in file tree

Philosophy: Core Vim editing with modern extensibility
`

const helpCompletion = `COMPLETION SYSTEM

Word completion with Ctrl-N/Ctrl-P in insert mode.

Features:
  - Stable sorting (shorter words first)
  - Match highlighting [prefix]remainder
  - Part of undo transaction
  - Vim-compatible word boundaries

Usage:
  1. Type a few characters
  2. Press Ctrl-N for next match
  3. Press Ctrl-P for previous match
  4. Esc to cancel
`

const helpUndo = `UNDO AND REDO SYSTEM

u           - Undo last change
Ctrl-R      - Redo

Transaction Grouping:
  - Each insert session = one undo
  - Newline breaks transaction
  - Completion included in transaction

Example:
  i              - Start undo group
  type "hello"   - All one undo
  Esc            - Commit group
  u              - Undoes entire "hello"
`

// GetHelp returns help content for a given topic
func GetHelp(topic string) (string, bool) {
	topic = strings.TrimSpace(strings.ToLower(topic))
	if topic == "vim" || topic == "differences" {
		topic = "vim-differences"
	}
	content, ok := helpTopics[topic]
	return content, ok
}

// ListHelpTopics returns all available help topics
func ListHelpTopics() []string {
	topics := make([]string, 0, len(helpTopics))
	for topic := range helpTopics {
		if topic != "" && topic != "main" {
			topics = append(topics, topic)
		}
	}
	return topics
}
