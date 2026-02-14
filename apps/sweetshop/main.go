package main

import (
	"fmt"
	"os"

	"github.com/bbsbb/go-edge/sweetshop/cmd"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "resources/config/"
	}

	command := "server"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	var err error
	switch command {
	case "server":
		err = cmd.RunServer(configPath)
	case "migrate":
		err = runMigrate(configPath)
	default:
		err = fmt.Errorf("unknown command: %s (expected: server, migrate)", command)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "sweetshop: %v\n", err)
		os.Exit(1)
	}
}

func runMigrate(configPath string) error {
	subcommand := "up"
	if len(os.Args) > 2 {
		subcommand = os.Args[2]
	}

	switch subcommand {
	case "up":
		return cmd.RunMigrateUp(configPath)
	case "reset":
		return cmd.RunMigrateReset(configPath)
	case "verify":
		return cmd.RunMigrateVerify(configPath)
	case "create":
		if len(os.Args) < 5 {
			return fmt.Errorf("usage: sweetshop migrate create NAME TYPE\n  TYPE must be \"sql\" or \"go\"")
		}
		return cmd.RunMigrateCreate(os.Args[3], os.Args[4], "internal/migrations/versions")
	default:
		return fmt.Errorf("unknown migrate subcommand: %s (expected: up, reset, verify, create)", subcommand)
	}
}
