package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// Config holds all editor configuration
type Config struct {
	TabWidth         int
	RelativeLineNums bool
	ShowLineNumbers  bool
	KeyMaps          map[string]KeyMap
	PluginDir        string
	Plugins          []string
}

// KeyMap represents a key mapping
type KeyMap struct {
	Mode string
	From string
	To   string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		TabWidth:         4,
		RelativeLineNums: false,
		ShowLineNumbers:  true,
		KeyMaps:          make(map[string]KeyMap),
		PluginDir:        getDefaultPluginDir(),
		Plugins:          []string{},
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

	defaultConfig := `-- Voidabyss Configuration File
-- Tab width
vb = {}
vb.opt = {}
vb.opt.tabwidth = 4
vb.opt.number = true
vb.opt.relativenumber = false

-- Key mappings
-- keymap("n", "jj", "<Esc>")

-- Plugins
vb.plugins = {}

-- Hooks
function on_startup()
end

function on_file_open()
end
`

	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}
