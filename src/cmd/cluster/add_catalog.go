package cluster

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/kubara-io/kubara/internal/catalog"
	"github.com/kubara-io/kubara/internal/config"
	"github.com/kubara-io/kubara/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

func CreateAddClusterCatalogCommand() *cli.Command {
	return &cli.Command{
		Name:        "add-catalog",
		Usage:       "Add a new catalog to your config",
		UsageText:   "kubara cluster add-catalog CLUSTER_NAME CATALOG_URL",
		Description: "Adds a new catalog to an existing config yaml",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: "cluster-name",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
			&cli.StringArg{
				Name: "catalog-url",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
		},

		Action: func(c context.Context, cmd *cli.Command) error {
			clusterName := cmd.StringArg("cluster-name")
			if len(clusterName) == 0 {
				cli.ShowSubcommandHelpAndExit(cmd, 1)
			}
			catalogURL := cmd.StringArg("catalog-url")
			if len(catalogURL) == 0 {
				cli.ShowSubcommandHelpAndExit(cmd, 1)
			}

			cwd, err := filepath.Abs(cmd.String("work-dir"))
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			catalogOptions, err := catalog.ResolveLoadOptions(cwd, "", cmd.StringSlice("catalog"), cmd.Bool("catalog-overwrite"))
			if err != nil {
				return fmt.Errorf("could not resolve catalog options: %w", err)
			}

			configFilePath, err := utils.GetFullPath(cmd.String("config-file"), cwd)
			if err != nil {
				return fmt.Errorf("get config file path: %w", err)
			}

			configStore := config.NewConfigStore(cwd, configFilePath, catalogOptions)
			err = configStore.Load()
			if err != nil {
				return fmt.Errorf("config load: %w", err)
			}
			currentConfig := configStore.GetConfig()

			clusters := currentConfig.Clusters
			for idx, existing := range clusters {
				if existing.Name == clusterName {
					if slices.Contains(currentConfig.Clusters[idx].Catalogs, catalogURL) {
						return fmt.Errorf("cluster %q already uses catalog %q", clusterName, catalogURL)
					}
					currentConfig.Clusters[idx].Catalogs = append(currentConfig.Clusters[idx].Catalogs, catalogURL)
					if err = configStore.SaveToFile(); err != nil {
						return fmt.Errorf("save config to file: %w", err)
					}
					log.Info().Msgf("added %q to catalogs of cluster %q", catalogURL, clusterName)
					return nil
				}
			}

			return fmt.Errorf("cluster %q does not exist", clusterName)
		},
	}
}
