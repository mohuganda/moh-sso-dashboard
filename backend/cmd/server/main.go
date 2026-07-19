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
	if migrationsRequested(os.Args[1:]) {
		if err := bootstrap.RunMigrations(); err != nil {
			fmt.Fprintf(os.Stderr, "database migration failed: %v\n", err)
			os.Exit(1)
		}
		return
	}
	bootstrap.Run()
}

func versionRequested(args []string) bool {
	return len(args) == 1 && (args[0] == "--version" || args[0] == "version")
}

func migrationsRequested(args []string) bool {
	return len(args) == 1 && (args[0] == "migrate" || args[0] == "--migrate-only")
}
