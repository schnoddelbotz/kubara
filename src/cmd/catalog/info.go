package catalog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"os"

	cat "github.com/kubara-io/kubara/internal/catalog"

	"github.com/urfave/cli/v3"
)

func NewCatalogInfo() *cli.Command {
	cmd := &cli.Command{
		Name:        "info",
		Usage:       "Extracts basic information from Catalog.yaml",
		UsageText:   "kubara catalog info CATALOG_FIELD",
		Description: "Extracts given jsonpath CATALOG_FIELD from Catalog.yaml, e.g. .Sepc.Version",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: "catalog-field",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
		},
		Action: func(c context.Context, cmd *cli.Command) error {
			fieldValue, err := CatalogInfo(cmd.StringArg("catalog-field"))
			if err != nil {
				return err
			}
			fmt.Println(fieldValue)
			return nil
		},
	}

	return cmd
}

func CatalogInfo(fieldName string) (string, error) {
	if _, err := os.Stat("Catalog.yaml"); errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("this directory is missing a Catalog.yaml")
	}
	manifest, err := cat.LoadCatalogManifest(".")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("info").Parse("{{" + fieldName + "}}")
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, manifest); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
