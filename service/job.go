package service

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/resource-aware-jds/compute-node/config"
	"github.com/sirupsen/logrus"
)

type jobService struct {
	appConfig config.Config
}

type JobService interface {
	RunJob(dockerImage string, name string, options types.ImagePullOptions, jobID string) error
}

func NewJobService(appConfig config.Config) JobService {
	return &jobService{
		appConfig: appConfig,
	}
}

func (j *jobService) RunJob(dockerImage string, name string, options types.ImagePullOptions, jobIdStr string) error {
	logrus.Info("Creating container: ", name, " with image: ", dockerImage)
	defer logrus.Info("Create container ", name, " success")

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.Error("Client connection error: ", err)
		return err
	}
	defer cli.Close()

	//Pull image
	out, err := cli.ImagePull(ctx, dockerImage, options)
	if err != nil {
		logrus.Warn("Pull image error: ", err)
	}
	defer out.Close()

	// Create container
	resp, err := cli.ContainerCreate(ctx, j.getContainerConfig(dockerImage, j.appConfig.GRPC_SERVER_PORT, jobIdStr), j.getHostConfig(), nil, nil, name)
	if err != nil {
		logrus.Error("Create container error: ", err)
		return err
	}

	// Start container
	if err := cli.ContainerStart(ctx, name, types.ContainerStartOptions{}); err != nil {
		logrus.Error(err)
		return err
	}

	fmt.Println(resp.ID)
	return nil
}

func (j *jobService) getHostConfig() *container.HostConfig {
	return &container.HostConfig{
		ExtraHosts: []string{"host.docker.internal:host-gateway"},
	}
}

func (j *jobService) getContainerConfig(dockerImage string, hostPort string, jobID string) *container.Config {
	return &container.Config{
		Image: dockerImage,
		Env:   []string{"HOST_PORT=" + hostPort, "JOB_ID=" + jobID},
	}
}
