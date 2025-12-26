package main

import (
	"fmt"

	"github.com/dragonbytelabs/voidabyss/internal/config"
)

func main() {
	loader := config.NewLoader()
	defer loader.Close()

	loader.SetupVbAPI()

	testConfig := `
-- Test leader expansion and keymap introspection
print("=== Testing Leader Expansion & Keymap Features ===\n")

-- Set leader key
vb.opt.leader = " "

-- Add keymaps with leader
vb.keymap("n", "<leader>w", ":write<CR>", { desc = "save file", noremap = true })
vb.keymap("n", "<leader>q", ":quit<CR>", { desc = "quit editor", noremap = true })
vb.keymap("n", "<leader>x", ":wq<CR>", { desc = "save and quit", noremap = false })

-- Add keymaps without leader
vb.keymap("i", "jj", "<Esc>", { desc = "exit insert", noremap = true })
vb.keymap("i", "jk", "<Esc>", { desc = "exit insert alt", noremap = true })

-- Add function keymap
vb.keymap("n", "<leader>h", function(ctx)
	print("Hello from function keymap!")
end, { desc = "say hello", noremap = true })

-- Test keymap.list() with no filters
print("✓ All keymaps:")
local all_maps = vb.keymap.list()
for _, map in ipairs(all_maps) do
	local noremap = map.noremap and "[noremap]" or "[remap]"
	local desc = map.desc ~= "" and " - " .. map.desc or ""
	print(string.format("  [%s] %s -> %s %s%s", map.mode, map.lhs, map.rhs, noremap, desc))
end

-- Test filtering by mode
print("\n✓ Normal mode keymaps only:")
local normal_maps = vb.keymap.list("n")
for _, map in ipairs(normal_maps) do
	print(string.format("  %s -> %s", map.lhs, map.rhs))
end

print("\n✓ Insert mode keymaps only:")
local insert_maps = vb.keymap.list("i")
for _, map in ipairs(insert_maps) do
	print(string.format("  %s -> %s", map.lhs, map.rhs))
end

-- Test filtering by lhs
print("\n✓ Keymap for ' w' (leader expanded):")
local specific_map = vb.keymap.list("n", " w")
for _, map in ipairs(specific_map) do
	print(string.format("  [%s] %s -> %s (%s)", map.mode, map.lhs, map.rhs, map.desc))
end

print("\n=== Testing Key Notation Precedence ===\n")

-- Add keymaps with different lengths
vb.keymap("n", "a", ":echo 'a'<CR>", { desc = "single a" })
vb.keymap("n", "ab", ":echo 'ab'<CR>", { desc = "double ab" })
vb.keymap("n", "abc", ":echo 'abc'<CR>", { desc = "triple abc" })

print("✓ Precedence test keymaps (longer should have higher precedence):")
local precedence_maps = vb.keymap.list("n")
for _, map in ipairs(precedence_maps) do
	if map.lhs == "a" or map.lhs == "ab" or map.lhs == "abc" then
		print(string.format("  %s -> %s", map.lhs, map.rhs))
	end
end

print("\n=== Testing Noremap vs Remap ===\n")

print("✓ Noremap keymaps (prevent recursive expansion):")
for _, map in ipairs(all_maps) do
	if map.noremap then
		print(string.format("  [%s] %s", map.mode, map.lhs))
	end
end

print("\n✓ Remap keymaps (allow recursive expansion):")
for _, map in ipairs(all_maps) do
	if not map.noremap then
		print(string.format("  [%s] %s", map.mode, map.lhs))
	end
end

print("\n✓ Test complete!")
`

	if err := loader.L.DoString(testConfig); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}
