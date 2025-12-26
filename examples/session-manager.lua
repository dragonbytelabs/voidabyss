-- Session manager plugin for VoidAbyss
-- Saves and restores open buffers between sessions

local M = {}

M.config = {
	auto_save = true,
	auto_restore = false,
	session_file = nil,  -- Will default to ~/.config/voidabyss/session.json
}

-- Get default session file path
function M.get_session_file()
	if M.config.session_file then
		return M.config.session_file
	end
	-- Use vb.state for persistent storage
	return "session"  -- Will use vb.state.get/set
end

-- Save current session
function M.save_session()
	local buf = vb.buf
	if not buf then
		return
	end

	-- Get list of open files
	local files = {}
	-- For now, just save current file
	if buf.filename and buf.filename ~= "" and buf.filename ~= "[No Name]" then
		table.insert(files, buf.filename)
	end

	-- Save to state
	vb.state.set("session_files", files)
	vb.state.set("session_time", os.time())

	vb.notify(string.format("Session saved (%d files)", #files), "info")
end

-- Restore previous session
function M.restore_session()
	local files = vb.state.get("session_files")
	if not files or #files == 0 then
		vb.notify("No saved session found", "warn")
		return
	end

	local loaded = 0
	for _, file in ipairs(files) do
		local ok, err = pcall(function()
			vb.cmd("e " .. file)
		end)
		if ok then
			loaded = loaded + 1
		end
	end

	vb.notify(string.format("Restored %d files from session", loaded), "info")
end

-- Setup function
function M.setup(opts)
	if opts then
		for k, v in pairs(opts) do
			M.config[k] = v
		end
	end

	-- Auto-save session on exit
	if M.config.auto_save then
		vb.on("VimLeave", function()
			M.save_session()
		end)
	end

	-- Auto-restore on startup
	if M.config.auto_restore then
		vb.on("VimEnter", function()
			-- Delay restoration slightly
			vb.schedule(function()
				M.restore_session()
			end)
		end)
	end

	vb.notify("Session manager loaded", "info")
end

-- Commands
vb.command("SessionSave", function()
	M.save_session()
end, { desc = "Save current session" })

vb.command("SessionRestore", function()
	M.restore_session()
end, { desc = "Restore previous session" })

vb.command("SessionClear", function()
	vb.state.set("session_files", {})
	vb.notify("Session cleared", "info")
end, { desc = "Clear saved session" })

M.setup()

return M
