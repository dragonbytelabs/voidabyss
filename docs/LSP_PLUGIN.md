# LSP Plugin Architecture

## Overview

VoidAbyss uses a plugin-based LSP architecture similar to Neovim, where LSP is not part of the core editor but instead runs as an independent plugin that attaches to buffers asynchronously.

## Why Plugin-Based?

**Problems with synchronous LSP:**
- Blocks the editor during initialization
- Server startup can take seconds (especially for large projects)
- Single LSP failure crashes the editor
- Tight coupling makes testing difficult

**Benefits of plugin architecture:**
- Editor never freezes or blocks
- LSP runs completely asynchronously
- User can choose which LSPs to enable
- Easy to add/remove/reload LSP without restarting
- Follows Neovim's proven design pattern

## Architecture

```
┌─────────────────────────────────────┐
│         Core Editor                 │
│   (No LSP dependencies)             │
│                                     │
│  • Buffer management                │
│  • Rendering                        │
│  • Input handling                   │
│  • Events (BufferOpened, etc.)      │
└──────────────┬──────────────────────┘
               │
               │ fires events
               ▼
┌─────────────────────────────────────┐
│      Event System (Lua)             │
│   internal/config/event_hooks.go    │
│                                     │
│  • EventBufferOpened                │
│  • EventBufferChanged               │
│  • EventBufferSaved                 │
│  • EventBufferClosed                │
└──────────────┬──────────────────────┘
               │
               │ triggers callbacks
               ▼
┌─────────────────────────────────────┐
│       LSP Plugin (Lua)              │
│   examples/lsp-plugin.lua           │
│                                     │
│  • Listens to buffer events         │
│  • Starts language servers          │
│  • Manages document sync            │
│  • Provides gd, gr, K commands      │
└──────────────┬──────────────────────┘
               │
               │ uses
               ▼
┌─────────────────────────────────────┐
│     LSP Client (Go)                 │
│   internal/lsp/client.go            │
│                                     │
│  • JSON-RPC communication           │
│  • Process management               │
│  • Async request/response           │
└─────────────────────────────────────┘
```

## Event System

The editor fires events that plugins can hook into:

```lua
-- Register event handler
events.on(events.BUFFER_OPENED, function(data)
    print("File opened:", data.filepath)
    print("Line:", data.line, "Col:", data.col)
end)
```

**Available Events:**
- `events.BUFFER_OPENED` - When a file is opened
- `events.BUFFER_CLOSED` - When a buffer is closed
- `events.BUFFER_CHANGED` - When buffer content changes
- `events.BUFFER_SAVED` - When buffer is saved to disk
- `events.CURSOR_MOVED` - When cursor position changes

**Event Data:**
```lua
{
    event = "BufferOpened",
    filepath = "/path/to/file.go",
    line = 10,       -- cursor line (0-based)
    col = 5,         -- cursor column (0-based)
    content = "..."  -- buffer content (for Changed event)
}
```

## LSP Lua API

The LSP Lua API provides functions for managing language servers:

### Client Management

```lua
-- Start an LSP client
client_id = lsp.start_client({
    cmd = "gopls",
    args = {},
    root_dir = "/path/to/project"
})

-- Stop an LSP client
lsp.stop_client(client_id)
```

### Document Synchronization

```lua
-- Notify LSP that document was opened
lsp.did_open(client_id, filepath, content)

-- Notify LSP of content changes
lsp.did_change(client_id, filepath, new_content)

-- Notify LSP of save
lsp.did_save(client_id, filepath)

-- Notify LSP of close
lsp.did_close(client_id, filepath)
```

### LSP Features

```lua
-- Go to definition
lsp.goto_definition(client_id, filepath, line, col, function(locations)
    if locations and #locations > 0 then
        local loc = locations[1]
        -- Jump to loc.filepath:loc.line:loc.col
    end
end)

-- Hover documentation
lsp.hover(client_id, filepath, line, col, function(info)
    if info then
        print("Hover:", info)
    end
end)
```

## Using the LSP Plugin

### 1. Copy the plugin to your config

```bash
cp examples/lsp-plugin.lua ~/.config/vb/lsp-plugin.lua
```

### 2. Load in your init.lua

```lua
-- Load the LSP plugin
local lsp = require('lsp-plugin')

-- Setup with desired language servers
lsp.setup({
    servers = {
        gopls = {
            cmd = "gopls",
            args = {},
            filetypes = { "go" },
            root_patterns = { "go.mod", ".git" },
        },
        pyright = {
            cmd = "pyright-langserver",
            args = { "--stdio" },
            filetypes = { "python", "py" },
            root_patterns = { "setup.py", ".git" },
        },
    },
})
```

