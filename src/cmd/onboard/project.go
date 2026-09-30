package onboard

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

func NewOnboardProjectCommand() *cli.Command {
	return &cli.Command{
		Name:        "project",
		Usage:       "Add a new project to a cluster's GitOps engine",
		UsageText:   "kubara onboard app CLUSTER_NAME",
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
			clusterName := cmd.StringArg("cluster-name")
			log.Info().Msgf("Cluster %q: onboard project ...", clusterName)
			return nil
		},
	}
}
