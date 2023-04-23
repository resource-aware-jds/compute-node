package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/resource-aware-jds/compute-node/config"
	"github.com/sirupsen/logrus"
)

type jobService struct {
	appConfig    config.Config
	dockerClient *client.Client
}

type JobService interface {
	RunJob(dockerImage string, name string, options types.ImagePullOptions, jobID string) error
	RemoveContainer(containerID string) error
}

func NewJobService(appConfig config.Config, dockerClient *client.Client) JobService {
	return &jobService{
		appConfig:    appConfig,
		dockerClient: dockerClient,
	}
}

func (j *jobService) RunJob(dockerImage string, name string, options types.ImagePullOptions, jobIdStr string) error {
	logrus.Info("Creating container: ", name, " with image: ", dockerImage)
	defer logrus.Info("Create container ", name, " success")

	ctx := context.Background()

	//Pull image
	out, err := j.dockerClient.ImagePull(ctx, dockerImage, options)
	if err != nil {
		logrus.Warn("Pull image error: ", err)
	} else {
		defer out.Close()
	}
	// Create container
	resp, err := j.dockerClient.ContainerCreate(ctx, j.getContainerConfig(dockerImage, j.appConfig.GRPC_SERVER_PORT, jobIdStr), j.getHostConfig(), nil, nil, name)
	if err != nil {
		logrus.Error("Create container error: ", err)
		return err
	}

	// Start container
	if err := j.dockerClient.ContainerStart(ctx, name, types.ContainerStartOptions{}); err != nil {
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

func (j *jobService) RemoveContainer(containerID string) error {
	ctx := context.Background()
	responseCh, errCh := j.dockerClient.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			logrus.Error(err)
			return err
		}
	case response := <-responseCh:
		if response.Error != nil {
			logrus.Error(response.Error)
			return errors.New(response.Error.Message)
		}
		err := j.dockerClient.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
		if err != nil {
			logrus.Error(err)
			return err
		}
	}
	return nil
}
