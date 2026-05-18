package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var skillsFS embed.FS

func SetSkillsFS(f embed.FS) {
	skillsFS = f
}

var skillsTarget string

func init() {
	skillsCmd := &cobra.Command{
		Use:   "skills",
		Short: "Manage bundled Claude skills",
	}

	skillsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List bundled skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := skillsFS.ReadDir("skills")
			if err != nil {
				return err
			}
			for _, e := range entries {
				if e.IsDir() {
					fmt.Println(e.Name())
				}
			}
			return nil
		},
	}

	skillsInstallCmd := &cobra.Command{
		Use:   "install [skill-name...]",
		Short: "Install bundled skills (default: all)",
		Long: `Install bundled skills to ~/.claude/skills/ (or --target).
With no arguments, installs all bundled skills.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := resolveSkillsTarget()
			if err != nil {
				return err
			}
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("create target: %w", err)
			}

			toInstall := args
			if len(toInstall) == 0 {
				entries, err := skillsFS.ReadDir("skills")
				if err != nil {
					return err
				}
				for _, e := range entries {
					if e.IsDir() {
						toInstall = append(toInstall, e.Name())
					}
				}
			}

			for _, name := range toInstall {
				if err := installSkill(name, target); err != nil {
					return fmt.Errorf("install %s: %w", name, err)
				}
				fmt.Printf("Installed %s → %s/%s\n", name, target, name)
			}
			return nil
		},
	}
	skillsInstallCmd.Flags().StringVar(&skillsTarget, "target", "", "Target directory (default: ~/.claude/skills)")

	skillsCmd.AddCommand(skillsListCmd)
	skillsCmd.AddCommand(skillsInstallCmd)
	rootCmd.AddCommand(skillsCmd)
}

func resolveSkillsTarget() (string, error) {
	if skillsTarget != "" {
		return skillsTarget, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "skills"), nil
}

func installSkill(name, target string) error {
	srcDir := filepath.Join("skills", name)
	if _, err := skillsFS.ReadDir(srcDir); err != nil {
		return fmt.Errorf("skill not found: %s", name)
	}
	return fs.WalkDir(skillsFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(target, name, rel)
		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}
		data, err := skillsFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, data, 0644)
	})
}
