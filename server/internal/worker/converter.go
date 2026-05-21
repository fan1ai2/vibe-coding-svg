package worker

import (
	"fmt"
	"os/exec"
	"strings"
)

func ConvertRasterToSVG(inputPath, outputPath string) error {
	cmd := exec.Command("vtracer", "--input", inputPath, "--output", outputPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("vtracer: %s: %w", string(out), err)
	}
	return nil
}

func CountSVGPaths(data []byte) int {
	return strings.Count(string(data), "<path ")
}
