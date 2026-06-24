package main

import (
	"fmt"
	"os"

	"github.com/moh-sso-dashboard/internal/bootstrap"
	"github.com/moh-sso-dashboard/internal/version"
)

func main() {
	if versionRequested(os.Args[1:]) {
		fmt.Println(version.String())
		return
	}
	bootstrap.Run()
}

func versionRequested(args []string) bool {
	return len(args) == 1 && (args[0] == "--version" || args[0] == "version")
}
