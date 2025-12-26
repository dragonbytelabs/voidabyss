package config

import (
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// TestHarness provides utilities for testing Lua configuration
type TestHarness struct {
	loader *Loader
	config *Config
	L      *lua.LState
}

// NewTestHarness creates a test harness for config testing
func NewTestHarness() *TestHarness {
	L := lua.NewState()
	cfg := DefaultConfig()

	loader := &Loader{
		L:             L,
		config:        cfg,
		Notifications: NewNotificationQueue(),
	}

	// Set up vb API
	loader.SetupVbAPI()

	return &TestHarness{
		loader: loader,
		config: cfg,
		L:      L,
	}
}

// LoadString loads and executes a Lua config string
func (h *TestHarness) LoadString(luaCode string) error {
	return h.L.DoString(luaCode)
}

// LoadFile loads and executes a Lua config file
func (h *TestHarness) LoadFile(path string) error {
	return h.L.DoFile(path)
}

// Close cleans up the test harness
func (h *TestHarness) Close() {
	if h.loader != nil {
		h.loader.Close()
	}
}

// GetOption returns an option value
func (h *TestHarness) GetOption(name string) interface{} {
	switch name {
	case "tabwidth":
		return h.config.Options.TabWidth
	case "expandtab":
		return h.config.Options.ExpandTab
	case "number":
		return h.config.Options.Number
	case "relativenumber":
		return h.config.Options.RelativeNumber
	case "leader":
		return h.config.Options.Leader
	default:
		return nil
	}
}

// GetKeymaps returns all keymaps for a mode
func (h *TestHarness) GetKeymaps(mode string) []KeyMapping {
	var result []KeyMapping
	for _, km := range h.config.KeyMappings {
		if km.Mode == mode {
			result = append(result, km)
		}
	}
	return result
}

// GetKeymap returns a specific keymap by mode and LHS
func (h *TestHarness) GetKeymap(mode, lhs string) *KeyMapping {
	for i := range h.config.KeyMappings {
		km := &h.config.KeyMappings[i]
		if km.Mode == mode && km.LHS == lhs {
			return km
		}
	}
	return nil
}

// GetCommands returns all registered commands
func (h *TestHarness) GetCommands() []Command {
	return h.config.Commands
}

// GetCommand returns a specific command by name
func (h *TestHarness) GetCommand(name string) *Command {
	for i := range h.config.Commands {
		cmd := &h.config.Commands[i]
		if cmd.Name == name {
			return cmd
		}
	}
	return nil
}

// GetEventHandlers returns all event handlers for an event
func (h *TestHarness) GetEventHandlers(event string) []EventHandler {
	var result []EventHandler
	for _, eh := range h.config.EventHandlers {
		if eh.Event == event {
			result = append(result, eh)
		}
	}
	return result
}

// GetState returns the state manager
func (h *TestHarness) GetState() *State {
	return h.config.State
}

// GetNotifications returns all notifications
func (h *TestHarness) GetNotifications() []Notification {
	var result []Notification
	for {
		notif := h.loader.Notifications.Pop()
		if notif == nil {
			break
		}
		result = append(result, *notif)
	}
	return result
}

// AssertOption checks an option value
func (h *TestHarness) AssertOption(t *testing.T, name string, expected interface{}) {
	t.Helper()
	got := h.GetOption(name)
	if got != expected {
		t.Errorf("Option %s = %v, want %v", name, got, expected)
	}
}

// AssertKeymapExists checks a keymap exists
func (h *TestHarness) AssertKeymapExists(t *testing.T, mode, lhs string) *KeyMapping {
	t.Helper()
	km := h.GetKeymap(mode, lhs)
	if km == nil {
		t.Errorf("Keymap %s %q not found", mode, lhs)
		return nil
	}
	return km
}

// AssertCommandExists checks a command exists
func (h *TestHarness) AssertCommandExists(t *testing.T, name string) *Command {
	t.Helper()
	cmd := h.GetCommand(name)
	if cmd == nil {
		t.Errorf("Command %q not found", name)
		return nil
	}
	return cmd
}

// AssertEventHandlerExists checks an event handler exists
func (h *TestHarness) AssertEventHandlerExists(t *testing.T, event string) bool {
	t.Helper()
	handlers := h.GetEventHandlers(event)
	if len(handlers) == 0 {
		t.Errorf("No event handlers for %q", event)
		return false
	}
	return true
}

// AssertStateValue checks a state value
func (h *TestHarness) AssertStateValue(t *testing.T, key string, expected interface{}) {
	t.Helper()
	got := h.GetState().Get(key, nil)
	if got != expected {
		t.Errorf("State[%s] = %v, want %v", key, got, expected)
	}
}

// Test examples

func TestHarness_Options(t *testing.T) {
	h := NewTestHarness()
	defer h.Close()

	err := h.LoadString(`
		vb.opt.tabwidth = 2
		vb.opt.expandtab = true
		vb.opt.number = false
		vb.opt.leader = ","
	`)
	if err != nil {
		t.Fatalf("LoadString failed: %v", err)
	}

	h.AssertOption(t, "tabwidth", 2)
	h.AssertOption(t, "expandtab", true)
	h.AssertOption(t, "number", false)
	h.AssertOption(t, "leader", ",")
}

func TestHarness_Keymaps(t *testing.T) {
	h := NewTestHarness()
	defer h.Close()

	err := h.LoadString(`
		vb.keymap("n", "dd", function(ctx)
			-- delete line
		end)
		
		vb.keymap("i", "jj", "<Esc>")
		
		vb.keymap("n", "<leader>w", ":w<CR>")
	`)
	if err != nil {
		t.Fatalf("LoadString failed: %v", err)
	}

	// Check function keymap
	km := h.AssertKeymapExists(t, "n", "dd")
	if km != nil && !km.IsFunc {
		t.Error("Expected dd to be a function keymap")
	}

	// Check string keymap
	km = h.AssertKeymapExists(t, "i", "jj")
	if km != nil {
		if km.IsFunc {
			t.Error("Expected jj to be a string keymap")
		}
		if km.RHS != "<Esc>" {
			t.Errorf("jj RHS = %q, want \"<Esc>\"", km.RHS)
		}
	}

	// Check leader keymap (should be expanded)
	km = h.AssertKeymapExists(t, "n", "\\w") // Default leader is backslash
	if km != nil {
		if km.RHS != ":w<CR>" {
			t.Errorf("<leader>w RHS = %q, want \":w<CR>\"", km.RHS)
		}
	}
}

func TestHarness_Commands(t *testing.T) {
	h := NewTestHarness()
	defer h.Close()

	err := h.LoadString(`
		vb.command("Hello", function(args, ctx)
			vb.notify("Hello " .. args)
		end)
	`)
	if err != nil {
		t.Fatalf("LoadString failed: %v", err)
	}

	cmd := h.AssertCommandExists(t, "Hello")
	if cmd == nil {
		return
	}

	// Execute the command
	err = h.loader.CallCommandFunction(cmd.Fn, "World")
	if err != nil {
		t.Errorf("CallCommandFunction failed: %v", err)
	}

	// Check notification
	notifs := h.GetNotifications()
	if len(notifs) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(notifs))
	}
	if !strings.Contains(notifs[0].Message, "Hello World") {
		t.Errorf("Notification = %q, want to contain \"Hello World\"", notifs[0].Message)
	}
}

