# Voidabyss

A modal text editor written in Go with Vim-like keybindings and Lua configuration.

## Features

### Core Editor
- **Modal editing**: Normal, Insert, Visual, Command, and Search modes
- **Piece table buffer**: Efficient undo/redo with O(1) operations
- **Text objects**: `iw/aw`, `iW/aW`, `ip/ap`, `i"/a"`, `i(/a(`, `i{/a{`, `i[/a[`
- **Visual selection**: Character and line-wise selection with highlighting
- **Search**: Forward/backward search with pattern highlighting
- **Marks**: Set and jump to marks (`m{a-z}`, `'{a-z}`)
- **Jump list**: Navigate through cursor history (`Ctrl+O`, `Ctrl+I`)
- **Registers**: Named registers, clipboard integration
- **Dot repeat**: Repeat last operation with `.`
- **Auto-indentation**: Smart indentation for code editing

### Multi-Buffer Support
- **Buffer management**: Open multiple files in tabs
- **Commands**: `:e file`, `:bn/:bnext`, `:bp/:bprev`, `:ls/:buffers`
- **Per-buffer state**: Independent cursor, marks, jump list, and dirty flag
- **Dirty tracking**: Visual indicator for unsaved changes

### Project Navigation
- **File tree**: Browse project files with `:Explore` or `:Ex`
- **Tree navigation**: 
  - `j/k` - move cursor
  - `Enter` - expand/collapse directories, open files
  - `c` - collapse current or parent directory
  - `q` - quit tree view
  - `Ctrl+W` - toggle focus between tree and buffer
- **Smart ignore**: Automatically ignores `.git`, `node_modules`, `dist`, etc.
- **Split view**: Tree panel + content area with vertical border

### Configuration System
- **Lua configuration**: Familiar Neovim-style API with `init.lua`
- **Config location**: 
  - macOS/Linux: `~/.config/voidabyss/init.lua`
  - Windows: `%APPDATA%\voidabyss\init.lua`
- **Display options**: Tab width, line numbers (absolute/relative)
- **Key mappings**: Define custom keybindings
- **Plugin system**: Load external plugins via RPC (framework ready)
- **Hooks**: `on_startup()`, `on_file_open()` for automation

See [CONFIG.md](docs/CONFIG.md) for detailed configuration documentation.

## Installation

### From Source

```bash
git clone https://github.com/dragonbytelabs/voidabyss.git
cd voidabyss
go build -o vb ./cmd/vb
./vb [file]
```

### Requirements

- Go 1.24.4 or later
- Terminal with tcell support

## Usage

```bash
# Open a file
vb myfile.txt

# Open current directory (project mode with file tree)
vb .

# Open specific directory
vb /path/to/project

# Show version
vb --version

# Show help
vb --help
```

## Key Bindings

### Normal Mode

#### Motion
- `h/j/k/l` - left/down/up/right
- `w/b` - word forward/backward
- `W/B` - WORD forward/backward
- `0/$` - start/end of line
- `gg/G` - first/last line
- `{/}` - paragraph up/down
- `%` - matching bracket

#### Editing
- `i/a` - insert before/after cursor
- `I/A` - insert at start/end of line
- `o/O` - open line below/above
- `x` - delete character
- `dd` - delete line
- `D` - delete to end of line
- `cc` - change line
- `C` - change to end of line
- `yy` - yank (copy) line
- `p/P` - paste after/before
- `u` - undo
- `Ctrl+R` - redo
- `.` - repeat last operation

#### Visual Mode
- `v` - start visual character mode
- `V` - start visual line mode
- `d/c/y` - delete/change/yank selection

#### Search
- `/` - search forward
- `?` - search backward
- `n/N` - next/previous match

#### Marks & Jumps
- `m{a-z}` - set mark
- `'{a-z}` - jump to mark
- `Ctrl+O/Ctrl+I` - jump backward/forward

#### File Tree
- `:Explore` or `:Ex` - open file tree
- `Ctrl+W` - toggle focus between tree and buffer

#### Buffers
- `:e filename` - open file
- `:bn` or `:bnext` - next buffer
- `:bp` or `:bprev` - previous buffer
- `:ls` or `:buffers` - list buffers

#### File Operations
- `:w` - save file
- `:q` - quit
- `:wq` - save and quit

### Command Mode
- `:` - enter command mode
- `Esc` - cancel command
- `Enter` - execute command

## Configuration

Create `~/.config/voidabyss/init.lua`:

```lua
-- Display settings
vb = {}
vb.opt = {}
vb.opt.tabwidth = 4
vb.opt.number = true
vb.opt.relativenumber = false

-- Key mappings
keymap("n", "jj", "<Esc>")
keymap("n", "<C-s>", ":w<CR>")

-- Plugins
vb.plugins = {
    -- Add plugins here
}
```

See [CONFIG.md](docs/CONFIG.md) for full configuration options.

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/editor -v
go test ./internal/config -v
go test ./core/buffer -v
```

Current test coverage: **131 tests** across all packages.

## Architecture

- **core/buffer**: Piece table implementation for efficient text editing
- **internal/editor**: Main editor logic, modes, and input handling
- **internal/config**: Lua configuration loader and API
- **cmd/vb**: Entry point and CLI

## Project Status

✅ **Implemented:**
- Complete modal editing system
- Multi-buffer support with per-buffer state
- File tree navigation with smart ignore rules
- Lua configuration system
- Line numbers (absolute/relative)
- Text objects and visual selection
- Comprehensive test suite

🚧 **In Progress:**
- Key mapping integration with input handler
- Plugin RPC communication
- Autocmd system

📋 **Planned:**
- Syntax highlighting
- LSP support
- More text objects
- Split windows
- Tabs
- Colorscheme support

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

See LICENSE file for details.

## Author

DragonByte Labs
