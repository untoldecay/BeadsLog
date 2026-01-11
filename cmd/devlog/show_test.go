package main

import (
	"testing"
)

func TestIsDate(t *testing.T) {
	tests := []struct {
		name string
		input string
		want bool
	}{
		{
			name: "valid date format",
			input: "2024-01-15",
			want: true,
		},
		{
			name: "valid date format with single digit month",
			input: "2024-1-15",
			want: false,
		},
		{
			name: "invalid format - missing dashes",
			input: "20240115",
			want: false,
		},
		{
			name: "filename with .md extension",
			input: "2024-01-15.md",
			want: false,
		},
		{
			name: "relative path",
			input: "entries/my-feature.md",
			want: false,
		},
		{
			name: "empty string",
			input: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDate(tt.input)
			if got != tt.want {
				t.Errorf("isDate(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestShowByDate(t *testing.T) {
	// Create a temporary index file
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{
			name:    "existing date",
			date:    "2024-01-15",
			wantErr: false,
		},
		{
			name:    "non-existing date",
			date:    "2024-12-31",
			wantErr: true,
		},
		{
			name:    "invalid date format",
			date:    "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			showOpts.IndexPath = "./test-index.md"
			err := showByDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("showByDate(%q) error = %v, wantErr %v", tt.date, err, tt.wantErr)
			}
		})
	}
}

func TestShowByFilename(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{
			name:    "existing markdown file",
			file:    "./test-index.md",
			wantErr: false,
		},
		{
			name:    "existing file without extension",
			file:    "./test-index",
			wantErr: false,
		},
		{
			name:    "non-existing file",
			file:    "./nonexistent.md",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := showByFilename(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("showByFilename(%q) error = %v, wantErr %v", tt.file, err, tt.wantErr)
			}
		})
	}
}
