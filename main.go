package main

import (
	"flag"
	"fmt"

	. "github.com/logrusorgru/aurora"

	"github.com/swartzrock/ecsview/cmd"
)

func main() {
	flag.Usage = func() {
		appName := BrightCyan("ecsview")
		fmt.Printf("Usage: %s\n\n%s uses your valid AWS session credentials to display a visual inspection of your account's ECS clusters.\n",
			appName, appName)
	}
	flag.Parse()

	cmd.Entrypoint()
}
