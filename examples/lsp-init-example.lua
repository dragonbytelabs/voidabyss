-- Example init.lua with LSP plugin
--
-- To use LSP, add this to your ~/.config/vb/init.lua:

-- Load the LSP plugin
local lsp_plugin = require('lsp-plugin')

-- Setup LSP with default servers
lsp_plugin.setup({
	servers = {
		gopls = {
			cmd = "gopls",
			args = {},
			filetypes = { "go" },
			root_patterns = { "go.mod", "go.sum", ".git" },
		},
	},
})

-- Create keybindings for LSP functions
-- gd = go to definition
vb.keymap.set('n', 'gd', function(ctx)
	lsp_plugin.goto_definition()
end)

-- K = hover documentation
vb.keymap.set('n', 'K', function(ctx)
	lsp_plugin.hover()
end)

-- Example: Additional LSP keybindings (to be implemented)
-- gr = find references
-- vb.keymap.set('n', 'gr', function(ctx)
--     lsp_plugin.find_references()
-- end)

-- grn = rename symbol
-- vb.keymap.set('n', 'grn', function(ctx)
--     lsp_plugin.rename()
-- end)

print("LSP plugin example loaded")
