package ecsview

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/swartzrock/ecsview/cmd/aws"
	"github.com/swartzrock/ecsview/cmd/utils"
)

// Stores information about an AWS ECS cluster
type ClusterData struct {
	Cluster          *aws.EcsCluster
	Services         []*ecs.Service
	Tasks            []*ecs.Task
	TaskDefArnLookup map[string]*ecs.TaskDefinition
	Containers       []*aws.EcsContainer
	Refreshed        time.Time
}

var clusters []*aws.EcsCluster
var clusterArnToEcsContainersMap = make(map[string][]*aws.EcsContainer)
var clusterArnToEcsDataMap = make(map[string]*ClusterData)

// Returns a slice of ECS Clusters. If this is the first time, the clusters and their instances are loaded and cached.
func GetClusters() []*aws.EcsCluster {
	if clusters == nil {
		loadClustersAndContainers()
	}

	return clusters
}

// Returns data about the given cluster
func GetClusterData(cluster *aws.EcsCluster) *ClusterData {
	if data, found := clusterArnToEcsDataMap[*cluster.ClusterArn]; found {
		return data
	}
	return loadAndSaveEcsData(cluster)
}

// Returns data about the cluster, freshly loaded from AWS
func RefreshClusterData(cluster *aws.EcsCluster) *ClusterData {
	clusterArnToEcsContainersMap[*cluster.ClusterArn] = nil
	return loadAndSaveEcsData(cluster)
}

// Returns the containers for a given cluster
func GetClusterContainers(cluster *aws.EcsCluster) []*aws.EcsContainer {
	return clusterArnToEcsContainersMap[*cluster.ClusterArn]
}

// Got an AWS error? Likely a credential issue, so let the user know they need to configure their credentials
func fatalAwsError(err error) {
	if err != nil {
		log.Fatal("Unable to locate AWS credentials. You can configure credentials by running 'aws configure'.")
	}
}

func loadClustersAndContainers() {

	clusterResults, err := aws.DescribeClusters()
	fatalAwsError(err)
	clusters = aws.NewEcsClusters(clusterResults)

	sort.SliceStable(clusters, func(i, j int) bool {
		return 0 > strings.Compare(*clusters[i].ClusterName, *clusters[j].ClusterName)
	})

	totalInstances := 0
	for _, cluster := range clusters {
		totalInstances += len(loadAndSaveClusterContainers(cluster))
	}
}

func loadAndSaveEcsData(cluster *aws.EcsCluster) *ClusterData {

	services, err := aws.DescribeClusterServices(cluster.Cluster)
	fatalAwsError(err)
	sort.SliceStable(services, func(i, j int) bool {
		return 0 > strings.Compare(*services[i].ServiceName, *services[j].ServiceName)
	})

	tasks, err := aws.DescribeClusterTasks(cluster.Cluster)
	fatalAwsError(err)
	sort.SliceStable(tasks, func(i, j int) bool {
		return 0 > strings.Compare(utils.RemoveAllRegex(`.*/`, *tasks[i].TaskDefinitionArn), utils.RemoveAllRegex(`.*/`, *tasks[j].TaskDefinitionArn))
	})

	taskDefinitions, err := aws.GetTaskDefinitions(tasks)
	fatalAwsError(err)
	taskDefinitionArnLookup := make(map[string]*ecs.TaskDefinition)
	for _, taskDef := range taskDefinitions {
		taskDefinitionArnLookup[*taskDef.TaskDefinitionArn] = taskDef
	}

	containerPluses := clusterArnToEcsContainersMap[*cluster.ClusterArn]
	// Check if the user refreshed, clearing our cache of instances
	if containerPluses == nil {
		containerPluses = loadAndSaveClusterContainers(cluster)
	}

	data := &ClusterData{
		Cluster:          cluster,
		Services:         services,
		Tasks:            tasks,
		TaskDefArnLookup: taskDefinitionArnLookup,
		Containers:       containerPluses,
		Refreshed:        time.Now(),
	}

	clusterArnToEcsDataMap[*cluster.ClusterArn] = data

	return data
}

func loadAndSaveClusterContainers(cluster *aws.EcsCluster) []*aws.EcsContainer {

	containers, err := aws.DescribeContainerInstances(cluster.Cluster)
	fatalAwsError(err)
	sort.SliceStable(containers, func(i, j int) bool {
		return 0 > strings.Compare(*containers[i].Ec2InstanceId, *containers[j].Ec2InstanceId)
	})
	containerPluses := aws.NewEcsContainers(containers)

	clusterArnToEcsContainersMap[*cluster.ClusterArn] = containerPluses
	return containerPluses
}
