package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.TabWidth != 4 {
		t.Errorf("expected TabWidth=4, got %d", cfg.TabWidth)
	}

	if cfg.RelativeLineNums != false {
		t.Errorf("expected RelativeLineNums=false")
	}

	if cfg.ShowLineNumbers != true {
		t.Errorf("expected ShowLineNumbers=true")
	}
}

func TestGetConfigDir(t *testing.T) {
	dir := GetConfigDir()
	if dir == "" {
		t.Error("GetConfigDir returned empty string")
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if filepath.Base(path) != "init.lua" {
		t.Errorf("config path should end with init.lua, got %s", filepath.Base(path))
	}
}

func TestLuaConfigLoading(t *testing.T) {
	tmpDir := t.TempDir()
	testConfigPath := filepath.Join(tmpDir, "init.lua")

	testConfig := `
vb.opt.tabwidth = 8
vb.opt.number = false
vb.opt.relativenumber = true

vb.keymap("n", "jj", "<Esc>")

vb.plugins = {"plugin1"}
`

	if err := os.WriteFile(testConfigPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	loader := NewLoader()
	defer loader.Close()

	loader.SetupVbAPI()
	if err := loader.L.DoFile(testConfigPath); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	loader.ExtractLegacyPlugins()
	cfg := loader.config

	if cfg.Options.TabWidth != 8 {
		t.Errorf("expected TabWidth=8, got %d", cfg.Options.TabWidth)
	}

	if cfg.Options.Number != false {
		t.Error("expected Number=false")
	}

	if cfg.Options.RelativeNumber != true {
		t.Error("expected RelativeNumber=true")
	}

	if len(cfg.KeyMappings) != 1 {
		t.Errorf("expected 1 keymap, got %d", len(cfg.KeyMappings))
	}

	if len(cfg.Plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(cfg.Plugins))
	}
}
