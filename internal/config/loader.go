package config

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// Loader handles loading and parsing Lua configuration
type Loader struct {
	L      *lua.LState
	config *Config
}

// NewLoader creates a new config loader
func NewLoader() *Loader {
	return &Loader{
		L:      lua.NewState(),
		config: DefaultConfig(),
	}
}

// Close closes the Lua state
func (l *Loader) Close() {
	l.L.Close()
}

// Load loads the configuration from init.lua
func (l *Loader) Load() (*Config, error) {
	configPath := GetConfigPath()

	l.SetupVbAPI()

	if err := l.L.DoFile(configPath); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	l.ExtractLegacyPlugins()

	// Load plugins after config
	pluginResults := l.LoadPlugins()
	l.config.LoadedPlugins = pluginResults

	return l.config, nil
}

// ExtractLegacyPlugins extracts plugins from vb.plugins table for backwards compatibility
func (l *Loader) ExtractLegacyPlugins() {
	vb := l.L.GetGlobal("vb")
	if vb == lua.LNil {
		return
	}

	vbTable, ok := vb.(*lua.LTable)
	if !ok {
		return
	}

	plugins := l.L.GetField(vbTable, "plugins")
	if pluginsTable, ok := plugins.(*lua.LTable); ok {
		pluginsTable.ForEach(func(key, value lua.LValue) {
			if str, ok := value.(lua.LString); ok {
				l.config.Plugins = append(l.config.Plugins, string(str))
			}
		})
	}
}

// LoadConfig is a convenience function to load configuration
func LoadConfig() (*Config, error) {
	if err := EnsureConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	if err := CreateDefaultConfig(); err != nil {
		return nil, fmt.Errorf("failed to create default config: %w", err)
	}

	loader := NewLoader()
	defer loader.Close()

	config, err := loader.Load()
	if err != nil {
		return DefaultConfig(), err
	}

	return config, nil
}
