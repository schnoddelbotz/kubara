package catalog

import (
	"context"
	"errors"
	"fmt"
	"os"

	cat "github.com/kubara-io/kubara/internal/catalog"

	"github.com/urfave/cli/v3"
)

func NewCatalogInfo() *cli.Command {
	cmd := &cli.Command{
		Name:        "info",
		Usage:       "Extracts basic information from Catalog.yaml (spec.version, metadata.name)",
		UsageText:   "kubara catalog info CATALOG_FIELD",
		Description: "Extracts given field from Catalog.yaml. Run this command from a catalog root that already contains Catalog.yaml.",
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

	// very limited yet -- psuedo-"yq syntax" fails for eg. metadata.annotations.org.opencontainers.... :/
	switch fieldName {
	case "metadata.name":
		return manifest.Metadata.Name, nil
	case "spec.version":
		return manifest.Spec.Version, nil
	default:
		return "", fmt.Errorf("only spec.version and metadata.name fields are supported")
	}
}
