# VoidAbyss Plugin System

VoidAbyss includes a powerful Lua-based plugin system that allows you to extend the editor with custom functionality. Plugins can respond to editor events, add commands, modify key mappings, and interact with buffers.

## Table of Contents

- [Quick Start](#quick-start)
- [Plugin Discovery](#plugin-discovery)
- [Plugin Structure](#plugin-structure)
- [Available Events](#available-events)
- [Lua API Reference](#lua-api-reference)
- [Example Plugins](#example-plugins)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Creating Your First Plugin

1. Create the plugins directory:
   ```bash
   mkdir -p ~/.config/voidabyss/plugins
   ```

2. Create a simple plugin `~/.config/voidabyss/plugins/hello.lua`:
   ```lua
   -- Simple hello world plugin
   vb.on("EditorReady", function()
       vb.notify("Hello from plugin!", "info")
   end)
   ```

3. Start VoidAbyss - the plugin will be auto-loaded and display a notification.

## Plugin Discovery

VoidAbyss automatically discovers and loads plugins in two ways:

### 1. Auto-Discovery (Recommended)

Place `.lua` files in `~/.config/voidabyss/plugins/`:
- All `.lua` files are automatically loaded on startup
- No configuration needed
- Plugins are isolated in their own namespace

### 2. Explicit Loading

Add plugins to your `init.lua`:
```lua
vb.plugins = {
    "plugin-name",              -- Loads ~/.config/voidabyss/plugins/plugin-name.lua
    "path/to/custom.lua",       -- Loads from config directory
    "~/absolute/path/plugin.lua" -- Loads from absolute path
}
```

## Plugin Structure

### Basic Plugin Template

```lua
-- Plugin description
local M = {}

-- Plugin configuration
M.config = {
    enabled = true,
    -- Add your config options here
}

-- Setup function (optional but recommended)
function M.setup(opts)
    -- Merge user config
    if opts then
        for k, v in pairs(opts) do
            M.config[k] = v
        end
    end
    
    -- Register event handlers
    vb.on("EditorReady", function()
        vb.notify("Plugin loaded!", "info")
    end)
    
    -- Add commands
    vb.command("MyCommand", function()
        -- Command implementation
    end, { desc = "My custom command" })
    
    -- Add keymaps
    vb.keymap("n", "<leader>mp", ":MyCommand<CR>", { desc = "My plugin command" })
end

-- Initialize with defaults
M.setup()

return M
```

## Available Events

VoidAbyss provides comprehensive event hooks for plugin integration:

### Editor Lifecycle
- **VimEnter** - Editor started and initialized
- **VimLeave** - Editor about to exit
- **EditorReady** - Editor ready for user input

### Buffer Events
- **BufEnter** - Entered a buffer (switched to it)
- **BufLeave** - Left a buffer (switching away)
- **BufNew** - New buffer created
- **BufRead** - Buffer loaded from file
- **BufDelete** - Buffer about to be deleted
- **BufWritePre** - Before writing buffer to file
- **BufWritePost** - After writing buffer to file

### Text Editing Events
- **TextChanged** - Text modified in normal/visual mode
- **TextChangedI** - Text modified in insert mode

### Mode Events
- **ModeChanged** - Editor mode changed (normal, insert, visual, etc.)
- **InsertEnter** - Entered insert mode
- **InsertLeave** - Left insert mode
- **VisualEnter** - Entered visual mode
- **VisualLeave** - Left visual mode

### Other Events
- **FileType** - File type detected/changed
- **SearchComplete** - Search operation completed

### Event Handler Examples

```lua
-- Simple event handler
vb.on("BufEnter", function()
    print("Entered buffer: " .. (vb.buf.filename or "unknown"))
end)

-- Event with context
vb.on("FileType", function(ctx)
    local filetype = ctx.filetype
    vb.notify("Detected filetype: " .. filetype, "info")
end)

-- One-time event
vb.on("EditorReady", function()
    -- This runs only once
end, { once = true })

-- Pattern-based event (advanced)
vb.on("BufRead", function()
    -- Only runs for .lua files
end, { pattern = "*.lua" })

-- Event groups (for cleanup)
vb.on("BufEnter", function()
    -- Your handler
end, { group = "MyPlugin" })

-- Later, remove all handlers in group
vb.clear_event_group("MyPlugin")
```

## Lua API Reference

### Core API

#### vb.notify(message, level)
Display a notification to the user.
- `message` (string): Message to display
- `level` (string): "info", "warn", or "error"

```lua
vb.notify("Operation complete", "info")
vb.notify("Warning: file not saved", "warn")
vb.notify("Error: operation failed", "error")
```

#### vb.cmd(command)
Execute an editor command.

```lua
vb.cmd("w")              -- Save file
vb.cmd("q")              -- Quit
vb.cmd("e filename.txt") -- Open file
```

#### vb.schedule(fn)
Schedule a function to run after current operations complete.

```lua
vb.schedule(function()
    vb.notify("Delayed notification", "info")
end)
```

### Buffer API

Access current buffer through `vb.buf`:

```lua
local buf = vb.buf

-- Buffer properties
buf.filename     -- Current file path
buf.dirty        -- true if modified
buf.filetype     -- Detected file type
buf.line_count   -- Number of lines
buf.index        -- Buffer index (1-based)
buf.total        -- Total number of buffers
```

### State Persistence

Persist data between editor sessions:

```lua
-- Save data
vb.state.set("my_plugin_data", { count = 42, enabled = true })

-- Load data (with default)
local data = vb.state.get("my_plugin_data", { count = 0, enabled = false })

-- Update count
data.count = data.count + 1
vb.state.set("my_plugin_data", data)
```

### Commands

Register custom commands:

```lua
vb.command("MyCommand", function(args)
    -- args is a table of arguments
    vb.notify("Command executed with " .. #args .. " args", "info")
end, { desc = "My custom command" })

-- Command with implementation
vb.command("HelloWorld", function()
    vb.notify("Hello, World!", "info")
end, { desc = "Greet the world" })
```

### Key Mappings

Add custom key mappings:

```lua
-- Basic keymap
vb.keymap("n", "<leader>h", ":HelloWorld<CR>", { desc = "hello world" })

-- Function keymap
vb.keymap("n", "<leader>t", function()
    vb.notify("Custom function called", "info")
end, { desc = "test function" })

-- Multiple modes
vb.keymap({"n", "v"}, "<leader>y", '"+y', { desc = "yank to clipboard" })
```

### Options

Access and modify editor options:

```lua
-- Get option
local tabwidth = vb.opt.tabwidth

-- Set option
vb.opt.tabwidth = 4
vb.opt.number = true
vb.opt.relativenumber = false
```

## Example Plugins

### 1. Auto-Save Plugin

Automatically saves files when leaving buffers:

```lua
local M = {}

M.config = { enabled = true, notify = true }

function M.auto_save()
    if not M.config.enabled or not vb.buf.dirty then
        return
    end
    
    vb.cmd("w")
    
    if M.config.notify then
        vb.notify("Auto-saved", "info")
    end
end

vb.on("BufLeave", function() M.auto_save() end)
vb.command("AutoSaveToggle", function()
    M.config.enabled = not M.config.enabled
    vb.notify("Auto-save " .. (M.config.enabled and "enabled" or "disabled"), "info")
end, { desc = "Toggle auto-save" })

return M
```

See `examples/auto-save.lua` for the full implementation.

### 2. Session Manager

Saves and restores open buffers:

```lua
local M = {}

function M.save_session()
    local files = {}
    if vb.buf.filename ~= "" then
        table.insert(files, vb.buf.filename)
    end
    vb.state.set("session_files", files)
    vb.notify("Session saved", "info")
end

function M.restore_session()
    local files = vb.state.get("session_files", {})
    for _, file in ipairs(files) do
        vb.cmd("e " .. file)
    end
    vb.notify("Session restored", "info")
end

vb.on("VimLeave", function() M.save_session() end)
vb.command("SessionRestore", function() M.restore_session() end, 
    { desc = "Restore previous session" })

return M
```

See `examples/session-manager.lua` for the full implementation.

### 3. Status Line Enhancement

Adds extra buffer information:

```lua
local M = {}

function M.get_buffer_stats()
    local buf = vb.buf
    local parts = {}
    
    if buf.filetype then
        table.insert(parts, buf.filetype)
    end
    
    table.insert(parts, string.format("%d lines", buf.line_count or 0))
    
    if buf.dirty then
        table.insert(parts, "[+]")
    end
    
    return table.concat(parts, " | ")
end

vb.on("BufEnter", function()
    vb.notify(M.get_buffer_stats(), "info")
end)

vb.command("BufInfo", function()
    vb.notify(M.get_buffer_stats(), "info")
end, { desc = "Show buffer info" })

return M
```

See `examples/status-enhanced.lua` for the full implementation.

## Best Practices

### 1. Use Modules

Always structure plugins as modules:
```lua
local M = {}
-- Plugin code
return M
```

### 2. Provide Configuration

Allow users to customize behavior:
```lua
M.config = {
    enabled = true,
    -- defaults
}

function M.setup(opts)
    if opts then
        for k, v in pairs(opts) do
            M.config[k] = v
        end
    end
end
```

### 3. Add Commands and Keymaps

Make features accessible:
```lua
vb.command("PluginFeature", function() M.do_something() end, 
    { desc = "Description" })
vb.keymap("n", "<leader>pf", ":PluginFeature<CR>", 
    { desc = "plugin feature" })
```

### 4. Use Event Groups

For easier cleanup:
```lua
vb.on("BufEnter", handler, { group = "MyPlugin" })
-- Later: vb.clear_event_group("MyPlugin")
```

### 5. Handle Errors

Protect against failures:
```lua
local ok, err = pcall(function()
    -- Risky operation
end)
if not ok then
    vb.notify("Error: " .. tostring(err), "error")
end
```

### 6. Debounce Frequent Events

Avoid performance issues:
```lua
local last_time = 0
vb.on("TextChanged", function()
    local now = os.time()
    if now - last_time < 1 then
        return  -- Skip if less than 1 second
    end
    last_time = now
    -- Do work
end)
```

### 7. Use Persistent State

For data that should survive restarts:
```lua
-- Save
vb.state.set("plugin_data", data)

-- Load with defaults
local data = vb.state.get("plugin_data", default_value)
```

## Troubleshooting

### Plugin Not Loading

1. Check file location: `~/.config/voidabyss/plugins/yourplugin.lua`
2. Check file extension: Must be `.lua`
3. Check for syntax errors: Run `lua yourplugin.lua` separately
4. Check logs: Look for error messages on startup

### Events Not Firing

1. Ensure event name is correct (case-sensitive)
2. Check if event is implemented (see Available Events)
3. Add debug notifications:
   ```lua
   vb.on("EventName", function()
       vb.notify("Event fired!", "info")
   end)
   ```

### State Not Persisting

1. Ensure using `vb.state.set()` not just local variables
2. Check state file: `~/.config/voidabyss/state.json`
3. Verify data is JSON-serializable (no functions, userdata)

### Command Not Found

1. Ensure command is registered with `vb.command()`
2. Check command name (case-sensitive)
3. Try `:checkhealth` to see registered commands

## Plugin Development Tips

### Debugging

Add debug mode to your plugin:
```lua
M.config.debug = false

function M.log(msg)
    if M.config.debug then
        vb.notify("[Plugin] " .. msg, "info")
    end
end
```

### Testing

Create a test command:
```lua
vb.command("PluginTest", function()
    M.run_tests()
end, { desc = "Test plugin functionality" })
```

### Documentation

Add help command:
```lua
vb.command("PluginHelp", function()
    print([[
Plugin Name - Description

Commands:
  :Command1 - Description
  :Command2 - Description

Keymaps:
  <leader>p1 - Feature 1
  <leader>p2 - Feature 2
]])
end, { desc = "Show plugin help" })
```

## Further Reading

- See `examples/` directory for complete plugin examples
- Check `init.lua` for configuration examples
- Read source code: `internal/config/lua_api.go` for API implementation
- Join discussions at [GitHub Discussions](https://github.com/dragonbytelabs/voidabyss)

## Contributing Plugins

Share your plugins with the community:
1. Create a GitHub repository for your plugin
2. Add documentation and examples
3. Submit to the VoidAbyss plugin directory (coming soon)
4. Share on GitHub Discussions

Happy plugin development! 🚀
