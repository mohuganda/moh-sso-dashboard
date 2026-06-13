package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var openAPIOutput string

var openAPICmd = &cobra.Command{
	Use:   "openapi",
	Short: "OpenAPI utilities",
}

var openAPIValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate that the starter OpenAPI document exists",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := os.Stat("docs/openapi.yaml")
		if err != nil {
			return err
		}
		if info.IsDir() || info.Size() == 0 {
			return fmt.Errorf("docs/openapi.yaml must be a non-empty file")
		}
		fmt.Println("OpenAPI document exists: docs/openapi.yaml")
		return nil
	},
}

var openAPIExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the starter OpenAPI document",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile("docs/openapi.yaml")
		if err != nil {
			return err
		}

		if openAPIOutput == "" || openAPIOutput == "-" {
			fmt.Print(string(data))
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(openAPIOutput), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(openAPIOutput, data, 0o644); err != nil {
			return err
		}
		fmt.Printf("OpenAPI document exported to %s\n", openAPIOutput)
		return nil
	},
}

func init() {
	openAPICmd.AddCommand(openAPIValidateCmd)
	openAPICmd.AddCommand(openAPIExportCmd)
	openAPIExportCmd.Flags().StringVarP(&openAPIOutput, "output", "o", "-", "Output path, or '-' for stdout")
}
