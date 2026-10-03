package onboard

import (
	"context"
	"embed"
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

// TODO: Templates likely should live in catalog...?
//
//go:embed templates
var templatesFS embed.FS

type templateData struct {
	Cluster     *config.Cluster
	Repository  Repository
	AppName     string
	ProjectName string
	Engine      string
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
		kubara onboard app <tab> ... <tab>
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

// shellComplete for onboard sub-commands. Use --generate-shell-completion to debug.
func shellComplete(ctx context.Context, cmd *cli.Command) {
	if cmd.NArg() > 0 {
		// complete flag values
		if len(os.Args) > 1 {
			lastWord := os.Args[len(os.Args)-2]
			switch lastWord {
			case "--engine":
				for _, e := range allowedEngines {
					fmt.Println(e)
				}
				return
			// cannot offer completion for these flags' values ...
			// abort completion here to let user know custom value is required.
			case "--app-name":
				return
			case "--project-name":
				return
			case "--repository-url":
				return
			case "--repository-name":
				return
			case "--repository-path":
				return
			}
		}

		// complete flags if first/required argument is provided.
		// does not offer completion for flag values [yet] (e.g. --engine argo-cd).
		var allFlags []cli.Flag
		allFlags = append(allFlags, cmd.Flags...)
		allFlags = append(allFlags, NewOnboardCommand().Flags...)
		for _, flag := range allFlags {
			isUsed := false
			for _, name := range flag.Names() {
				for _, arg := range os.Args {
					if arg == "--"+name || arg == "-"+name {
						isUsed = true
						break
					}
				}
				if isUsed {
					break
				}
			}
			usageText := ""
			if vf, ok := flag.(interface{ GetUsage() string }); ok {
				usageText = vf.GetUsage()
			}
			if !isUsed && len(flag.Names()) > 0 {
				// fixme: fish only for now
				fmt.Printf("--%s\t%s\n", flag.Names()[0], usageText)
			}
		}
		return
	}
	// complete arg - cluster to apply onboarding command to, from config.yaml
	clusters, err := getClusters(cmd)
	if err != nil {
		return
	}
	for _, cluster := range clusters {
		fmt.Println(cluster.Name)
	}
}

func getClusterByName(cmd *cli.Command, name string) (*config.Cluster, error) {
	clusters, err := getClusters(cmd)
	if err != nil {
		return nil, err
	}
	clusterNames := make([]string, len(clusters))
	for idx, cluster := range clusters {
		if cluster.Name == name {
			return &cluster, nil
		}
		clusterNames[idx] = cluster.Name
	}
	return nil, fmt.Errorf("cluster %q not found in config; available: %v", name, clusterNames)
}

func getClusters(cmd *cli.Command) ([]config.Cluster, error) {
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

	return configStore.GetConfig().Clusters, nil
}

func execTemplate(tplt, outFileName string, toStdout bool, data templateData) error {
	tmpl, err := template.ParseFS(templatesFS, filepath.Join("templates", tplt))
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
