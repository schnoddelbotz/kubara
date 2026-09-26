package cluster

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/kubara-io/kubara/internal/catalog"
	"github.com/kubara-io/kubara/internal/config"
	"github.com/kubara-io/kubara/internal/utils"

	"github.com/urfave/cli/v3"
)

func CreateClusterInfo() *cli.Command {
	cmd := &cli.Command{
		Name:        "info",
		Usage:       "Extracts basic information from config.yaml",
		UsageText:   "kubara catalog info FIELD",
		Description: "Extracts given jsonpath FIELD from config.yaml, e.g. .Sepc.Version",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: "field",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
		},
		Action: func(c context.Context, cmd *cli.Command) error {
			cwd, err := filepath.Abs(cmd.String("work-dir"))
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			configFilePath, err := utils.GetFullPath(cmd.String("config-file"), cwd)
			if err != nil {
				return fmt.Errorf("get config file path: %w", err)
			}

			catalogOptions, err := catalog.ResolveLoadOptions(cwd, "", cmd.StringSlice("catalog"), cmd.Bool("catalog-overwrite"))
			if err != nil {
				return fmt.Errorf("could not resolve catalog options: %w", err)
			}

			configStore := config.NewConfigStore(cwd, configFilePath, catalogOptions)
			err = configStore.Load()
			if err != nil {
				return fmt.Errorf("config load: %w", err)
			}

			clusters := configStore.GetConfig()

			fieldValue, err := ConfigInfo(clusters, cmd.StringArg("field"))
			if err != nil {
				return err
			}
			fmt.Println(fieldValue)
			return nil
		},
	}

	return cmd
}

func ConfigInfo(cfg *config.Config, fieldName string) (string, error) {
	tmpl, err := template.New("info").Parse("{{" + fieldName + "}}")
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
