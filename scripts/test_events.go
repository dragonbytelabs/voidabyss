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
print("=== Testing Enhanced Event System ===\n")

-- Test 1: Basic event with pattern
print("✓ Test 1: Pattern matching")
vb.on("BufRead", function(ctx, data)
	print("  Go file opened!")
end, { pattern = "*.go" })

vb.on("BufRead", function(ctx, data)
	print("  Lua file opened!")
end, { pattern = "*.lua" })

-- Test 2: Once flag
print("\n✓ Test 2: Once flag")
local once_count = 0
vb.on("EditorReady", function(ctx)
	once_count = once_count + 1
	print(string.format("  EditorReady fired (count: %d)", once_count))
end, { once = true })

-- Test 3: Event groups
print("\n✓ Test 3: Event groups with augroup")

vb.augroup("MyPlugin", function()
	vb.on("BufWritePre", function(ctx)
		print("  [MyPlugin] Pre-write handler 1")
	end)
	
	vb.on("BufWritePre", function(ctx)
		print("  [MyPlugin] Pre-write handler 2")
	end)
	
	vb.on("BufWritePost", function(ctx)
		print("  [MyPlugin] Post-write handler")
	end)
end)

vb.augroup("AnotherPlugin", function()
	vb.on("BufWritePre", function(ctx)
		print("  [AnotherPlugin] Pre-write handler")
	end)
end)

-- Test 4: Manual group assignment
print("\n✓ Test 4: Manual group assignment")
vb.on("ModeChanged", function(ctx)
	print("  Manual group handler")
end, { group = "ManualGroup" })

-- List handlers
print("\n=== Handler Summary ===")
`

	if err := loader.L.DoString(testConfig); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	cfg, _ := loader.Load()

	// Print handler statistics
	fmt.Println("\n✓ Event Handlers by Type:")
	eventCounts := make(map[string]int)
	for _, h := range cfg.EventHandlers {
		eventCounts[h.Event]++
	}
	for event, count := range eventCounts {
		fmt.Printf("  %s: %d handlers\n", event, count)
	}

	fmt.Println("\n✓ Event Handlers by Group:")
	groupInfo := cfg.EventGroupInfo()
	for group, count := range groupInfo {
		fmt.Printf("  %s: %d handlers\n", group, count)
	}

	// Test pattern matching
	fmt.Println("\n=== Testing Pattern Matching ===")
	testPatterns := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"*.go", "main.go", true},
		{"*.go", "test.lua", false},
		{"*.lua", "config.lua", true},
		{"src/*.go", "src/main.go", true},
		{"*.txt", "readme.md", false},
	}

	for _, test := range testPatterns {
		result := config.MatchPattern(test.pattern, test.path)
		status := "✓"
		if result != test.want {
			status = "✗"
		}
		fmt.Printf("%s Pattern %q matches %q: %v (expected %v)\n",
			status, test.pattern, test.path, result, test.want)
	}

	// Test event dispatching
	fmt.Println("\n=== Testing Event Dispatch ===")

	fmt.Println("\n✓ Dispatching BufRead event for 'main.go':")
	count := cfg.DispatchEvent("BufRead", map[string]interface{}{
		"path": "main.go",
	})
	fmt.Printf("  Executed %d handlers\n", count)

	fmt.Println("\n✓ Dispatching BufRead event for 'config.lua':")
	count = cfg.DispatchEvent("BufRead", map[string]interface{}{
		"path": "config.lua",
	})
	fmt.Printf("  Executed %d handlers\n", count)

	fmt.Println("\n✓ Dispatching BufRead event for 'readme.txt' (no match):")
	count = cfg.DispatchEvent("BufRead", map[string]interface{}{
		"path": "readme.txt",
	})
	fmt.Printf("  Executed %d handlers\n", count)

	// Test once flag
	fmt.Println("\n✓ Testing 'once' flag (dispatching EditorReady twice):")
	fmt.Println("  First dispatch:")
	count1 := cfg.DispatchEvent("EditorReady", map[string]interface{}{})
	fmt.Printf("    Executed %d handlers\n", count1)

	fmt.Println("  Second dispatch (should execute 0 due to once=true):")
	count2 := cfg.DispatchEvent("EditorReady", map[string]interface{}{})
	fmt.Printf("    Executed %d handlers\n", count2)

	// Test group management
	fmt.Println("\n✓ Testing group management:")
	fmt.Printf("  Handlers before clearing MyPlugin: %d\n", len(cfg.EventHandlers))
	cleared := cfg.ClearEventGroup("MyPlugin")
	fmt.Printf("  Cleared %d handlers from MyPlugin group\n", cleared)
	fmt.Printf("  Handlers after clearing: %d\n", len(cfg.EventHandlers))

	fmt.Println("\n✓ Remaining groups:")
	groupInfo = cfg.EventGroupInfo()
	for group, count := range groupInfo {
		fmt.Printf("  %s: %d handlers\n", group, count)
	}
}
