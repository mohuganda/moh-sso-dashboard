package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/moh-sso-dashboard/internal/version"
	"github.com/spf13/cobra"
)

var versionOutput string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print backend build information",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return writeVersion(cmd.OutOrStdout(), versionOutput)
	},
}

func init() {
	versionCmd.Flags().StringVar(&versionOutput, "output", "text", "Output format: text or json")
}

func writeVersion(writer io.Writer, output string) error {
	switch strings.ToLower(strings.TrimSpace(output)) {
	case "", "text":
		_, err := fmt.Fprintln(writer, version.String())
		return err
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(version.Get())
	default:
		return fmt.Errorf("unsupported output format %q; expected text or json", output)
	}
}
