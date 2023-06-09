package handler

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/resource-aware-jds/common-go/proto"
	"github.com/resource-aware-jds/compute-node/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"strconv"
)

type JobHandler struct {
	proto.UnimplementedComputeNodeServer
	jobService service.JobService
}

func NewJobGrpcServer(jobService service.JobService) *JobHandler {
	return &JobHandler{
		jobService: jobService,
	}
}

func (j *JobHandler) SendJob(context context.Context, job *proto.Job) (*emptypb.Empty, error) {
	jobIdStr := strconv.Itoa(int(job.JobID))
	containerName := "rajds-" + jobIdStr
	err := j.jobService.RunJob(job.DockerImage, containerName, types.ImagePullOptions{}, jobIdStr)
	return &emptypb.Empty{}, err
}

func (j *JobHandler) ReportJob(context context.Context, report *proto.ReportJobRequest) (*emptypb.Empty, error) {
	if report.TotalJob == report.CurrentJob {
		jobIdStr := strconv.Itoa(int(report.JobID))
		containerName := "rajds-" + jobIdStr
		go j.jobService.RemoveContainer(containerName)
	}
	logrus.Info("Job id: ", report.JobID, " Current: ", report.CurrentJob, " Total: ", report.TotalJob)
	return &emptypb.Empty{}, nil
}
