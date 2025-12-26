package editor

import (
	"github.com/gdamore/tcell/v2"
)

// ColorScheme defines colors for various UI elements
type ColorScheme struct {
	Name string

	// Editor UI
	Background   tcell.Color
	Foreground   tcell.Color
	LineNumber   tcell.Color
	StatusLine   tcell.Color
	StatusLineBg tcell.Color
	Visual       tcell.Color
	VisualBg     tcell.Color
	Search       tcell.Color
	SearchBg     tcell.Color
	Cursor       tcell.Color
	CursorBg     tcell.Color

	// Syntax highlighting
	Keyword  tcell.Color
	Function tcell.Color
	Type     tcell.Color
	String   tcell.Color
	Number   tcell.Color
	Comment  tcell.Color
	Constant tcell.Color
	Property tcell.Color
	Operator tcell.Color
	Variable tcell.Color

	// File tree
	TreeDirectory tcell.Color
	TreeFile      tcell.Color
	TreeCursor    tcell.Color
	TreeCursorBg  tcell.Color
	TreeBorder    tcell.Color
}

// Built-in color schemes
var colorSchemes = map[string]*ColorScheme{
	"default": {
		Name:          "default",
		Background:    tcell.ColorBlack,
		Foreground:    tcell.ColorWhite,
		LineNumber:    tcell.ColorYellow,
		StatusLine:    tcell.ColorWhite,
		StatusLineBg:  tcell.ColorBlue,
		Visual:        tcell.ColorWhite,
		VisualBg:      tcell.ColorBlue,
		Search:        tcell.ColorBlack,
		SearchBg:      tcell.ColorYellow,
		Cursor:        tcell.ColorWhite,
		CursorBg:      tcell.ColorBlack,
		Keyword:       tcell.ColorPurple,
		Function:      tcell.ColorBlue,
		Type:          tcell.NewRGBColor(0, 200, 200), // Cyan
		String:        tcell.ColorGreen,
		Number:        tcell.NewRGBColor(255, 165, 0), // Orange
		Comment:       tcell.ColorGray,
		Constant:      tcell.ColorRed,
		Property:      tcell.NewRGBColor(173, 216, 230), // Light blue
		Operator:      tcell.ColorWhite,
		Variable:      tcell.ColorWhite,
		TreeDirectory: tcell.ColorGreen,
		TreeFile:      tcell.ColorWhite,
		TreeCursor:    tcell.ColorWhite,
		TreeCursorBg:  tcell.ColorDarkGreen,
		TreeBorder:    tcell.ColorGray,
	},
	"monokai": {
		Name:          "monokai",
		Background:    tcell.NewRGBColor(39, 40, 34),
		Foreground:    tcell.NewRGBColor(248, 248, 242),
		LineNumber:    tcell.NewRGBColor(144, 144, 144),
		StatusLine:    tcell.NewRGBColor(248, 248, 242),
		StatusLineBg:  tcell.NewRGBColor(39, 40, 34),
		Visual:        tcell.NewRGBColor(248, 248, 242),
		VisualBg:      tcell.NewRGBColor(73, 72, 62),
		Search:        tcell.NewRGBColor(0, 0, 0),
		SearchBg:      tcell.NewRGBColor(255, 255, 0),
		Cursor:        tcell.NewRGBColor(248, 248, 242),
		CursorBg:      tcell.NewRGBColor(249, 38, 114),
		Keyword:       tcell.NewRGBColor(249, 38, 114),
		Function:      tcell.NewRGBColor(166, 226, 46),
		Type:          tcell.NewRGBColor(102, 217, 239),
		String:        tcell.NewRGBColor(230, 219, 116),
		Number:        tcell.NewRGBColor(174, 129, 255),
		Comment:       tcell.NewRGBColor(117, 113, 94),
		Constant:      tcell.NewRGBColor(174, 129, 255),
		Property:      tcell.NewRGBColor(166, 226, 46),
		Operator:      tcell.NewRGBColor(249, 38, 114),
		Variable:      tcell.NewRGBColor(248, 248, 242),
		TreeDirectory: tcell.NewRGBColor(166, 226, 46),
		TreeFile:      tcell.NewRGBColor(248, 248, 242),
		TreeCursor:    tcell.NewRGBColor(248, 248, 242),
		TreeCursorBg:  tcell.NewRGBColor(73, 72, 62),
		TreeBorder:    tcell.NewRGBColor(117, 113, 94),
	},
	"gruvbox": {
		Name:          "gruvbox",
		Background:    tcell.NewRGBColor(40, 40, 40),
		Foreground:    tcell.NewRGBColor(235, 219, 178),
		LineNumber:    tcell.NewRGBColor(124, 111, 100),
		StatusLine:    tcell.NewRGBColor(235, 219, 178),
		StatusLineBg:  tcell.NewRGBColor(60, 56, 54),
		Visual:        tcell.NewRGBColor(235, 219, 178),
		VisualBg:      tcell.NewRGBColor(102, 92, 84),
		Search:        tcell.NewRGBColor(40, 40, 40),
		SearchBg:      tcell.NewRGBColor(250, 189, 47),
		Cursor:        tcell.NewRGBColor(40, 40, 40),
		CursorBg:      tcell.NewRGBColor(235, 219, 178),
		Keyword:       tcell.NewRGBColor(251, 73, 52),
		Function:      tcell.NewRGBColor(184, 187, 38),
		Type:          tcell.NewRGBColor(250, 189, 47),
		String:        tcell.NewRGBColor(184, 187, 38),
		Number:        tcell.NewRGBColor(211, 134, 155),
		Comment:       tcell.NewRGBColor(146, 131, 116),
		Constant:      tcell.NewRGBColor(211, 134, 155),
		Property:      tcell.NewRGBColor(142, 192, 124),
		Operator:      tcell.NewRGBColor(235, 219, 178),
		Variable:      tcell.NewRGBColor(131, 165, 152),
		TreeDirectory: tcell.NewRGBColor(142, 192, 124),
		TreeFile:      tcell.NewRGBColor(235, 219, 178),
		TreeCursor:    tcell.NewRGBColor(235, 219, 178),
		TreeCursorBg:  tcell.NewRGBColor(102, 92, 84),
		TreeBorder:    tcell.NewRGBColor(124, 111, 100),
	},
	"solarized-dark": {
		Name:          "solarized-dark",
		Background:    tcell.NewRGBColor(0, 43, 54),
		Foreground:    tcell.NewRGBColor(131, 148, 150),
		LineNumber:    tcell.NewRGBColor(88, 110, 117),
		StatusLine:    tcell.NewRGBColor(131, 148, 150),
		StatusLineBg:  tcell.NewRGBColor(7, 54, 66),
		Visual:        tcell.NewRGBColor(131, 148, 150),
		VisualBg:      tcell.NewRGBColor(7, 54, 66),
		Search:        tcell.NewRGBColor(0, 0, 0),
		SearchBg:      tcell.NewRGBColor(181, 137, 0),
		Cursor:        tcell.NewRGBColor(0, 43, 54),
		CursorBg:      tcell.NewRGBColor(131, 148, 150),
		Keyword:       tcell.NewRGBColor(133, 153, 0),
		Function:      tcell.NewRGBColor(38, 139, 210),
		Type:          tcell.NewRGBColor(181, 137, 0),
		String:        tcell.NewRGBColor(42, 161, 152),
		Number:        tcell.NewRGBColor(203, 75, 22),
		Comment:       tcell.NewRGBColor(88, 110, 117),
		Constant:      tcell.NewRGBColor(203, 75, 22),
		Property:      tcell.NewRGBColor(42, 161, 152),
		Operator:      tcell.NewRGBColor(131, 148, 150),
		Variable:      tcell.NewRGBColor(38, 139, 210),
		TreeDirectory: tcell.NewRGBColor(38, 139, 210),
		TreeFile:      tcell.NewRGBColor(131, 148, 150),
		TreeCursor:    tcell.NewRGBColor(131, 148, 150),
		TreeCursorBg:  tcell.NewRGBColor(7, 54, 66),
		TreeBorder:    tcell.NewRGBColor(88, 110, 117),
	},
	"dracula": {
		Name:          "dracula",
		Background:    tcell.NewRGBColor(40, 42, 54),
		Foreground:    tcell.NewRGBColor(248, 248, 242),
		LineNumber:    tcell.NewRGBColor(98, 114, 164),
		StatusLine:    tcell.NewRGBColor(248, 248, 242),
		StatusLineBg:  tcell.NewRGBColor(68, 71, 90),
		Visual:        tcell.NewRGBColor(248, 248, 242),
		VisualBg:      tcell.NewRGBColor(68, 71, 90),
		Search:        tcell.NewRGBColor(40, 42, 54),
		SearchBg:      tcell.NewRGBColor(241, 250, 140),
		Cursor:        tcell.NewRGBColor(40, 42, 54),
		CursorBg:      tcell.NewRGBColor(248, 248, 242),
		Keyword:       tcell.NewRGBColor(255, 121, 198),
		Function:      tcell.NewRGBColor(80, 250, 123),
		Type:          tcell.NewRGBColor(139, 233, 253),
		String:        tcell.NewRGBColor(241, 250, 140),
		Number:        tcell.NewRGBColor(189, 147, 249),
		Comment:       tcell.NewRGBColor(98, 114, 164),
		Constant:      tcell.NewRGBColor(189, 147, 249),
		Property:      tcell.NewRGBColor(80, 250, 123),
		Operator:      tcell.NewRGBColor(255, 121, 198),
		Variable:      tcell.NewRGBColor(248, 248, 242),
		TreeDirectory: tcell.NewRGBColor(139, 233, 253),
		TreeFile:      tcell.NewRGBColor(248, 248, 242),
		TreeCursor:    tcell.NewRGBColor(248, 248, 242),
		TreeCursorBg:  tcell.NewRGBColor(68, 71, 90),
		TreeBorder:    tcell.NewRGBColor(98, 114, 164),
	},
}

// GetColorScheme returns the color scheme by name, or default if not found
func GetColorScheme(name string) *ColorScheme {
	if scheme, ok := colorSchemes[name]; ok {
		return scheme
	}
	return colorSchemes["default"]
}

// ListColorSchemes returns all available color scheme names
func ListColorSchemes() []string {
	schemes := make([]string, 0, len(colorSchemes))
	for name := range colorSchemes {
		schemes = append(schemes, name)
	}
	return schemes
}

// ApplyColorScheme applies a color scheme to the editor
func (e *Editor) ApplyColorScheme(scheme *ColorScheme) {
	if scheme == nil {
		scheme = colorSchemes["default"]
	}

	if e.config != nil {
		// Store scheme name in config for persistence
		e.config.ColorScheme = scheme.Name
	}

	// Color scheme will be used in draw() function
	e.statusMsg = "color scheme: " + scheme.Name
}
