package onboard

import (
	"fmt"
	"path/filepath"

	"github.com/kubara-io/kubara/internal/catalog"
	"github.com/kubara-io/kubara/internal/config"
	"github.com/kubara-io/kubara/internal/utils"
	"github.com/urfave/cli/v3"
)

type templateData struct {
	Cluster *config.Cluster

	AppName        string
	ProjectName    string
	RepositoryURL  string
	RepositoryPath string
}

func NewOnboardCommand() *cli.Command {
	return &cli.Command{
		Name:        "onboard",
		Usage:       "Onboard projects, repositories and apps to GitOps engine",
		UsageText:   "kubara onboard [command]",
		Description: "Simplifies workload onboarding using templates",
		Commands: []*cli.Command{
			NewOnboardAppCommand(),
			NewOnboardAppSetCommand(),
			NewOnboardProjectCommand(),
			NewOnboardRepositoryCommand(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "engine",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
				Value: "argo-cd",
			},
		},
	}
}

func getClusterByName(cmd *cli.Command, name string) (*config.Cluster, error) {
	cwd, err := filepath.Abs(cmd.String("work-dir"))
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}
	configFilePath, err := utils.GetFullPath(cmd.String("config-file"), cwd)
	if err != nil {
		return nil, fmt.Errorf("get config file path: %w", err)
	}

	catalogOptions, err := catalog.ResolveLoadOptions(cwd, "", cmd.StringSlice("catalog"), cmd.Bool("catalog-overwrite"))
	if err != nil {
		return nil, fmt.Errorf("could not resolve catalog options: %w", err)
	}

	configStore := config.NewConfigStore(cwd, configFilePath, catalogOptions)
	err = configStore.Load()
	if err != nil {
		return nil, fmt.Errorf("config load: %w", err)
	}

	clusters := configStore.GetConfig().Clusters

	clusterNames := make([]string, len(clusters))
	for idx, cluster := range clusters {
		if cluster.Name == name {
			return &cluster, nil
		}
		clusterNames[idx] = cluster.Name
	}
	return nil, fmt.Errorf("cluster %q not found in config; available: %v", name, clusterNames)
}
