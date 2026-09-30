package onboard

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"slices"
	"text/template"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

//go:embed app.tplt
var appTemplate string

var allowedEngines = []string{"argo-cd"} // must match component directory name

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
				Name: "engine",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
				Value: "argo-cd",
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
			// all parameters can be given as arguments
			// if arguments are missing, get them interactively
			//  - if project not given, run onboard project first
			//  - if repository not given, run onboard repository first
			//
			engine := cmd.String("engine") // hmm. comes from config soon...?
			// clusterName := cmd.StringArg("cluster-name")
			// if len(clusterName) == 0 {
			// 	cli.ShowSubcommandHelpAndExit(cmd, 1)
			// }
			cluster, err := getClusterByName(cmd, cmd.StringArg("cluster-name"))
			if err != nil {
				return err
			}

			log.Info().Msgf("Cluster %q: onboard app to %s ...", cluster.Name, engine)

			tmpl, err := template.New("app").Parse(appTemplate)
			if err != nil {
				return err
			}

			data := templateData{
				Cluster:        cluster,
				AppName:        "myapp",
				ProjectName:    cluster.Name, // by default, let override
				RepositoryURL:  "https://foo",
				RepositoryPath: "app-x",
			}

			return tmpl.Execute(os.Stdout, data)

			// outputFile, err := os.Create("output.txt")
			// if err != nil {
			// 	panic(err)
			// }
			// defer outputFile.Close()
			// return tmpl.Execute(outputFile, data)
		},
	}
}
