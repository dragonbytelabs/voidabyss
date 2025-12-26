# Voidabyss Configuration System

The configuration system uses Lua (via gopher-lua) to provide a flexible and familiar configuration experience similar to Neovim.

## Configuration File

The configuration file is located at:
- **macOS/Linux**: `~/.config/voidabyss/init.lua`
- **Windows**: `%APPDATA%\voidabyss\init.lua`

The config directory and a default `init.lua` are automatically created on first run.

## Configuration Options

### Display Settings

```lua
-- Tab width (number of spaces per tab)
vb.opt.tabwidth = 4

-- Show absolute line numbers
vb.opt.number = true

-- Show relative line numbers (relative to cursor position)
vb.opt.relativenumber = false
```

### Key Mappings

Use the `keymap()` function to define custom key mappings:

```lua
-- Format: map(mode, from, to)
-- Modes:
--   "n" = normal mode
--   "i" = insert mode  
--   "v" = visual mode
--   "c" = command mode

-- Example mappings:
keymap("n", "jj", "<Esc>")         -- Exit insert mode with jj
keymap("n", "<C-s>", ":w<CR>")     -- Save with Ctrl+S
keymap("n", "<leader>e", ":Explore<CR>")  -- Open file tree
```

### Plugins

The plugin system uses RPC to communicate with the editor:

```lua
-- List of plugins to load
vb.plugins = {
    "username/plugin-name",
    "another/plugin"
}
```

### Hooks

Define custom functions that run at specific times:

```lua
-- Called when the editor starts
function on_startup()
    print("Voidabyss started!")
end

-- Called when a file is opened
function on_file_open()
    print("File opened!")
end
```

## Example Configuration

Here's a complete example `init.lua`:

```lua
-- Voidabyss Configuration

-- Display settings
vb = {}
vb.opt = {}
vb.opt.tabwidth = 4
vb.opt.number = true
vb.opt.relativenumber = false

-- Key mappings
keymap("n", "jj", "<Esc>")
keymap("n", "<C-s>", ":w<CR>")
keymap("n", "<leader>e", ":Explore<CR>")
keymap("n", "<leader>f", "/")

-- Plugins
vb.plugins = {
    -- Add your plugins here
}

-- Custom functions
function on_startup()
    -- Initialization code
end

function on_file_open()
    -- Per-file setup
end
```

## Testing Your Configuration

After editing your `init.lua`, restart Voidabyss. If there are any errors in your configuration, the editor will fall back to default settings and continue to work.

## Current Features

✅ **Implemented:**
- Tab width configuration
- Line number display (absolute/relative)
- Custom key mappings (framework ready, integration pending)
- Plugin loading mechanism (framework ready)
- Lua API for configuration

🚧 **In Progress:**
- Key mapping integration with input handler
- Plugin RPC communication
- Autocmd system

## Next Steps

The configuration system is now ready for:
1. **Key mapping integration**: Apply custom keymaps in the input handler
2. **Plugin system**: Full RPC communication between plugins and editor
3. **Autocmds**: Event-driven automation (BufEnter, BufWrite, etc.)
4. **More options**: Colorschemes, statusline customization, etc.
