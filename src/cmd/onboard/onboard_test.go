// own package because of circular dependencies otherwise of cmd and cluster
package onboard_test

import (
	"testing"

	"github.com/kubara-io/kubara/cmd/onboard"
	"github.com/stretchr/testify/assert"
)

func TestNewOnboardCommand(t *testing.T) {
	command := onboard.NewOnboardCommand()

	assert.Equal(t, "onboard", command.Name)
	assert.Equal(t, "Onboard projects, repositories and apps to GitOps engine", command.Usage)
	assert.Equal(t, "kubara onboard [command]", command.UsageText)
	assert.Equal(t, "Simplifies workload onboarding using templates", command.Description)

	appCommand := onboard.NewOnboardAppCommand()

	assert.Equal(t, "app", appCommand.Name)
	assert.Equal(t, "Add a new app to a cluster", appCommand.Usage)
	assert.Equal(t, "kubara onboard app CLUSTER_NAME", appCommand.UsageText)
	assert.Equal(t, "Adds a new application deployment to the GitOps engine", appCommand.Description)

	// + appSet

	projectCommand := onboard.NewOnboardProjectCommand()

	assert.Equal(t, "project", projectCommand.Name)
	assert.Equal(t, "Add a new project to a cluster's GitOps engine", projectCommand.Usage)
	assert.Equal(t, "kubara onboard project CLUSTER_NAME", projectCommand.UsageText)
	assert.Equal(t, "Add a new project to a cluster's GitOps engine", projectCommand.Description)

	repositoryCommand := onboard.NewOnboardRepositoryCommand()

	assert.Equal(t, "repository", repositoryCommand.Name)
	assert.Equal(t, "Add a new repository to a cluster's GitOps engine", repositoryCommand.Usage)
	assert.Equal(t, "kubara onboard repository CLUSTER_NAME", repositoryCommand.UsageText)
	assert.Equal(t, "Add a new repository to a cluster's GitOps engine", repositoryCommand.Description)
}

// func TestListAllClustersNoError(t *testing.T) {
// 	dir := t.TempDir()
// 	configPath := testutil.CreateTestConfig(t, dir, testutil.CreateTestCluster(t))

// 	testutil.CreateDefaultGenerateTestEnv(t, dir)

// 	cliFlags := flags.NewGlobalFlags().CLIFlags()
// 	app := testutil.CreateTestAppWithFlags(cliFlags, cluster.NewClusterCommand())
// 	args := []string{"kubara", "--config-file", configPath, "--work-dir", dir, "cluster", "list"}
// 	err := app.Run(context.Background(), args)
// 	require.NoError(t, err)
// }

// func TestAddNewSpokesCluster(t *testing.T) {
// 	spokeName := "coolNewSpoke"
// 	dir := t.TempDir()
// 	configPath := testutil.CreateTestConfig(t, dir, testutil.CreateTestCluster(t))

// 	testutil.CreateDefaultGenerateTestEnv(t, dir)

// 	cliFlags := flags.NewGlobalFlags().CLIFlags()
// 	app := testutil.CreateTestAppWithFlags(cliFlags, cluster.NewClusterCommand())
// 	args := []string{
// 		"kubara",
// 		"--config-file", configPath,
// 		"--work-dir", dir,
// 		"--catalog", testutil.GeneralCatalogPath(),
// 		"cluster", "add", spokeName,
// 	}
// 	err := app.Run(context.Background(), args)

// 	require.NoError(t, err)
// }
