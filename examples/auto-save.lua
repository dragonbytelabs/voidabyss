-- Auto-save plugin for VoidAbyss
-- Automatically saves buffers when leaving them or after a period of inactivity

local M = {}

-- Configuration
M.config = {
	enabled = true,
	notify = true,  -- Show notification when auto-saving
	save_on_buf_leave = true,
	save_on_focus_lost = false,
}

-- Track last save time for debouncing
local last_save_time = {}

-- Auto-save function
function M.auto_save()
	if not M.config.enabled then
		return
	end

	-- Get current buffer info
	local buf = vb.buf
	if not buf then
		return
	end

	-- Don't save if buffer is not modified
	if not buf.dirty then
		return
	end

	-- Don't save unnamed buffers
	if buf.filename == "" or buf.filename == "[No Name]" then
		return
	end

	-- Debounce: don't save more than once per second
	local now = os.time()
	local last = last_save_time[buf.filename] or 0
	if now - last < 1 then
		return
	end

	-- Save the buffer
	local ok, err = pcall(function()
		vb.cmd("w")
	end)

	if ok then
		last_save_time[buf.filename] = now
		if M.config.notify then
			vb.notify("Auto-saved: " .. buf.filename, "info")
		end
	else
		vb.notify("Auto-save failed: " .. tostring(err), "error")
	end
end

-- Setup function - register event handlers
function M.setup(opts)
	-- Merge user config with defaults
	if opts then
		for k, v in pairs(opts) do
			M.config[k] = v
		end
	end

	-- Register BufLeave event to auto-save
	if M.config.save_on_buf_leave then
		vb.on("BufLeave", function()
			M.auto_save()
		end)
	end

	-- Register TextChanged event for periodic saves
	-- Note: This fires frequently, so we debounce in auto_save()
	vb.on("TextChanged", function()
		M.auto_save()
	end)

	vb.notify("Auto-save plugin loaded", "info")
end

-- Enable/disable commands
vb.command("AutoSaveEnable", function()
	M.config.enabled = true
	vb.notify("Auto-save enabled", "info")
end, { desc = "Enable auto-save" })

vb.command("AutoSaveDisable", function()
	M.config.enabled = false
	vb.notify("Auto-save disabled", "info")
end, { desc = "Disable auto-save" })

vb.command("AutoSaveToggle", function()
	M.config.enabled = not M.config.enabled
	local status = M.config.enabled and "enabled" or "disabled"
	vb.notify("Auto-save " .. status, "info")
end, { desc = "Toggle auto-save" })

-- Initialize with default config
M.setup()

return M
