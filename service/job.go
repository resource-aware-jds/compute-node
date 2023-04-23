package service

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

type jobService struct {
}

type JobService interface {
	RunJob(dockerImage string, name string, options types.ImagePullOptions) error
}

func NewJobService() JobService {
	return &jobService{}
}

func (j *jobService) RunJob(dockerImage string, name string, options types.ImagePullOptions) error {
	logrus.Info("Creating container: ", name, " with image: ", dockerImage)
	defer logrus.Info("Create container ", name, " success")

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.Error("Client connection error: ", err)
		return err
	}
	defer cli.Close()

	// Pull image
	out, err := cli.ImagePull(ctx, dockerImage, options)
	if err != nil {
		logrus.Error("Pull image error: ", err)
		return err
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	// Create container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: dockerImage,
	}, nil, nil, nil, name)
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
