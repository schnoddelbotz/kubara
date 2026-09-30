package onboard

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

func NewOnboardAppSetCommand() *cli.Command {
	return &cli.Command{
		Name:        "appset",
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

		Action: func(c context.Context, cmd *cli.Command) error {
			// todo: same base behaviour like app.go
			clusterName := cmd.StringArg("cluster-name")
			log.Info().Msgf("Cluster %q: onboard appset ...", clusterName)
			return nil
		},
	}
}
