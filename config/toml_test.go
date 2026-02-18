package config

import (
	"strings"
	"testing"
)

func TestParseTOML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "basic key value",
			input: "key = value",
			want:  map[string]string{"key": "value"},
		},
		{
			name:  "multiple keys",
			input: "key1 = value1\nkey2 = value2",
			want:  map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name: "with spaces",
			input: `
  key1   =   value1
  key2 = value2
`,
			want: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name: "with comments",
			input: `
# This is a comment
key1 = value1
# Another comment
key2 = value2
`,
			want: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name: "with empty lines",
			input: `
key1 = value1

key2 = value2

`,
			want: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:  "with double quotes",
			input: `key = "value"`,
			want:  map[string]string{"key": "value"},
		},
		{
			name:  "with single quotes",
			input: `key = 'value'`,
			want:  map[string]string{"key": "value"},
		},
		{
			name:  "complex value with quotes",
			input: `key = " hello world "`,
			want:  map[string]string{"key": " hello world "},
		},
		{
			name:    "invalid format",
			input:   "invalid line",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTOML(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTOML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for k, v := range tt.want {
					if got[k] != v {
						t.Errorf("ParseTOML()[%s] = %v, want %v", k, got[k], v)
					}
				}
				if len(got) != len(tt.want) {
					t.Errorf("ParseTOML() returned %d items, want %d", len(got), len(tt.want))
				}
			}
		})
	}
}

func TestParseTOMLFile(t *testing.T) {
	t.Run("existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := tmpDir + "/test.conf"

		err := writeFile(filePath, "key = value\nkey2 = value2")
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		result, err := ParseTOMLFile(filePath)
		if err != nil {
			t.Fatalf("ParseTOMLFile() error = %v", err)
		}

		if result["key"] != "value" {
			t.Errorf("ParseTOMLFile()[key] = %v, want value", result["key"])
		}
		if result["key2"] != "value2" {
			t.Errorf("ParseTOMLFile()[key2] = %v, want value2", result["key2"])
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := ParseTOMLFile("/nonexistent/path/file.conf")
		if err == nil {
			t.Error("ParseTOMLFile() should return error for nonexistent file")
		}
	})
}
