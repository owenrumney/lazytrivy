package dockerClient

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/context"
	"github.com/docker/cli/cli/flags"
)

func getHostEndpoint() (dockerHost string, skipTlsVerify bool, err error) {

	cli, err := command.NewDockerCli()
	if err != nil {
		return "unix:///var/run/docker.sock", true, err
	}

	cli.Initialize(flags.NewClientOptions())
	currentContext := cli.CurrentContext()
	metadata, err := cli.ContextStore().GetMetadata(currentContext)
	if err != nil {
		return "unix:///var/run/docker.sock", true, err
	}

	if endpoint, ok := metadata.Endpoints["docker"]; ok {
		if deets, ok := endpoint.(context.EndpointMetaBase); ok {
			return deets.Host, deets.SkipTLSVerify, nil
		}
	}

	return "unix:///var/run/docker.sock", true, nil
}
