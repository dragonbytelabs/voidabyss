package main

import (
	"fmt"

	"github.com/dragonbytelabs/voidabyss/internal/config"
)

func main() {
	// Create config loader with test configuration
	loader := config.NewLoader()
	defer loader.Close()

	// Setup API
	loader.SetupVbAPI()

	// Load some test config
	testConfig := `
-- Test Configuration
vb.opt.tabwidth = 2
vb.opt.leader = " "
vb.opt.number = true
vb.opt.relativenumber = true

-- Keymaps
vb.keymap("n", "<leader>w", ":w<CR>", { desc = "save file" })
vb.keymap("n", "<leader>q", ":q<CR>", { desc = "quit" })
vb.keymap("i", "jj", "<Esc>", { desc = "exit insert" })
vb.keymap("i", "<C-n>", function(ctx) ctx:complete_next() end, { desc = "next completion" })

-- Commands
vb.command("W", ":w<CR>", { desc = "write file" })
vb.command("Version", function() vb.notify("v" .. vb.version, "info") end, { desc = "show version" })

-- Events
vb.on("EditorReady", function(ctx)
	vb.notify("Ready!", "info")
end)

vb.on("BufWritePost", function(ctx, data)
	vb.notify("Saved!", "info")
end)

vb.on("ModeChanged", function(ctx, data)
	-- mode change handler
end)

-- State
local count = vb.state.get("boot_count", 0)
vb.state.set("boot_count", count + 1)
vb.state.set("test_key", "test_value")

-- Run checkhealth
vb.checkhealth()
`

	if err := loader.L.DoString(testConfig); err != nil {
		fmt.Printf("Error loading test config: %v\n", err)
		return
	}
}
