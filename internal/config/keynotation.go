package config

import (
	"strings"
)

// KeyNotation handles parsing and expansion of vim-style key notation
// Supports: <C-x>, <S-x>, <M-x>, <A-x>, <CR>, <Esc>, <Tab>, <BS>, <Space>, <leader>, etc.

// ExpandLeader replaces <leader> with the actual leader key
func ExpandLeader(key, leader string) string {
	// Case insensitive replacement
	result := strings.ReplaceAll(key, "<leader>", leader)
	result = strings.ReplaceAll(result, "<Leader>", leader)
	result = strings.ReplaceAll(result, "<LEADER>", leader)
	return result
}

// ParseKeyNotation converts vim-style key notation to internal representation
// For now, this is a simple parser. Future: handle <C-x>, <S-x>, etc.
func ParseKeyNotation(key string) string {
	// Common special keys
	replacements := map[string]string{
		"<CR>":     "\r",
		"<LF>":     "\n",
		"<Esc>":    "\x1b",
		"<Tab>":    "\t",
		"<BS>":     "\x08",
		"<Del>":    "\x7f",
		"<Space>":  " ",
		"<Bar>":    "|",
		"<Bslash>": "\\",
		"<lt>":     "<",
	}

	result := key
	for notation, replacement := range replacements {
		result = strings.ReplaceAll(result, notation, replacement)
	}

	return result
}

// IsPrefix checks if 'prefix' is a prefix of 'full'
func IsPrefix(prefix, full string) bool {
	return strings.HasPrefix(full, prefix)
}

// KeyPrecedence determines the priority of a keymap
// Higher numbers = higher priority
// Rules:
//  1. Exact match beats prefix
//  2. Longer LHS beats shorter
//  3. User mappings override defaults (via order)
//  4. Later mappings override earlier (via order)
func KeyPrecedence(lhs string) int {
	// Base score on length (longer = more specific)
	score := len(lhs) * 100

	// Bonus for special keys (more intentional)
	if strings.Contains(lhs, "<") {
		score += 50
	}

	return score
}

// KeyMappingsByPrecedence sorts keymaps by precedence
// Returns a copy sorted by precedence (highest first)
func KeyMappingsByPrecedence(mappings []KeyMapping, mode string) []KeyMapping {
	// Filter by mode
	filtered := []KeyMapping{}
	for _, km := range mappings {
		if km.Mode == mode {
			filtered = append(filtered, km)
		}
	}

	// Sort by precedence (stable sort - later mappings win on ties)
	// Using a simple bubble sort to maintain order
	sorted := make([]KeyMapping, len(filtered))
	copy(sorted, filtered)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			scoreI := KeyPrecedence(sorted[i].LHS)
			scoreJ := KeyPrecedence(sorted[j].LHS)

			// Higher score should come first
			// On tie, preserve original order (later wins)
			if scoreJ > scoreI {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// FindKeyMapping finds the best matching keymap for the given keys
// Returns the mapping and whether it's an exact match or prefix
func FindKeyMapping(keys string, mappings []KeyMapping, mode string) (*KeyMapping, bool, bool) {
	sorted := KeyMappingsByPrecedence(mappings, mode)

	var exactMatch *KeyMapping
	var prefixMatch *KeyMapping

	for i := range sorted {
		km := &sorted[i]
		if km.LHS == keys {
			exactMatch = km
			break
		}
		if IsPrefix(keys, km.LHS) {
			if prefixMatch == nil {
				prefixMatch = km
			}
		}
	}

	if exactMatch != nil {
		return exactMatch, true, false
	}
	if prefixMatch != nil {
		return prefixMatch, false, true
	}

	return nil, false, false
}

// NormalizeKeyMapping expands leader and normalizes key notation at mapping time
func NormalizeKeyMapping(km *KeyMapping, leader string) {
	// Expand leader in LHS
	km.LHS = ExpandLeader(km.LHS, leader)

	// Expand leader in RHS (if it's a string command)
	if !km.IsFunc && km.RHS != "" {
		km.RHS = ExpandLeader(km.RHS, leader)
	}

	// Note: We don't parse key notation here because we want to preserve
	// the original notation for display purposes. Parsing happens at lookup time.
}

// ListKeymaps returns all keymaps for a given mode (or all if mode is empty)
func ListKeymaps(mappings []KeyMapping, mode string, lhsFilter string) []KeyMapping {
	result := []KeyMapping{}

	for _, km := range mappings {
		// Filter by mode if specified
		if mode != "" && km.Mode != mode {
			continue
		}

		// Filter by LHS if specified
		if lhsFilter != "" && km.LHS != lhsFilter {
			continue
		}

		result = append(result, km)
	}

	return result
}
