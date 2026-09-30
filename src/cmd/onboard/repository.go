package onboard

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

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

		Action: func(c context.Context, cmd *cli.Command) error {
			// is using vault optional at all?
			// https://docs.kubara.io/latest-stable/5_workload_onboarding/add_app_repository/
			//
			// instruct user on how to add credential to vault?
			//   ... or: patch vault directly?
			//   ... note: in our case, bw is authoratative
			//
			//
			clusterName := cmd.StringArg("cluster-name")
			log.Info().Msgf("Cluster %q: onboard repo ...", clusterName)
			return nil
		},
	}
}
