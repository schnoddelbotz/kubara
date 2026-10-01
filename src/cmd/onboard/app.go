package onboard

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"text/template"

	"github.com/kubara-io/kubara/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

// for now/hack, embed templates -- might make more sense to have them in catalog?
//
//go:embed app.tplt
var appTemplate string

var allowedEngines = []string{"argo-cd"} // must match gitops engine component directory name

func NewOnboardAppCommand() *cli.Command {
	return &cli.Command{
		Name:        "app",
		Usage:       "Add a new app to a cluster",
		UsageText:   "kubara onboard app CLUSTER_NAME",
		Description: "Adds a new application deployment to the GitOps engine",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: "cluster-name",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "app-name",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
			&cli.StringFlag{
				Name: "project-name",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
			&cli.StringFlag{
				Name: "repository-url",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
			&cli.StringFlag{
				Name: "repository-path",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			val := cmd.String("engine")
			if !slices.Contains(allowedEngines, val) {
				return ctx, fmt.Errorf("invalid value %q for --engine; must be one of %v", val, allowedEngines)
			}
			return ctx, nil
		},
		Action: func(c context.Context, cmd *cli.Command) error {
			engine := cmd.String("engine") // hmm. comes from config soon...?
			clusterName := cmd.StringArg("cluster-name")
			if len(clusterName) == 0 {
				cli.ShowSubcommandHelpAndExit(cmd, 1)
			}
			cluster, err := getClusterByName(cmd, cmd.StringArg("cluster-name"))
			if err != nil {
				return err
			}
			log.Info().Msgf("Onboarding new app to cluster %q using engine %q", cluster.Name, engine)

			appName := cmd.String("app-name")
			if len(appName) == 0 {
				return fmt.Errorf("missing required --app-name")
			}
			projectName := cmd.String("project-name")
			if len(projectName) == 0 {
				projectName = cluster.Name + "-" + cluster.Stage
				log.Warn().Msgf("No --project-name given, using cluster's default project: %q", projectName)
			}
			repoURL := cmd.String("repository-url")
			if len(repoURL) == 0 {
				repoURL = cluster.ArgoCD.Repo.Git.Components.URL
				log.Warn().Msgf("No --repository-url given, using cluster's components repo: %s", repoURL)
			}
			repoPath := cmd.String("repository-path")
			if len(repoPath) == 0 {
				return fmt.Errorf("missing required --repository-path")
			}

			tmpl, err := template.New("app").Parse(appTemplate)
			if err != nil {
				return err
			}
			data := templateData{
				Cluster:        cluster,
				AppName:        appName,
				ProjectName:    projectName,
				RepositoryURL:  repoURL,
				RepositoryPath: repoPath,
			}

			outFileName := filepath.Join("platform-configs", clusterName, "helm", engine, "values-app-"+appName+".yaml")
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
		},
	}
}
