# Advanced Features Guide

This guide covers VoidAbyss's advanced features including code folding and custom color schemes.

## Table of Contents

- [Code Folding](#code-folding)
- [Color Schemes](#color-schemes)
- [Future Features](#future-features)

---

## Code Folding

VoidAbyss includes intelligent code folding powered by tree-sitter. Folds are automatically detected based on the AST structure of your code, ensuring semantic folding for functions, classes, blocks, and other multi-line constructs.

### Supported Languages

Code folding works with any language that has tree-sitter support:
- **Go**: Functions, methods, structs, interfaces, type declarations, if/for blocks
- **Python**: Functions, classes, if/for/while/with/try statements
- **JavaScript/TypeScript**: Functions, arrow functions, classes, methods, objects, arrays, blocks

### Fold Commands

#### Command Mode
Use these commands in command mode (`:` prefix):

| Command | Description |
|---------|-------------|
| `:fold` | Toggle fold at cursor position |
| `:foldopen` or `:fo` | Open fold at cursor |
| `:foldclose` or `:fc` | Close fold at cursor |
| `:foldall` or `:fca` | Fold all foldable regions |
| `:unfoldall` or `:ufa` | Unfold all regions |

#### Normal Mode Keybindings
Use these Vim-style keybindings in normal mode:

| Key | Action |
|-----|--------|
| `za` | Toggle fold at cursor |
| `zo` | Open fold at cursor |
| `zc` | Close fold at cursor |
| `zM` | Fold all (fold more) |
| `zR` | Unfold all (reduce folds) |

### Fold Indicators

When line numbers are enabled, fold indicators appear next to line numbers:

- `▼` - Indicates a foldable region that is currently open
- `▶` - Indicates a folded region (collapsed)
- ` ` - Regular line (not foldable)

### How Folding Works

1. **Automatic Detection**: When a file is opened or edited, tree-sitter parses the code and identifies foldable regions based on AST nodes (functions, classes, blocks, etc.)

2. **Fold State Preservation**: Fold states are maintained when you edit the file. If you fold a function and add code inside it, the fold remains closed.

3. **Visual Feedback**: Folded regions are hidden from view. The first line of the fold shows the fold indicator (`▶`), and cursor movement skips over hidden lines.

4. **Incremental Updates**: Folds are automatically updated after each edit to reflect the current code structure.

### Example Workflow

```go
// Open a Go file with functions
:e myfile.go

// Fold the current function
za

// Fold all functions to get an overview
zM

// Unfold to see details
zR

// Toggle specific fold
zo
```

### Tips

- Use `zM` to collapse all folds when exploring a new file
- Use `za` frequently to toggle folds at your cursor position  
- Fold indicators help you quickly see code structure at a glance
- Folds are preserved during buffer switches

---

## Color Schemes

VoidAbyss includes 5 built-in color schemes with carefully chosen colors for syntax highlighting and UI elements.

### Available Color Schemes

| Name | Description | Style |
|------|-------------|-------|
| `default` | Classic terminal colors | Dark background, vibrant syntax colors |
| `monokai` | Popular Sublime Text theme | Dark, high contrast |
| `gruvbox` | Retro groove colors | Warm, low contrast |
| `solarized-dark` | Precision colors for readability | Cool, balanced contrast |
| `dracula` | Dark theme with vibrant colors | Modern, eye-friendly |

### Changing Color Schemes

#### Command Mode
```vim
:colorscheme monokai
:colorscheme gruvbox
:colorscheme solarized-dark
:colorscheme dracula
:colorscheme default
```

#### List Available Schemes
```vim
:colorschemes
```

This shows all available schemes with `*` marking the currently active one.

### Configuration

Set your preferred color scheme in `~/.config/voidabyss/init.lua`:

```lua
vb.opt.colorscheme = "monokai"
```

The color scheme setting is automatically persisted in your configuration.

### Color Scheme Details

#### Default
- Classic terminal aesthetic
- High contrast for readability
- Yellow line numbers, purple keywords, blue functions
- Green strings, orange numbers, gray comments

#### Monokai
- Inspired by Sublime Text's Monokai theme
- Dark charcoal background (#272822)
- Vibrant pink keywords, yellow strings
- Great for long coding sessions

#### Gruvbox
- Retro-inspired warm colors
- Brown/orange tones for reduced eye strain
- Excellent for extended use
- Balanced contrast

#### Solarized Dark
- Precision colors designed for readability
- Blue-tinted dark background
- Scientifically selected color palette
- Reduces eye fatigue

#### Dracula
- Modern dark theme with vibrant accents
- Pink keywords, cyan types, green functions
- Very popular in the developer community
- High readability with style

### What Gets Colored

Color schemes affect:

**Syntax Highlighting:**
- Keywords (if, for, func, class, etc.)
- Functions and methods
- Types and classes
- Strings
- Numbers
- Comments
- Constants
- Properties and variables
- Operators

**Editor UI:**
- Background and foreground
- Line numbers
- Status line
- Visual selection
- Search matches
- Cursor
- File tree (directory colors, borders, cursor)

### Creating Custom Color Schemes

To add your own color scheme, modify `internal/editor/colorscheme.go` and add a new entry to the `colorSchemes` map:

```go
"mytheme": {
    Name:           "mytheme",
    Background:     tcell.NewRGBColor(20, 20, 20),
    Foreground:     tcell.NewRGBColor(220, 220, 220),
    LineNumber:     tcell.NewRGBColor(100, 100, 100),
    // ... define all colors
},
```

All colors use `tcell.Color` which supports:
- Named colors: `tcell.ColorBlue`, `tcell.ColorRed`
- RGB colors: `tcell.NewRGBColor(r, g, b)` where r, g, b are 0-255

### Syntax Highlighting Colors

The following syntax elements are themed:

| Element | Example | Default Color |
|---------|---------|---------------|
| Keyword | `func`, `if`, `class` | Purple |
| Function | Function names | Blue |
| Type | `string`, `int`, class names | Cyan |
| String | `"hello"` | Green |
| Number | `42`, `3.14` | Orange |
| Comment | `// comment` | Gray (dimmed) |
| Constant | `true`, `false`, `nil` | Red/Orange |
| Property | Object properties | Light Blue |

These colors change based on the active color scheme.

### Tips

- Try different schemes in different lighting conditions
- Dark schemes (monokai, dracula, gruvbox) are great for night coding
- Solarized-dark is excellent for reducing eye strain
- Use `:colorschemes` to quickly browse and compare
- The active scheme is shown in `:set` output

---

## Future Features

The following advanced features are planned for future releases:

### Jump to Definition
- Navigate to function/class definitions using tree-sitter
- Cross-file navigation for imports
- Keybinding: `gd` (go to definition)

### Smart AST-based Selections
- Select entire functions, classes, blocks with a single command
- Expand/contract selections semantically
- Perfect for refactoring large blocks of code

### LSP Integration
- Language Server Protocol support
- Auto-completion from LSP
- Real-time diagnostics
- Hover documentation
- Symbol references

### Enhanced Folding
- Fold by indent level
- Fold by syntax element (all functions, all classes)
- Persist fold states across sessions
- Manual fold regions

---

## Configuration Examples

### Minimal Setup

```lua
-- ~/.config/voidabyss/init.lua

-- Use monokai color scheme
vb.opt.colorscheme = "monokai"

-- Enable line numbers (folds show indicators here)
vb.opt.number = true
```

### Power User Setup

```lua
-- Advanced configuration

-- Color scheme
vb.opt.colorscheme = "dracula"

-- Line numbers with relative numbering
vb.opt.number = true
vb.opt.relativenumber = true

-- Key mappings for quick fold operations
vb.keymap("n", "<leader>z", ":fold<CR>", { desc = "toggle fold" })
vb.keymap("n", "<leader>Z", ":foldall<CR>", { desc = "fold all" })
vb.keymap("n", "<leader>u", ":unfoldall<CR>", { desc = "unfold all" })

-- Auto-fold on opening large files
vb.on("BufRead", function()
    local lineCount = vb.buf.line_count or 0
    if lineCount > 500 then
        vb.cmd("foldall")
        vb.notify("Auto-folded large file", "info")
    end
end)

-- Change color scheme based on time of day
local hour = tonumber(os.date("%H"))
if hour >= 8 and hour < 18 then
    -- Daytime: lighter scheme
    vb.opt.colorscheme = "default"
else
    -- Nighttime: eye-friendly dark scheme
    vb.opt.colorscheme = "gruvbox"
end
```

---

## Troubleshooting

### Folding Issues

**Folds don't appear:**
- Ensure the file has tree-sitter support (Go, Python, JavaScript, TypeScript)
- Check that line numbers are enabled: `:set number`
- Verify the file has been parsed: Make a small edit to trigger reparse

**Folds disappear after editing:**
- This is expected - folds update based on the current AST
- If code structure changes significantly, old folds may not apply
- Use `za` to re-fold after major edits

**Can't fold at cursor position:**
- Ensure cursor is on or near a foldable line (function start, class start, etc.)
- Use `▼` indicators in line numbers to find foldable regions
- Try pressing `za` on the line above or below

### Color Scheme Issues

**Colors don't change:**
- Restart VoidAbyss after changing color scheme in init.lua
- Or use `:colorscheme <name>` to change immediately
- Check terminal supports 256 colors or true color

**Colors look different than expected:**
- Terminal color support varies
- Some terminals don't support RGB colors
- Try a different terminal emulator (iTerm2, Alacritty, etc.)

**How to reset to default:**
```vim
:colorscheme default
```

---

## Keyboard Reference

### Folding

| Command | Description |
|---------|-------------|
| `za` | Toggle fold at cursor |
| `zo` | Open fold |
| `zc` | Close fold |
| `zM` | Fold all |
| `zR` | Unfold all |
| `:fold` | Toggle fold (command mode) |
| `:foldall` | Fold all (command mode) |
| `:unfoldall` | Unfold all (command mode) |

### Color Schemes

| Command | Description |
|---------|-------------|
| `:colorscheme <name>` | Change color scheme |
| `:colorschemes` | List all schemes |
| `:set` | Show current scheme (and other settings) |

---

## Performance Notes

- **Code Folding**: Minimal performance impact. Folds are computed only when code changes.
- **Color Schemes**: No performance impact. Colors are applied during rendering.
- **Tree-sitter**: Incremental parsing is very fast (< 1ms for typical edits).

Both features are designed to work smoothly even with large files (10,000+ lines).

---

## Contributing

Have ideas for improving these features? Contributions are welcome!

- Add new color schemes
- Suggest fold behaviors
- Report bugs or unexpected behavior

See the main repository for contribution guidelines.

---

**Happy coding with VoidAbyss!** 🚀

Use `za` to fold and `:colorscheme` to customize your experience.
