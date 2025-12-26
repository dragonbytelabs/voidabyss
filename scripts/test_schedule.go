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
-- Test vb.schedule
print("=== Testing vb.schedule ===\n")

-- Schedule a function
vb.schedule(function()
	print("✓ Scheduled function 1 executed")
end)

vb.schedule(function()
	print("✓ Scheduled function 2 executed")
end)

-- Nested schedule
vb.schedule(function()
	print("✓ Scheduled function 3 (will schedule another)")
	vb.schedule(function()
		print("  ✓ Nested scheduled function executed")
	end)
end)

print("✓ All functions scheduled\n")
`

	if err := loader.L.DoString(testConfig); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== Processing Scheduled Functions (Tick 1) ===")
	cfg, _ := loader.Load()
	cfg.ProcessScheduledFunctions(loader.L)

	fmt.Println("\n=== Processing Scheduled Functions (Tick 2 - for nested) ===")
	cfg.ProcessScheduledFunctions(loader.L)

	fmt.Println("\n=== Testing Error Handling ===")

	// Test error handling
	errorTest := `
vb.schedule(function()
	print("Before error")
	error("Intentional error for testing")
	print("After error (should not print)")
end)

vb.schedule(function()
	print("✓ This should still run after previous error")
end)
`

	if err := loader.L.DoString(errorTest); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	cfg.ProcessScheduledFunctions(loader.L)
}
