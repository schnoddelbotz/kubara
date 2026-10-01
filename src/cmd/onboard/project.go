package onboard

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

//go:embed project.tplt
var projectTemplate string

func NewOnboardProjectCommand() *cli.Command {
	return &cli.Command{
		Name:        "project",
		Usage:       "Add a new project to a cluster's GitOps engine",
		UsageText:   "kubara onboard project CLUSTER_NAME",
		Description: "Add a new project to a cluster's GitOps engine",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: "cluster-name",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
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
			log.Info().Msgf("Onboarding new project to cluster %q using engine %q", cluster.Name, engine)

			projectName := cmd.String("project-name")
			if len(projectName) == 0 {
				return fmt.Errorf("Missing required --project-name flag")
			}

			outFileName := filepath.Join("platform-configs", clusterName, "helm", engine, "values-project-"+projectName+".yaml")
			return execTemplate(projectTemplate, outFileName, cmd.Bool("stdout"), templateData{
				Cluster:     cluster,
				ProjectName: projectName,
				Repository: Repository{
					URL: "*", // FIXME
				},
			})
		},
	}
}
