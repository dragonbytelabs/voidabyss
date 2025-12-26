package config

import (
	"path/filepath"
)

// EventManager handles event dispatching and filtering
type EventManager struct {
	handlers []EventHandler
}

// MatchPattern checks if a file path matches a pattern
// Supports simple glob patterns like "*.go", "src/**", etc.
func MatchPattern(pattern, path string) bool {
	if pattern == "" {
		return true // Empty pattern matches everything
	}

	// Simple glob matching
	matched, err := filepath.Match(pattern, filepath.Base(path))
	if err != nil {
		return false
	}

	// Also try matching full path
	if !matched {
		matched, _ = filepath.Match(pattern, path)
	}

	return matched
}

// DispatchEvent fires all matching event handlers for the given event
// Returns the number of handlers executed
func (c *Config) DispatchEvent(event string, data map[string]interface{}) int {
	path := ""
	if pathVal, ok := data["path"].(string); ok {
		path = pathVal
	}

	executed := 0
	remainingHandlers := []EventHandler{}

	for i := range c.EventHandlers {
		handler := &c.EventHandlers[i]

		// Skip if not the right event
		if handler.Event != event {
			remainingHandlers = append(remainingHandlers, *handler)
			continue
		}

		// Skip if already fired (for once=true)
		if handler.Opts.Once && handler.Fired {
			continue // Don't add to remaining (remove it)
		}

		// Check pattern match
		if handler.Opts.Pattern != "" && path != "" {
			if !MatchPattern(handler.Opts.Pattern, path) {
				remainingHandlers = append(remainingHandlers, *handler)
				continue
			}
		}

		// Execute handler
		// TODO: Call Lua function with data
		executed++

		// Mark as fired if once=true
		if handler.Opts.Once {
			handler.Fired = true
		}

		// Keep handler if not once, or if once but not yet fired
		if !handler.Opts.Once || !handler.Fired {
			remainingHandlers = append(remainingHandlers, *handler)
		}
	}

	// Update handlers list (remove fired once handlers)
	c.EventHandlers = remainingHandlers

	return executed
}

// ClearEventGroup removes all event handlers in a group
func (c *Config) ClearEventGroup(group string) int {
	if group == "" {
		return 0
	}

	cleared := 0
	remaining := []EventHandler{}

	for _, handler := range c.EventHandlers {
		if handler.Opts.Group == group {
			cleared++
		} else {
			remaining = append(remaining, handler)
		}
	}

	c.EventHandlers = remaining
	return cleared
}

// ListEventHandlers returns handlers filtered by event type and/or group
func (c *Config) ListEventHandlers(event, group string) []EventHandler {
	result := []EventHandler{}

	for _, handler := range c.EventHandlers {
		// Filter by event if specified
		if event != "" && handler.Event != event {
			continue
		}

		// Filter by group if specified
		if group != "" && handler.Opts.Group != group {
			continue
		}

		result = append(result, handler)
	}

	return result
}

// EventGroupInfo returns statistics about event handlers by group
func (c *Config) EventGroupInfo() map[string]int {
	groups := make(map[string]int)

	for _, handler := range c.EventHandlers {
		group := handler.Opts.Group
		if group == "" {
			group = "<default>"
		}
		groups[group]++
	}

	return groups
}

// GlobMatch performs simple glob matching (*, ?, [])
func GlobMatch(pattern, str string) bool {
	// Simple implementation - just use filepath.Match for now
	matched, _ := filepath.Match(pattern, str)
	return matched
}
