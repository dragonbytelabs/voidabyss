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

	l.extractConfig()

	return l.config, nil
}

// SetupVbAPI sets up the vb table and functions in Lua
func (l *Loader) SetupVbAPI() {
	vbTable := l.L.NewTable()
	l.L.SetGlobal("vb", vbTable)

	optTable := l.L.NewTable()
	l.L.SetField(vbTable, "opt", optTable)

	pluginsTable := l.L.NewTable()
	l.L.SetField(vbTable, "plugins", pluginsTable)

	l.L.SetGlobal("keymap", l.L.NewFunction(l.luaMap))
}

// luaMap handles key mapping from Lua
func (l *Loader) luaMap(L *lua.LState) int {
	mode := L.CheckString(1)
	from := L.CheckString(2)
	to := L.CheckString(3)

	key := mode + ":" + from
	l.config.KeyMaps[key] = KeyMap{
		Mode: mode,
		From: from,
		To:   to,
	}

	return 0
}

// extractConfig extracts configuration values from Lua state
func (l *Loader) extractConfig() {
	vb := l.L.GetGlobal("vb")
	if vb == lua.LNil {
		return
	}

	vbTable, ok := vb.(*lua.LTable)
	if !ok {
		return
	}

	opt := l.L.GetField(vbTable, "opt")
	if optTable, ok := opt.(*lua.LTable); ok {
		if tabWidth := l.L.GetField(optTable, "tabwidth"); tabWidth.Type() == lua.LTNumber {
			l.config.TabWidth = int(lua.LVAsNumber(tabWidth))
		}

		if number := l.L.GetField(optTable, "number"); number.Type() == lua.LTBool {
			l.config.ShowLineNumbers = lua.LVAsBool(number)
		}

		if relNumber := l.L.GetField(optTable, "relativenumber"); relNumber.Type() == lua.LTBool {
			l.config.RelativeLineNums = lua.LVAsBool(relNumber)
		}
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
