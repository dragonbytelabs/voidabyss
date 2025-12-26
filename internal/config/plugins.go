package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// PluginInfo holds information about a loaded plugin
type PluginInfo struct {
	Name      string      // Plugin name/identifier
	Path      string      // Full path to plugin file
	Namespace *lua.LTable // Plugin's isolated namespace
	Loaded    bool        // Whether plugin loaded successfully
	Error     string      // Error message if load failed
}

// LoadPlugins loads all plugins from the vb.plugins table
// Each plugin gets its own namespace to avoid global pollution
func (l *Loader) LoadPlugins() []PluginInfo {
	l.ExtractLegacyPlugins()

	if len(l.config.Plugins) == 0 {
		return []PluginInfo{}
	}

	results := make([]PluginInfo, 0, len(l.config.Plugins))

	for _, pluginSpec := range l.config.Plugins {
		info := l.loadSinglePlugin(pluginSpec)
		results = append(results, info)
	}

	return results
}

// loadSinglePlugin loads a single plugin with namespace isolation
func (l *Loader) loadSinglePlugin(spec string) PluginInfo {
	// Parse plugin spec (can be "name" or "path/to/plugin.lua")
	pluginPath := l.resolvePluginPath(spec)

	info := PluginInfo{
		Name:   spec,
		Path:   pluginPath,
		Loaded: false,
	}

	// Check if file exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		info.Error = fmt.Sprintf("plugin file not found: %s", pluginPath)
		return info
	}

	// Create isolated namespace for plugin
	namespace := l.L.NewTable()

	// Give plugin access to vb API
	vb := l.L.GetGlobal("vb")
	l.L.SetField(namespace, "vb", vb)

	// Give plugin access to standard Lua libraries
	l.L.SetField(namespace, "print", l.L.GetGlobal("print"))
	l.L.SetField(namespace, "type", l.L.GetGlobal("type"))
	l.L.SetField(namespace, "tostring", l.L.GetGlobal("tostring"))
	l.L.SetField(namespace, "tonumber", l.L.GetGlobal("tonumber"))
	l.L.SetField(namespace, "ipairs", l.L.GetGlobal("ipairs"))
	l.L.SetField(namespace, "pairs", l.L.GetGlobal("pairs"))
	l.L.SetField(namespace, "next", l.L.GetGlobal("next"))
	l.L.SetField(namespace, "select", l.L.GetGlobal("select"))
	l.L.SetField(namespace, "string", l.L.GetGlobal("string"))
	l.L.SetField(namespace, "table", l.L.GetGlobal("table"))
	l.L.SetField(namespace, "math", l.L.GetGlobal("math"))
	l.L.SetField(namespace, "os", l.L.GetGlobal("os"))

	// Store namespace
	info.Namespace = namespace

	// Load plugin in protected mode
	if err := l.loadPluginFile(pluginPath, namespace); err != nil {
		info.Error = err.Error()
		return info
	}

	info.Loaded = true
	return info
}

// loadPluginFile loads a plugin file into a namespace
func (l *Loader) loadPluginFile(path string, namespace *lua.LTable) error {
	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read plugin: %w", err)
	}

	// Create a function that sets the environment to namespace
	chunk := fmt.Sprintf(`
local function load_plugin_chunk()
	%s
end
return load_plugin_chunk
`, string(content))

	// Compile the chunk
	fn, err := l.L.LoadString(chunk)
	if err != nil {
		return fmt.Errorf("failed to compile plugin: %w", err)
	}

	// Call to get the wrapped function
	l.L.Push(fn)
	if err := l.L.PCall(0, 1, nil); err != nil {
		return fmt.Errorf("failed to prepare plugin: %w", err)
	}

	// Get the returned function
	wrappedFn := l.L.Get(-1)
	l.L.Pop(1)

	if wrappedFn.Type() != lua.LTFunction {
		return fmt.Errorf("plugin compilation failed")
	}

	// Set up environment for the function
	// Note: In Lua 5.1 (which gopher-lua implements), we need to use setfenv
	l.L.SetFEnv(wrappedFn.(*lua.LFunction), namespace)

	// Execute the plugin in protected mode
	l.L.Push(wrappedFn)
	if err := l.L.PCall(0, 0, nil); err != nil {
		return fmt.Errorf("plugin execution error: %w", err)
	}

	return nil
}

// resolvePluginPath resolves a plugin spec to a full path
func (l *Loader) resolvePluginPath(spec string) string {
	// If it's already a path (contains / or ends with .lua), use it directly
	if strings.Contains(spec, "/") || strings.HasSuffix(spec, ".lua") {
		// Expand ~ to home directory
		if strings.HasPrefix(spec, "~/") {
			home, _ := os.UserHomeDir()
			spec = filepath.Join(home, spec[2:])
		}

		// If absolute, return as-is
		if filepath.IsAbs(spec) {
			return spec
		}

		// Otherwise, relative to config dir
		configDir := GetConfigDir()
		return filepath.Join(configDir, spec)
	}

	// Otherwise, look in standard plugin directory
	// ~/.config/voidabyss/plugins/<name>.lua
	configDir := GetConfigDir()
	pluginDir := filepath.Join(configDir, "plugins")
	return filepath.Join(pluginDir, spec+".lua")
}

// GetPluginDir returns the standard plugin directory
func GetPluginDir() string {
	configDir := GetConfigDir()
	return filepath.Join(configDir, "plugins")
}

// EnsurePluginDir creates the plugin directory if it doesn't exist
func EnsurePluginDir() error {
	pluginDir := GetPluginDir()
	return os.MkdirAll(pluginDir, 0755)
}

// DiscoverPlugins finds all .lua files in the plugin directory
// and adds them to the config if not already present
func (l *Loader) DiscoverPlugins() error {
	pluginDir := GetPluginDir()

	// Ensure directory exists
	if err := EnsurePluginDir(); err != nil {
		return err
	}

	// Read directory
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	// Track existing plugins to avoid duplicates
	existing := make(map[string]bool)
	for _, p := range l.config.Plugins {
		existing[p] = true
	}

	// Add any .lua files not already in the list
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".lua") {
			continue
		}

		// Remove .lua extension for plugin name
		pluginName := strings.TrimSuffix(name, ".lua")

		// Skip if already in list
		if existing[pluginName] {
			continue
		}

		// Add to plugins list
		l.config.Plugins = append(l.config.Plugins, pluginName)
	}

	return nil
}
