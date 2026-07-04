package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// GenerateVirtualToolCreatorSkill copies the virtual-tool-creator skill
// into the generated project as .agents/skills/virtual-tool-creator/.
func (g *Generator) GenerateVirtualToolCreatorSkill() error {
	destDir := filepath.Join(g.outputDir, ".agents", "skills", "virtual-tool-creator")
	skillPrefix := "skills/virtual-tool-creator"

	return fs.WalkDir(templatesFS, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasPrefix(path, "skills/virtual-tool-creator") {
			return nil
		}

		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded %s: %w", path, err)
		}

		relPath := strings.TrimPrefix(path, skillPrefix+"/")
		destPath := filepath.Join(destDir, relPath)

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", filepath.Dir(destPath), err)
		}
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", destPath, err)
		}
		return nil
	})
}
