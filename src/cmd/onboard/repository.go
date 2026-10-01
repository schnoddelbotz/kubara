package onboard

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

//go:embed repository.tplt
var repositoryTemplate string

func NewOnboardRepositoryCommand() *cli.Command {
	return &cli.Command{
		Name:        "repository",
		Usage:       "Add a new repository to a cluster's GitOps engine",
		UsageText:   "kubara onboard repository CLUSTER_NAME",
		Description: "Add a new repository to a cluster's GitOps engine",
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
				Name: "repository-name",
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
				Name: "repository-user",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
				Value: "oauth2",
			},
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
			log.Info().Msgf("Onboarding new repository to cluster %q using engine %q", cluster.Name, engine)

			repoName := cmd.String("repository-name")
			if len(repoName) == 0 {
				return fmt.Errorf("missing required --repository-name")
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
			repoUser := cmd.String("repository-user")

			outFileName := filepath.Join("platform-configs", clusterName, "helm", engine, "values-repository-"+repoName+".yaml")
			return execTemplate(repositoryTemplate, outFileName, cmd.Bool("stdout"), templateData{
				Cluster:     cluster,
				AppName:     repoName,
				ProjectName: projectName,
				Repository: Repository{
					Name: repoName,
					URL:  repoURL,
					User: repoUser,
				},
			})
		},
	}
}
