package config

import (
	"testing"
)

func TestExpandLeader(t *testing.T) {
	tests := []struct {
		key    string
		leader string
		want   string
	}{
		{"<leader>w", " ", " w"},
		{"<Leader>q", ",", ",q"},
		{"<LEADER>x", "\\", "\\x"},
		{"<leader>", " ", " "},
		{"jj", " ", "jj"},
		{"<leader>a<leader>b", " ", " a b"},
	}

	for _, tt := range tests {
		got := ExpandLeader(tt.key, tt.leader)
		if got != tt.want {
			t.Errorf("ExpandLeader(%q, %q) = %q, want %q", tt.key, tt.leader, got, tt.want)
		}
	}
}

func TestParseKeyNotation(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"<CR>", "\r"},
		{"<Esc>", "\x1b"},
		{"<Tab>", "\t"},
		{"<BS>", "\x08"},
		{"<Space>", " "},
		{"jj", "jj"},
		{"<CR><CR>", "\r\r"},
	}

	for _, tt := range tests {
		got := ParseKeyNotation(tt.input)
		if got != tt.want {
			t.Errorf("ParseKeyNotation(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestKeyPrecedence(t *testing.T) {
	tests := []struct {
		lhs  string
		want int
	}{
		{"a", 100},         // 1 char * 100
		{"ab", 200},        // 2 chars * 100
		{"abc", 300},       // 3 chars * 100
		{"<C-x>", 550},     // 5 chars * 100 + 50 (special key bonus)
		{"<leader>w", 950}, // 9 chars * 100 + 50
	}

	for _, tt := range tests {
		got := KeyPrecedence(tt.lhs)
		if got != tt.want {
			t.Errorf("KeyPrecedence(%q) = %d, want %d", tt.lhs, got, tt.want)
		}
	}
}

func TestKeyMappingsByPrecedence(t *testing.T) {
	mappings := []KeyMapping{
		{Mode: "n", LHS: "a"},
		{Mode: "n", LHS: "abc"},
		{Mode: "n", LHS: "ab"},
		{Mode: "i", LHS: "jj"},
	}

	sorted := KeyMappingsByPrecedence(mappings, "n")

	// Should be sorted by length (longer first)
	if len(sorted) != 3 {
		t.Fatalf("Expected 3 mappings, got %d", len(sorted))
	}

	if sorted[0].LHS != "abc" {
		t.Errorf("First should be 'abc', got %q", sorted[0].LHS)
	}
	if sorted[1].LHS != "ab" {
		t.Errorf("Second should be 'ab', got %q", sorted[1].LHS)
	}
	if sorted[2].LHS != "a" {
		t.Errorf("Third should be 'a', got %q", sorted[2].LHS)
	}
}

func TestFindKeyMapping(t *testing.T) {
	mappings := []KeyMapping{
		{Mode: "n", LHS: "a", RHS: "cmd_a"},
		{Mode: "n", LHS: "ab", RHS: "cmd_ab"},
		{Mode: "n", LHS: "abc", RHS: "cmd_abc"},
	}

	// Test exact match
	km, exact, prefix := FindKeyMapping("ab", mappings, "n")
	if km == nil {
		t.Fatal("Expected to find mapping")
	}
	if !exact || prefix {
		t.Errorf("Expected exact match, got exact=%v prefix=%v", exact, prefix)
	}
	if km.RHS != "cmd_ab" {
		t.Errorf("Expected cmd_ab, got %q", km.RHS)
	}

	// Test prefix match
	km, exact, prefix = FindKeyMapping("a", mappings, "n")
	if km == nil {
		t.Fatal("Expected to find mapping")
	}
	// "a" is an exact match for the "a" mapping
	if !exact || prefix {
		t.Errorf("Expected exact match for 'a', got exact=%v prefix=%v", exact, prefix)
	}

	// Test no match
	km, exact, prefix = FindKeyMapping("xyz", mappings, "n")
	if km != nil {
		t.Errorf("Expected no match, got %v", km)
	}
}

func TestNormalizeKeyMapping(t *testing.T) {
	tests := []struct {
		name    string
		km      KeyMapping
		leader  string
		wantLHS string
		wantRHS string
	}{
		{
			name:    "expand leader in LHS",
			km:      KeyMapping{LHS: "<leader>w", RHS: ":w<CR>"},
			leader:  " ",
			wantLHS: " w",
			wantRHS: ":w<CR>",
		},
		{
			name:    "expand leader in RHS",
			km:      KeyMapping{LHS: "a", RHS: "<leader>x"},
			leader:  ",",
			wantLHS: "a",
			wantRHS: ",x",
		},
		{
			name:    "expand leader in both",
			km:      KeyMapping{LHS: "<leader>a", RHS: "<leader>b"},
			leader:  "\\",
			wantLHS: "\\a",
			wantRHS: "\\b",
		},
		{
			name:    "function mapping ignores RHS",
			km:      KeyMapping{LHS: "<leader>f", RHS: "", IsFunc: true},
			leader:  " ",
			wantLHS: " f",
			wantRHS: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			km := tt.km
			NormalizeKeyMapping(&km, tt.leader)

			if km.LHS != tt.wantLHS {
				t.Errorf("LHS = %q, want %q", km.LHS, tt.wantLHS)
			}
			if km.RHS != tt.wantRHS {
				t.Errorf("RHS = %q, want %q", km.RHS, tt.wantRHS)
			}
		})
	}
}

func TestListKeymaps(t *testing.T) {
	mappings := []KeyMapping{
		{Mode: "n", LHS: "a", RHS: "cmd_a"},
		{Mode: "n", LHS: "b", RHS: "cmd_b"},
		{Mode: "i", LHS: "jj", RHS: "<Esc>"},
		{Mode: "v", LHS: "x", RHS: "d"},
	}

	// List all
	all := ListKeymaps(mappings, "", "")
	if len(all) != 4 {
		t.Errorf("Expected 4 mappings, got %d", len(all))
	}

	// Filter by mode
	normal := ListKeymaps(mappings, "n", "")
	if len(normal) != 2 {
		t.Errorf("Expected 2 normal mode mappings, got %d", len(normal))
	}

	// Filter by LHS
	specific := ListKeymaps(mappings, "", "jj")
	if len(specific) != 1 {
		t.Errorf("Expected 1 mapping for 'jj', got %d", len(specific))
	}
	if specific[0].Mode != "i" {
		t.Errorf("Expected insert mode, got %q", specific[0].Mode)
	}

	// Filter by both
	both := ListKeymaps(mappings, "n", "a")
	if len(both) != 1 {
		t.Errorf("Expected 1 mapping, got %d", len(both))
	}
}
