-- Voidabyss Complete API Example
-- This file demonstrates all available configuration options

-- ============================================================================
-- OPTIONS (vb.opt)
-- ============================================================================

-- Core/Editor settings
vb.opt.tabwidth = 4              -- Number of spaces per tab
vb.opt.expandtab = false         -- Use spaces instead of tabs
vb.opt.cursorline = false        -- Highlight current line
vb.opt.relativenumber = false    -- Show relative line numbers
vb.opt.number = true             -- Show absolute line numbers
vb.opt.wrap = true               -- Wrap long lines
vb.opt.scrolloff = 3             -- Lines to keep above/below cursor
vb.opt.leader = " "              -- Leader key (space is popular)

-- Alternative syntax using methods:
-- vb.opt:set("tabwidth", 4)
-- local tw = vb.opt:get("tabwidth")

-- ============================================================================
-- KEYMAPS (vb.keymap)
-- ============================================================================

-- Format: vb.keymap(mode, lhs, rhs, opts)
-- Modes: "n" (normal), "i" (insert), "v" (visual), "c" (command)

-- Save and quit
vb.keymap("n", "<leader>w", ":w<CR>", { desc = "save file" })
vb.keymap("n", "<leader>q", ":q<CR>", { desc = "quit" })
vb.keymap("n", "<leader>x", ":wq<CR>", { desc = "save and quit" })

-- Fast escape from insert mode
vb.keymap("i", "jj", "<Esc>", { desc = "exit insert mode" })
vb.keymap("i", "jk", "<Esc>", { desc = "exit insert mode (alt)" })

-- File tree
vb.keymap("n", "<leader>e", ":Explore<CR>", { desc = "toggle file tree" })
vb.keymap("n", "<leader>f", "/", { desc = "search" })

-- Buffer navigation
vb.keymap("n", "<leader>n", ":bnext<CR>", { desc = "next buffer" })
vb.keymap("n", "<leader>p", ":bprev<CR>", { desc = "previous buffer" })
vb.keymap("n", "<leader>l", ":ls<CR>", { desc = "list buffers" })

-- Completion (with function callbacks)
vb.keymap("i", "<C-n>", function(ctx)
	ctx:complete_next()
end, { desc = "next completion" })

vb.keymap("i", "<C-p>", function(ctx)
	ctx:complete_prev()
end, { desc = "previous completion" })

vb.keymap("i", "<C-e>", function(ctx)
	ctx:complete_cancel()
end, { desc = "cancel completion" })

-- ============================================================================
-- COMMANDS (vb.command)
-- ============================================================================

-- Custom commands
vb.command("W", ":w<CR>", { desc = "write file (uppercase)" })
vb.command("Q", ":q<CR>", { desc = "quit (uppercase)" })

-- Command with function callback
vb.command("Echo", function(args, ctx)
	local msg = args[1] or "Hello!"
	vb.notify(msg, "info")
end, { nargs = "?", desc = "echo message" })

vb.command("Version", function(args, ctx)
	vb.notify("Voidabyss version: " .. vb.version, "info")
end, { desc = "show version" })

-- ============================================================================
-- EVENTS / AUTOCMDS (vb.on)
-- ============================================================================

-- Editor ready event
vb.on("EditorReady", function(ctx)
	local leader = vb.opt:get("leader")
	vb.notify("Voidabyss " .. vb.version .. " ready! (leader: " .. leader .. ")", "info")
end)

-- File write events
vb.on("BufWritePre", function(ctx, data)
	-- You could add auto-formatting here
	vb.notify("Saving...", "info")
end)

vb.on("BufWritePost", function(ctx, data)
	vb.notify("Saved: " .. (data.path or "unknown"), "info")
end)

-- File open event
vb.on("BufRead", function(ctx, data)
	vb.notify("Opened: " .. (data.path or "unknown"), "info")
end)

-- Mode change event
vb.on("ModeChanged", function(ctx, data)
	-- Optional: notify on mode changes
	-- vb.notify("Mode: " .. data.mode, "info")
end)

-- ============================================================================
-- PERSISTENT STATE (vb.state)
-- ============================================================================

-- Track boot count
local boot_count = vb.state.get("boot_count", 0)
vb.state.set("boot_count", boot_count + 1)

-- Store last opened files (example)
local recent_files = vb.state.get("recent_files", {})
-- vb.state.set("recent_files", recent_files)

-- View all state keys
local all_keys = vb.state.keys()
if #all_keys > 0 then
	vb.notify("State keys: " .. table.concat(all_keys, ", "), "info")
end

-- Delete state key (example)
-- vb.state.del("some_key")

-- ============================================================================
-- BUFFER API (vb.buf)
-- ============================================================================

-- Note: Buffer operations are context-dependent and work best in
-- event handlers or commands where there's an active buffer

-- Example: Auto-trim trailing whitespace on save
vb.on("BufWritePre", function(ctx, data)
	-- This is a stub - actual implementation needs editor integration
	-- local text = vb.buf.get_text()
	-- local trimmed = text:gsub("%s+\n", "\n")
	-- vb.buf.set_text(trimmed)
end)

-- ============================================================================
-- FEATURE DETECTION (vb.has)
-- ============================================================================

-- Check if features are available before using them
if vb.has("keymap.leader") then
	vb.notify("Leader key mappings supported", "info")
end

if vb.has("state.persistent") then
	vb.notify("Persistent state supported", "info")
end

if vb.has("buffer.api") then
	vb.notify("Buffer API available", "info")
end

-- Check for non-existent feature
if not vb.has("syntax.highlighting") then
	-- vb.notify("Syntax highlighting not yet available", "warn")
end

-- ============================================================================
-- PLUGINS
-- ============================================================================

-- List of plugins to load (future: RPC communication)
vb.plugins = {
	-- "username/plugin-name",
	-- "another/plugin",
}

-- ============================================================================
-- CUSTOM FUNCTIONS
-- ============================================================================

-- Helper function example
function print_stats()
	local stats = {
		"Voidabyss " .. vb.version,
		"Tab width: " .. vb.opt:get("tabwidth"),
		"Boot count: " .. vb.state.get("boot_count", 0),
	}
	
	for _, stat in ipairs(stats) do
		vb.notify(stat, "info")
	end
end

-- You can call custom functions from keymaps
vb.keymap("n", "<leader>s", function(ctx)
	print_stats()
end, { desc = "show stats" })

-- ============================================================================
-- COMPLETION
-- ============================================================================

vb.notify("Config loaded successfully!", "info")
