package main

import (
	"fmt"
	"os"

	"github.com/pkopriv2/golang-sdk/lang/tool"
	"github.com/pkopriv2/services-catalog/cli"
)

var (
	MainTool = tool.NewTool(
		tool.ToolDef{
			Name:  "catalog",
			Usage: "catalog [cmd] [arg]*",
			Desc: `
`,
		},
		cli.StartCommand,
		cli.ListServicesCommand,
		cli.LoadServicesCommand,
	)
)

func main() {
	env, err := tool.NewEnvironment("~/.konghq/config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to initialize environment: %v\n", err)
		os.Exit(1)
		return
	}

	defer fmt.Println()
	if err := tool.Run(env, MainTool, os.Args); err != nil {
		os.Exit(1)
		return
	}
}
