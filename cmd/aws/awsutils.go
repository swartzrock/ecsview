package aws

import (
	"context"

	"github.com/google/go-github/v33/github"

	"github.com/swartzrock/ecsview/cmd/utils"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

var sess *session.Session

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
}

// Return a slice of the ECS clusters in the current AWS account
func DescribeClusters() ([]*ecs.Cluster, error) {

	client := ecs.New(sess)
	var describeErr error

	var clusters []*ecs.Cluster
	include := "STATISTICS"

	err := client.ListClustersPagesWithContext(context.Background(), &ecs.ListClustersInput{}, func(output *ecs.ListClustersOutput, b bool) bool {
		if len(output.ClusterArns) == 0 {
			return false
		}
		clusterDetails, err := client.DescribeClusters(&ecs.DescribeClustersInput{
			Clusters: output.ClusterArns,
			Include:  []*string{&include},
		})
		if err != nil {
			describeErr = err
			return false
		}

		clusters = append(clusters, clusterDetails.Clusters...)
		return true
	})

	// If List didn't return an error, use the Describe error
	if err == nil {
		err = describeErr
	}

	return clusters, err
}

// Return a slice of the services in the given ECS cluster
func DescribeClusterServices(c *ecs.Cluster) ([]*ecs.Service, error) {
	client := ecs.New(sess)
	var describeErr error

	var services []*ecs.Service

	err := client.ListServicesPagesWithContext(context.Background(), &ecs.ListServicesInput{Cluster: c.ClusterArn}, func(output *ecs.ListServicesOutput, b bool) bool {
		if len(output.ServiceArns) == 0 {
			return false
		}
		serviceDetails, err := client.DescribeServices(&ecs.DescribeServicesInput{
			Cluster:  c.ClusterArn,
			Services: output.ServiceArns,
		})
		if err != nil {
			describeErr = err
			return false
		}

		services = append(services, serviceDetails.Services...)
		return true
	})

	// If List didn't return an error, use the Describe error
	if err == nil {
		err = describeErr
	}

	return services, err
}

// Return a slice of the tasks in the given ECS cluster
func DescribeClusterTasks(c *ecs.Cluster) ([]*ecs.Task, error) {
	client := ecs.New(sess)
	var describeErr error

	var tasks []*ecs.Task

	err := client.ListTasksPagesWithContext(context.Background(), &ecs.ListTasksInput{Cluster: c.ClusterArn}, func(output *ecs.ListTasksOutput, b bool) bool {
		if len(output.TaskArns) == 0 {
			return false
		}
		taskDetails, err := client.DescribeTasks(&ecs.DescribeTasksInput{
			Cluster: c.ClusterArn,
			Tasks:   output.TaskArns,
		})
		if err != nil {
			describeErr = err
			return false
		}

		tasks = append(tasks, taskDetails.Tasks...)
		return true
	})

	// If List didn't return an error, use the Describe error
	if err == nil {
		err = describeErr
	}

	return tasks, err
}

// Return a slice of the task definitions in the given ECS tasks
func GetTaskDefinitions(tasks []*ecs.Task) ([]*ecs.TaskDefinition, error) {

	// Dedupe the task definitions by arn
	taskDefArns := make(map[string]bool)
	for _, task := range tasks {
		taskDefArns[*task.TaskDefinitionArn] = true
	}

	client := ecs.New(sess)
	taskDefinitions := make([]*ecs.TaskDefinition, 0)
	for taskDefArn := range taskDefArns {
		output, err := client.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{TaskDefinition: &taskDefArn})
		if err != nil {
			return nil, err
		}
		taskDefinitions = append(taskDefinitions, output.TaskDefinition)
	}

	return taskDefinitions, nil
}

// Return a short version of the task definition arn
func ShortenTaskDefArn(taskDefinitionArn *string) string {
	return utils.RemoveAllRegex(`.*/`, *taskDefinitionArn)
}

// Return a slice of the container instances in the given ECS cluster
func DescribeContainerInstances(c *ecs.Cluster) ([]*ecs.ContainerInstance, error) {
	client := ecs.New(sess)
	var describeErr error
	var containerInstances []*ecs.ContainerInstance

	err := client.ListContainerInstancesPagesWithContext(context.Background(), &ecs.ListContainerInstancesInput{Cluster: c.ClusterArn}, func(output *ecs.ListContainerInstancesOutput, b bool) bool {
		if len(output.ContainerInstanceArns) == 0 {
			return false
		}
		containerDetails, err := client.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
			Cluster:            c.ClusterArn,
			ContainerInstances: output.ContainerInstanceArns,
		})
		if err != nil {
			describeErr = err
			return false
		}

		containerInstances = append(containerInstances, containerDetails.ContainerInstances...)
		return true
	})

	// If List didn't return an error, use the Describe error
	if err == nil {
		err = describeErr
	}

	return containerInstances, err
}

// Read the latest released ECS Agent from Github
func GetLatestECSAgentVersion() (*string, error) {
	githubClient := github.NewClient(nil)
	releases, _, err := githubClient.Repositories.ListReleases(context.Background(), "aws", "amazon-ecs-agent", nil)
	if err != nil {
		return nil, err
	}
	latestRelease := utils.RemoveAllRegex("^v", *releases[0].TagName)
	return &latestRelease, nil
}
