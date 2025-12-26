package config

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// Config holds all editor configuration
type Config struct {
	Options       *Options
	KeyMappings   []KeyMapping
	Commands      []Command
	EventHandlers []EventHandler
	Plugins       []string
	LoadedPlugins []PluginInfo
	PluginDir     string
	State         *State

	// Scheduled functions (for vb.schedule)
	scheduledFns []*lua.LFunction
	scheduleMu   sync.Mutex

	// Legacy fields for backwards compatibility
	TabWidth         int
	RelativeLineNums bool
	ShowLineNumbers  bool
	KeyMaps          map[string]KeyMap // Deprecated, use KeyMappings
}

// KeyMap represents a key mapping (legacy)
type KeyMap struct {
	Mode string
	From string
	To   string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	state := NewState()
	state.Load() // Load existing state if available

	return &Config{
		Options:       DefaultOptions(),
		KeyMappings:   []KeyMapping{},
		Commands:      []Command{},
		EventHandlers: []EventHandler{},
		KeyMaps:       make(map[string]KeyMap),
		PluginDir:     getDefaultPluginDir(),
		Plugins:       []string{},
		State:         state,
		// Legacy fields
		TabWidth:         4,
		RelativeLineNums: false,
		ShowLineNumbers:  true,
	}
}

// GetConfigDir returns the OS-specific config directory
func GetConfigDir() string {
	var configDir string

	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config", "voidabyss")
	case "linux":
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config", "voidabyss")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		configDir = filepath.Join(appData, "voidabyss")
	default:
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".voidabyss")
	}

	return configDir
}

// GetConfigPath returns the path to init.lua
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "init.lua")
}

func getDefaultPluginDir() string {
	return filepath.Join(GetConfigDir(), "plugins")
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configDir := GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	pluginDir := filepath.Join(configDir, "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

	return nil
}

// CreateDefaultConfig creates a default init.lua if it doesn't exist
func CreateDefaultConfig() error {
	configPath := GetConfigPath()

	if _, err := os.Stat(configPath); err == nil {
		return nil
	}

	defaultConfig := `-- VoidAbyss Configuration File
-- See https://github.com/dragonbytelabs/voidabyss for full documentation

-- Options
vb.opt.tabwidth = 4
vb.opt.expandtab = false
vb.opt.number = true
vb.opt.relativenumber = false
vb.opt.leader = " "  -- Space as leader key

-- Keymaps
-- Format: vb.keymap(mode, lhs, rhs, { desc = "description" })
-- Modes: "n" (normal), "i" (insert), "v" (visual), "c" (command)

vb.keymap("n", "<leader>w", ":w<CR>", { desc = "save file" })
vb.keymap("n", "<leader>q", ":q<CR>", { desc = "quit" })
vb.keymap("i", "jj", "<Esc>", { desc = "exit insert mode" })

-- Commands
-- vb.command("W", ":w<CR>", { desc = "write file" })
vb.command("Checkhealth", function() vb.checkhealth() end, { desc = "check editor health" })
vb.command("Keymap", function(args)
	local mode = args[1]
	local lhs = args[2]
	local maps = vb.keymap.list(mode, lhs)
	
	if #maps == 0 then
		vb.notify("No keymaps found", "info")
		return
	end
	
	print("\n=== Keymaps ===")
	for _, map in ipairs(maps) do
		local noremap = map.noremap and "[noremap]" or "[remap]"
		local desc = map.desc ~= "" and " - " .. map.desc or ""
		print(string.format("  [%s] %s -> %s %s%s", map.mode, map.lhs, map.rhs, noremap, desc))
	end
	print("")
end, { desc = "list keymaps" })

-- Events
vb.on("EditorReady", function(ctx)
	vb.notify("VoidAbyss ready!", "info")
end)

-- Persistent state example
-- local count = vb.state.get("boot_count", 0)
-- vb.state.set("boot_count", count + 1)

-- Plugins
vb.plugins = {
	-- "username/plugin-name",
}
`

	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}
