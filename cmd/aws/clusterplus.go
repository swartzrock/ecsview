package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/thoas/go-funk"
)

// Adds helpful functions to an ecs.Cluster
type EcsCluster struct {
	*ecs.Cluster
}

func NewEcsCluster(cluster *ecs.Cluster) *EcsCluster {
	return &EcsCluster{
		cluster,
	}
}

func NewEcsClusters(clusters []*ecs.Cluster) []*EcsCluster {
	return funk.Map(clusters, func(c *ecs.Cluster) *EcsCluster {
		return NewEcsCluster(c)
	}).([]*EcsCluster)
}

func (c *EcsCluster) GetClusterType() string {
	nonZeroPairs := funk.Filter(c.Statistics, func(pair *ecs.KeyValuePair) bool {
		return *pair.Value != "0"
	}).([]*ecs.KeyValuePair)

	clusterType := "EC2"
	if len(nonZeroPairs) > 0 && strings.Contains(*nonZeroPairs[0].Value, "Fargate") {
		clusterType = "Fargate"
	}

	return clusterType
}
