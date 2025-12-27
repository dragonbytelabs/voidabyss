-- LSP Plugin for VoidAbyss
-- This plugin provides Language Server Protocol support
-- Similar to Neovim's nvim-lspconfig plugin

local M = {}

-- Active LSP clients by filetype
M.clients = {}

-- Document state (tracks which files are opened in which clients)
M.documents = {}

-- Configuration for different language servers
M.servers = {
	gopls = {
		cmd = "gopls",
		args = {},
		filetypes = { "go" },
		root_patterns = { "go.mod", "go.sum", ".git" },
	},
	pyright = {
		cmd = "pyright-langserver",
		args = { "--stdio" },
		filetypes = { "python", "py" },
		root_patterns = { "pyproject.toml", "setup.py", ".git" },
	},
	rust_analyzer = {
		cmd = "rust-analyzer",
		args = {},
		filetypes = { "rust", "rs" },
		root_patterns = { "Cargo.toml", ".git" },
	},
	ts_ls = {
		cmd = "typescript-language-server",
		args = { "--stdio" },
		filetypes = { "typescript", "javascript", "ts", "js" },
		root_patterns = { "package.json", "tsconfig.json", ".git" },
	},
}

-- Helper to detect filetype from extension
local function get_filetype(filepath)
	local ext = filepath:match("%.([^%.]+)$")
	if not ext then
		return nil
	end
	
	-- Map extensions to filetypes
	local ext_map = {
		go = "go",
		py = "python",
		rs = "rust",
		ts = "typescript",
		js = "javascript",
	}
	
	return ext_map[ext]
end

-- Helper to find project root
local function find_root_dir(filepath, patterns)
	-- For now, just return the directory of the file
	-- TODO: Walk up directory tree looking for patterns
	return filepath:match("(.+)/[^/]+$") or "."
end

-- Start an LSP client for a server
local function start_client(server_name, config, root_dir)
	local client_id = server_name .. ":" .. root_dir
	
	-- Check if client already exists
	if M.clients[client_id] then
		return client_id
	end
	
	-- Start the client
	local success = pcall(function()
		lsp.start_client({
			cmd = config.cmd,
			args = config.args or {},
			root_dir = root_dir,
		})
	end)
	
	if success then
		M.clients[client_id] = {
			server_name = server_name,
			config = config,
			root_dir = root_dir,
		}
		vb.notify("LSP: Started " .. server_name, "info")
	else
		vb.notify("LSP: Failed to start " .. server_name, "error")
	end
	
	return client_id
end

-- Attach LSP to a buffer
local function attach_buffer(filepath, filetype)
	-- Find matching server
	local server_name, server_config
	for name, config in pairs(M.servers) do
		for _, ft in ipairs(config.filetypes) do
			if ft == filetype then
				server_name = name
				server_config = config
				break
			end
		end
		if server_name then
			break
		end
	end
	
	if not server_name then
		return nil
	end
	
	-- Find root directory
	local root_dir = find_root_dir(filepath, server_config.root_patterns)
	
	-- Start or get existing client
	local client_id = start_client(server_name, server_config, root_dir)
	
	-- Read file content
	local file = io.open(filepath, "r")
	if not file then
		return nil
	end
	local content = file:read("*all")
	file:close()
	
	-- Notify LSP of opened document
	lsp.did_open(client_id, filepath, content)
	
	-- Track document
	M.documents[filepath] = {
		client_id = client_id,
		version = 1,
	}
	
	return client_id
end

-- Setup function called from init.lua
function M.setup(opts)
	opts = opts or {}
	
	-- Merge user server configs with defaults
	if opts.servers then
		for name, config in pairs(opts.servers) do
			M.servers[name] = config
		end
	end
	
	-- Hook into buffer events
	events.on(events.BUFFER_OPENED, function(data)
		local filetype = get_filetype(data.filepath)
		if filetype then
			attach_buffer(data.filepath, filetype)
		end
	end)
	
	events.on(events.BUFFER_CHANGED, function(data)
		local doc = M.documents[data.filepath]
		if doc then
			doc.version = doc.version + 1
			lsp.did_change(doc.client_id, data.filepath, data.content)
		end
	end)
	
	events.on(events.BUFFER_SAVED, function(data)
		local doc = M.documents[data.filepath]
		if doc then
			lsp.did_save(doc.client_id, data.filepath)
		end
	end)
	
	events.on(events.BUFFER_CLOSED, function(data)
		local doc = M.documents[data.filepath]
		if doc then
			lsp.did_close(doc.client_id, data.filepath)
			M.documents[data.filepath] = nil
		end
	end)
	
	vb.notify("LSP plugin loaded", "info")
end

-- Public API for keybindings
function M.goto_definition()
	local ctx = vb.current_context()
	if not ctx then
		return
	end
	
	local filepath = ctx.file
	local line = ctx.line
	local col = ctx.col
	
	local doc = M.documents[filepath]
	if not doc then
		vb.notify("LSP not attached to this buffer", "warn")
		return
	end
	
	-- Request definition
	lsp.goto_definition(doc.client_id, filepath, line, col, function(locations)
		if not locations or #locations == 0 then
			vb.notify("No definition found", "info")
			return
		end
		
		-- Jump to first location
		local loc = locations[1]
		-- TODO: Actually open the file and jump to location
		vb.notify("Found definition at " .. loc.filepath .. ":" .. loc.line, "info")
	end)
end

function M.hover()
	local ctx = vb.current_context()
	if not ctx then
		return
	end
	
	local filepath = ctx.file
	local line = ctx.line
	local col = ctx.col
	
	local doc = M.documents[filepath]
	if not doc then
		vb.notify("LSP not attached to this buffer", "warn")
		return
	end
	
	lsp.hover(doc.client_id, filepath, line, col, function(info)
		if info then
			vb.notify("Hover: " .. info, "info")
		end
	end)
end

return M