### 3. Setup keybindings

```lua
-- Go to definition
vb.keymap.set('n', 'gd', function(ctx)
    lsp.goto_definition()
end)

-- Hover documentation
vb.keymap.set('n', 'K', function(ctx)
    lsp.hover()
end)
```

## How It Works

### 1. Buffer Opening

When you open a Go file:
1. Editor fires `EventBufferOpened` event
2. LSP plugin receives event
3. Plugin checks filetype (`*.go` → `gopls`)
4. Plugin starts gopls if not already running
5. Plugin calls `lsp.did_open()` with file content
6. LSP server analyzes the file in background

**User experience:** Editor opens instantly, LSP attaches asynchronously

### 2. Go-To-Definition

When you press `gd`:
1. Keybinding calls `lsp.goto_definition()`
2. Plugin checks if LSP is attached to current buffer
3. Plugin sends definition request to gopls
4. gopls responds with location (async, doesn't block)
5. Callback opens file and jumps to location

**User experience:** If LSP not ready, shows message. Otherwise jumps to definition.

### 3. File Changes

When you edit in insert mode:
1. Editor fires `EventBufferChanged` on mode exit
2. Plugin receives updated content
3. Plugin calls `lsp.did_change()` with new content
4. gopls re-analyzes file in background

**User experience:** Editor never blocks, changes sync in background

## Comparison to Neovim

### Neovim Approach

```lua
vim.api.nvim_create_autocmd("LspAttach", {
    callback = function(event)
        vim.keymap.set('n', 'gd', vim.lsp.buf.definition, { buffer = event.buf })
        vim.keymap.set('n', 'K', vim.lsp.buf.hover, { buffer = event.buf })
    end
})

require('lspconfig').gopls.setup({})
```

### VoidAbyss Approach

```lua
events.on(events.BUFFER_OPENED, function(data)
    -- Start LSP for buffer
end)

vb.keymap.set('n', 'gd', function(ctx)
    lsp.goto_definition()
end)

lsp.setup({ servers = { gopls = {} } })
```

**Key similarity:** Both use event-driven architecture where LSP attaches asynchronously

## Benefits

### For Users
- Editor starts instantly
- No freezing when opening files
- Can disable LSP per-filetype
- Easy to configure in Lua

### For Developers
- Core editor has no LSP dependencies
- LSP code is testable in isolation
- Can swap LSP implementations
- Follows separation of concerns

## Future Enhancements

- [ ] Auto-install language servers (like mason.nvim)
- [ ] Multiple LSP clients per buffer
- [ ] Diagnostics rendering in gutter
- [ ] Code actions menu
- [ ] Symbol search across workspace
- [ ] Semantic highlighting
- [ ] Inlay hints

## Migration Guide

If you were using the old built-in LSP (that froze the editor):

**Before:**
```
# LSP was automatic, but blocked editor
# gd would freeze while LSP initialized
```

**After:**
```lua
-- Add to ~/.config/vb/init.lua
local lsp = require('lsp-plugin')
lsp.setup({ servers = { gopls = {} } })

vb.keymap.set('n', 'gd', function() lsp.goto_definition() end)
```

**Changes:**
- LSP is now opt-in (add to your init.lua)
- Must setup keybindings explicitly
- Editor never freezes
- Shows "LSP not ready" if you use gd too quickly after opening file

## Troubleshooting

**"LSP not ready" message:**
- Wait 1-2 seconds after opening file for LSP to initialize
- Check that language server is installed (`which gopls`)
- Check notifications for error messages

**Language server not starting:**
- Verify server config in init.lua
- Check that server binary is in PATH
- Look for error messages in status line

**Definition not found:**
- Ensure go.mod exists (gopls needs it)
- Try saving file first (`:w`)
- Server might still be indexing project

## See Also

- [examples/lsp-plugin.lua](../examples/lsp-plugin.lua) - Full plugin implementation
- [examples/lsp-init-example.lua](../examples/lsp-init-example.lua) - Example configuration
- [internal/config/event_hooks.go](../internal/config/event_hooks.go) - Event system
- [internal/config/lua_lsp.go](../internal/config/lua_lsp.go) - LSP Lua API
- [internal/lsp/](../internal/lsp/) - LSP client implementation
