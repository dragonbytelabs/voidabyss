package config

const Version = "0.1.0"

// Options holds all configurable editor options
type Options struct {
	// Core/editor
	TabWidth       int
	ExpandTab      bool
	CursorLine     bool
	RelativeNumber bool
	Wrap           bool
	ScrollOff      int
	Leader         string
	Number         bool

	// UI
	StatusLine string
}

// DefaultOptions returns default option values
func DefaultOptions() *Options {
	return &Options{
		TabWidth:       4,
		ExpandTab:      false,
		CursorLine:     false,
		RelativeNumber: false,
		Wrap:           true,
		ScrollOff:      0,
		Leader:         "\\",
		Number:         true,
		StatusLine:     "default",
	}
}

// KeyMapOpts represents keymap options
type KeyMapOpts struct {
	Noremap bool
	Silent  bool
	Desc    string
	Expr    bool
	Nowait  bool
}

// DefaultKeyMapOpts returns default keymap options
func DefaultKeyMapOpts() KeyMapOpts {
	return KeyMapOpts{
		Noremap: true,
		Silent:  true,
		Desc:    "",
		Expr:    false,
		Nowait:  false,
	}
}

// KeyMapping represents a key mapping with callback support
type KeyMapping struct {
	Mode   string
	LHS    string
	RHS    string      // String command or empty if function
	Fn     interface{} // Lua function reference
	Opts   KeyMapOpts
	IsFunc bool // True if Fn is set
}

// CommandOpts represents command options
type CommandOpts struct {
	NArgs int // 0, 1, or -1 for "?"
	Desc  string
}

// Command represents a user command
type Command struct {
	Name   string
	RHS    string      // String command or empty if function
	Fn     interface{} // Lua function reference
	Opts   CommandOpts
	IsFunc bool
}

// EventOpts represents autocmd event options
type EventOpts struct {
	Pattern string
	Once    bool
}

// EventHandler represents an event/autocmd handler
type EventHandler struct {
	Event string
	Fn    interface{} // Lua function reference
	Opts  EventOpts
}

// Features tracks available features for vb.has()
var Features = map[string]bool{
	"keymap.leader":       true,
	"events.bufwritepost": true,
	"completion.ctrl-n":   true,
	"opt.tabwidth":        true,
	"opt.expandtab":       true,
	"opt.leader":          true,
	"buffer.api":          true,
	"state.persistent":    true,
	"notify":              true,
}
