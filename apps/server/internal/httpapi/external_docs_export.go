package httpapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type externalDocExportAsset struct {
	sourcePath string
	outputName string
}

var externalDocExportAssets = []externalDocExportAsset{
	{
		sourcePath: "externaldocs/openapi.yaml",
		outputName: "openapi.yaml",
	},
	{
		sourcePath: "externaldocs/integration-guide.md",
		outputName: "integration-guide.md",
	},
}

func ExportExternalDocs(outputDir string) error {
	targetDir := strings.TrimSpace(outputDir)
	if targetDir == "" {
		return fmt.Errorf("external docs output dir is required")
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create external docs output dir %q: %w", targetDir, err)
	}

	for _, asset := range externalDocExportAssets {
		body, err := externalDocsFS.ReadFile(asset.sourcePath)
		if err != nil {
			return fmt.Errorf("read embedded external doc %q: %w", asset.sourcePath, err)
		}

		outputPath := filepath.Join(targetDir, asset.outputName)
		if err := os.WriteFile(outputPath, body, 0o644); err != nil {
			return fmt.Errorf("write external doc %q: %w", outputPath, err)
		}
	}

	return nil
}