func TestHarness_Events(t *testing.T) {
	h := NewTestHarness()
	defer h.Close()

	err := h.LoadString(`
		vb.on("BufWritePost", function(ctx, data)
			vb.notify("Saved: " .. (data.file or "unknown"))
		end)
	`)
	if err != nil {
		t.Fatalf("LoadString failed: %v", err)
	}

	h.AssertEventHandlerExists(t, "BufWritePost")

	// Fire the event
	handlers := h.GetEventHandlers("BufWritePost")
	if len(handlers) > 0 {
		eventData := map[string]interface{}{
			"file": "test.txt",
		}
		err = h.loader.CallEventHandler(handlers[0].Fn, eventData)
		if err != nil {
			t.Errorf("CallEventHandler failed: %v", err)
		}

		// Check notification
		notifs := h.GetNotifications()
		if len(notifs) != 1 {
			t.Fatalf("Expected 1 notification, got %d", len(notifs))
		}
		if !strings.Contains(notifs[0].Message, "Saved: test.txt") {
			t.Errorf("Notification = %q, want to contain \"Saved: test.txt\"", notifs[0].Message)
		}
	}
}

func TestHarness_State(t *testing.T) {
	h := NewTestHarness()
	defer h.Close()

	err := h.LoadString(`
		vb.state.set("last_file", "test.txt")
		vb.state.set("count", 42)
		vb.state.set("items", {1, 2, 3})
	`)
	if err != nil {
		t.Fatalf("LoadString failed: %v", err)
	}

	h.AssertStateValue(t, "last_file", "test.txt")
	h.AssertStateValue(t, "count", 42.0) // Numbers are float64 in JSON

	// Check array
	items := h.GetState().Get("items", nil)
	itemsArr, ok := items.([]interface{})
	if !ok {
		t.Fatalf("items is not []interface{}, got %T", items)
	}
	if len(itemsArr) != 3 {
		t.Errorf("len(items) = %d, want 3", len(itemsArr))
	}
}

func TestHarness_Schedule(t *testing.T) {
	h := NewTestHarness()
	defer h.Close()

	err := h.LoadString(`
		vb.schedule(function()
			vb.notify("Scheduled task executed")
		end)
	`)
	if err != nil {
		t.Fatalf("LoadString failed: %v", err)
	}

	// Check scheduled functions
	h.config.scheduleMu.Lock()
	count := len(h.config.scheduledFns)
	h.config.scheduleMu.Unlock()

	if count != 1 {
		t.Errorf("Expected 1 scheduled function, got %d", count)
	}

	// Process scheduled functions
	h.config.ProcessScheduledFunctions(h.L)

	// Check notification
	notifs := h.GetNotifications()
	if len(notifs) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(notifs))
	}
	if !strings.Contains(notifs[0].Message, "Scheduled task executed") {
		t.Errorf("Notification = %q, want to contain \"Scheduled task executed\"", notifs[0].Message)
	}
}
