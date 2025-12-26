package editor

import "testing"

func TestFiletypeDetection(t *testing.T) {
	tests := []struct {
		filename string
		expected string
		comment  string
	}{
		{"test.go", "go", "//"},
		{"script.py", "python", "#"},
		{"app.js", "javascript", "//"},
		{"component.tsx", "typescript", "//"},
		{"main.c", "c", "//"},
		{"program.cpp", "cpp", "//"},
		{"lib.rs", "rust", "//"},
		{"script.rb", "ruby", "#"},
		{"index.php", "php", "//"},
		{"Main.java", "java", "//"},
		{"config.lua", "lua", "--"},
		{"setup.sh", "shell", "#"},
		{"config.yml", "yaml", "#"},
		{"data.json", "json", ""},
		{"README.md", "markdown", ""},
		{"Makefile", "make", "#"},
		{"Dockerfile", "dockerfile", "#"},
		{"style.css", "css", "/*"},
	}

	for _, tt := range tests {
		ft := detectFiletype(tt.filename)
		if ft == nil {
			t.Errorf("detectFiletype(%q) returned nil, expected %q", tt.filename, tt.expected)
			continue
		}
		if ft.Name != tt.expected {
			t.Errorf("detectFiletype(%q) = %q, expected %q", tt.filename, ft.Name, tt.expected)
		}
		if ft.Comment != tt.comment {
			t.Errorf("detectFiletype(%q).Comment = %q, expected %q", tt.filename, ft.Comment, tt.comment)
		}
	}
}

func TestFiletypeDetectionNoExtension(t *testing.T) {
	// Test that unknown files return nil
	ft := detectFiletype("unknown_file_without_extension")
	if ft != nil {
		t.Errorf("detectFiletype('unknown_file_without_extension') should return nil, got %v", ft)
	}

	// Test empty filename
	ft = detectFiletype("")
	if ft != nil {
		t.Errorf("detectFiletype('') should return nil, got %v", ft)
	}
}

func TestCommentToggle(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		initial  string
		expected string
	}{
		{
			name:     "comment Go code",
			filename: "test.go",
			initial:  "package main",
			expected: "// package main",
		},
		{
			name:     "uncomment Go code",
			filename: "test.go",
			initial:  "// package main",
			expected: "package main",
		},
		{
			name:     "comment Python code",
			filename: "test.py",
			initial:  "def hello():",
			expected: "# def hello():",
		},
		{
			name:     "uncomment Python with space",
			filename: "test.py",
			initial:  "# def hello():",
			expected: "def hello():",
		},
		{
			name:     "comment with leading whitespace",
			filename: "test.go",
			initial:  "    fmt.Println()",
			expected: "    // fmt.Println()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed := newTestEditor(t, tt.initial)
			ed.filename = tt.filename

			// Toggle comment on line 0
			ed.toggleLineComment(0)

			result := ed.buffer.String()
			if result != tt.expected {
				t.Errorf("toggleLineComment() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
