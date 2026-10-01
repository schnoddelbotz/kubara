package onboard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"text/template"

	"github.com/kubara-io/kubara/internal/catalog"
	"github.com/kubara-io/kubara/internal/config"
	"github.com/kubara-io/kubara/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

type templateData struct {
	Cluster     *config.Cluster
	Repository  Repository
	AppName     string
	ProjectName string
}

type Repository struct {
	Name   string
	URL    string
	Path   string
	User   string
	Secret RepoSecret
}

type RepoSecret struct {
	RemoteKey         string
	RemoteKeyProperty string
	Kind              string
	Name              string
}

func NewOnboardCommand() *cli.Command {
	/*
		kubara onboard project homelab-jh --stdout --project-name fooproj
		kubara onboard repository homelab-jh --stdout --repository-name foo --project-name fooproj
		kubara onboard app homelab-jh --stdout --app-name foo-app --repository-path foo
	*/
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
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			val := cmd.String("engine")
			if !slices.Contains(allowedEngines, val) {
				return ctx, fmt.Errorf("invalid value %q for --engine; must be one of %v", val, allowedEngines)
			}
			return ctx, nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				// will vanish once available via config.yaml
				Name:  "engine",
				Usage: "GitOps engine to use",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
				Value: "argo-cd",
			},
			&cli.StringFlag{
				Name:  "project-name",
				Usage: "GitOps engine project name",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
			&cli.BoolFlag{
				Name:  "stdout",
				Usage: "Print onboard output to stdout, not to file",
			},
		},
	}
}

func getClusterByName(cmd *cli.Command, name string) (*config.Cluster, error) {
	// the code in here is stolen/duplicated from cluster/list.go
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

func execTemplate(tplt, outFileName string, toStdout bool, data templateData) error {
	tmpl, err := template.New("app").Parse(tplt)
	if err != nil {
		return err
	}
	if toStdout {
		log.Info().Msgf("Printing to stdout instead of %s", outFileName)
		return tmpl.Execute(os.Stdout, data)
	}

	fileExists, _ := utils.FileExist(outFileName)
	if fileExists {
		return fmt.Errorf("refusing to overwrite existing overlay %q; manually remove it first", outFileName)
	}
	log.Info().Msgf("Writing app overlay to: %s", outFileName)

	outputFile, err := os.Create(outFileName)
	if err != nil {
		return err
	}
	defer func() { _ = outputFile.Close() }()
	return tmpl.Execute(outputFile, data)
}
