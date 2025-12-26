-- Status line enhancements plugin for VoidAbyss
-- Adds file type, line count, and buffer info to status line

local M = {}

M.config = {
	show_filetype = true,
	show_line_count = true,
	show_buffer_index = true,
	show_file_encoding = false,
}

-- Get current buffer stats
function M.get_buffer_stats()
	local buf = vb.buf
	if not buf then
		return ""
	end

	local parts = {}

	-- Buffer index
	if M.config.show_buffer_index then
		table.insert(parts, string.format("[%d/%d]", buf.index or 1, buf.total or 1))
	end

	-- File type
	if M.config.show_filetype and buf.filetype and buf.filetype ~= "" then
		table.insert(parts, buf.filetype)
	end

	-- Line count
	if M.config.show_line_count then
		local lines = buf.line_count or 0
		table.insert(parts, string.format("%d lines", lines))
	end

	-- Modified indicator
	if buf.dirty then
		table.insert(parts, "[+]")
	end

	return table.concat(parts, " | ")
end

-- Update status line
function M.update_status()
	local stats = M.get_buffer_stats()
	if stats ~= "" then
		-- This would integrate with status line rendering
		-- For now, we just track the info
		M.last_status = stats
	end
end

-- Setup function
function M.setup(opts)
	if opts then
		for k, v in pairs(opts) do
			M.config[k] = v
		end
	end

	-- Update status on various events
	vb.on("BufEnter", function()
		M.update_status()
	end)

	vb.on("TextChanged", function()
		M.update_status()
	end)

	vb.on("FileType", function()
		M.update_status()
	end)

	vb.notify("Status plugin loaded", "info")
end

-- Command to show current buffer stats
vb.command("BufInfo", function()
	local stats = M.get_buffer_stats()
	if stats ~= "" then
		vb.notify(stats, "info")
	else
		vb.notify("No buffer information available", "warn")
	end
end, { desc = "Show buffer information" })

M.setup()

return M
