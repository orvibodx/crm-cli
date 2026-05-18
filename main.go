package main

import "github.com/orvibo/crm-cli/cmd"

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
