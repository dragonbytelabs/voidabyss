package main

import (
	"fmt"

	"github.com/dragonbytelabs/voidabyss/internal/config"
)

func main() {
	fmt.Println("=== VoidAbyss Comprehensive API Test ===\n")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		fmt.Println("Using default config")
		cfg = config.DefaultConfig()
	}

	// Display version
	fmt.Printf("Version: %s\n\n", config.Version)

	// Display options
	fmt.Println("=== Options ===")
	fmt.Printf("Tab Width: %d\n", cfg.Options.TabWidth)
	fmt.Printf("Expand Tab: %v\n", cfg.Options.ExpandTab)
	fmt.Printf("Number: %v\n", cfg.Options.Number)
	fmt.Printf("Relative Number: %v\n", cfg.Options.RelativeNumber)
	fmt.Printf("Leader: %q\n", cfg.Options.Leader)
	fmt.Printf("Cursor Line: %v\n", cfg.Options.CursorLine)
	fmt.Printf("Wrap: %v\n", cfg.Options.Wrap)
	fmt.Printf("Scroll Off: %d\n", cfg.Options.ScrollOff)

	// Display keymaps
	fmt.Printf("\n=== Keymaps (%d) ===\n", len(cfg.KeyMappings))
	for i, km := range cfg.KeyMappings {
		fmt.Printf("%d. [%s] %s -> ", i+1, km.Mode, km.LHS)
		if km.IsFunc {
			fmt.Printf("<function> (%s)\n", km.Opts.Desc)
		} else {
			fmt.Printf("%s (%s)\n", km.RHS, km.Opts.Desc)
		}
	}

	// Display commands
	fmt.Printf("\n=== Commands (%d) ===\n", len(cfg.Commands))
	for i, cmd := range cfg.Commands {
		fmt.Printf("%d. :%s -> ", i+1, cmd.Name)
		if cmd.IsFunc {
			fmt.Printf("<function> (%s)\n", cmd.Opts.Desc)
		} else {
			fmt.Printf("%s (%s)\n", cmd.RHS, cmd.Opts.Desc)
		}
	}

	// Display event handlers
	fmt.Printf("\n=== Event Handlers (%d) ===\n", len(cfg.EventHandlers))
	for i, eh := range cfg.EventHandlers {
		fmt.Printf("%d. %s", i+1, eh.Event)
		if eh.Opts.Pattern != "" {
			fmt.Printf(" (pattern: %s)", eh.Opts.Pattern)
		}
		if eh.Opts.Once {
			fmt.Printf(" [once]")
		}
		fmt.Println()
	}

	// Display plugins
	fmt.Printf("\n=== Plugins (%d) ===\n", len(cfg.Plugins))
	for i, plugin := range cfg.Plugins {
		fmt.Printf("%d. %s\n", i+1, plugin)
	}

	// Display persistent state
	fmt.Printf("\n=== Persistent State ===\n")
	keys := cfg.State.Keys()
	fmt.Printf("Keys: %d\n", len(keys))
	for _, key := range keys {
		val := cfg.State.Get(key, nil)
		fmt.Printf("  %s = %v\n", key, val)
	}

	// Test features
	fmt.Printf("\n=== Feature Checks ===\n")
	features := []string{
		"keymap.leader",
		"events.bufwritepost",
		"opt.tabwidth",
		"buffer.api",
		"state.persistent",
		"nonexistent.feature",
	}
	for _, feature := range features {
		has := config.Features[feature]
		status := "❌"
		if has {
			status = "✅"
		}
		fmt.Printf("%s %s\n", status, feature)
	}

	fmt.Println("\n✅ Comprehensive API working!")
}
