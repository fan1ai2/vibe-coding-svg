package worker

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCountSVGPaths(t *testing.T) {
	tests := []struct {
		name  string
		svg   string
		count int
	}{
		{"empty", "", 0},
		{"no path", `<svg><rect/></svg>`, 0},
		{"one path", `<svg><path d="M0 0"/></svg>`, 1},
		{"three paths", `<svg><path d="M0 0"/><path d="M1 1"/><path d="M2 2"/></svg>`, 3},
		{"path in path", `<svg><path d="M0 0"/><path d=""/></svg>`, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountSVGPaths([]byte(tt.svg))
			if got != tt.count {
				t.Errorf("CountSVGPaths() = %d, want %d", got, tt.count)
			}
		})
	}
}

func TestConvertRasterToSVG(t *testing.T) {
	if _, err := exec.LookPath("vtracer"); err != nil {
		t.Skip("vtracer not installed")
	}

	inPath := "testdata/test.png"
	if _, err := os.Stat(inPath); os.IsNotExist(err) {
		t.Skip("testdata/test.png not found")
	}

	outPath := t.TempDir() + "/test.svg"
	if err := ConvertRasterToSVG(inPath, outPath); err != nil {
		t.Fatalf("ConvertRasterToSVG() error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(data) == 0 {
		t.Error("output SVG is empty")
	}
	if !strings.Contains(string(data), "<svg") && !strings.Contains(string(data), "<path") {
		t.Error("output does not appear to be valid SVG")
	}
}
