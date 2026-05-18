package main

import (
	"embed"

	"github.com/orvibodx/crm-cli/cmd"
)

//go:embed all:skills
var skillsFS embed.FS

func init() {
	cmd.SetSkillsFS(skillsFS)
}
