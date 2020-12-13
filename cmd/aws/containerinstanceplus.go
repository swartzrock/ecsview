package aws

import (
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/thoas/go-funk"
)

// Adds helpful functions to an ecs.ContainerInstance
type EcsContainer struct {
	*ecs.ContainerInstance
}

func NewEcsContainer(instance *ecs.ContainerInstance) *EcsContainer {
	return &EcsContainer{
		instance,
	}
}

func NewEcsContainers(containers []*ecs.ContainerInstance) []*EcsContainer {
	return funk.Map(containers, func(c *ecs.ContainerInstance) *EcsContainer {
		return NewEcsContainer(c)
	}).([]*EcsContainer)
}

func (i *EcsContainer) GetRemainingResourceValue(resourceName string) *int64 {
	for _, r := range i.RemainingResources {
		if resourceName == *r.Name && r.IntegerValue != nil {
			return r.IntegerValue
		}
	}
	return nil
}

func (i *EcsContainer) GetRegisteredResourceValue(resourceName string) *int64 {
	for _, r := range i.RegisteredResources {
		if resourceName == *r.Name && r.IntegerValue != nil {
			return r.IntegerValue
		}
	}
	return nil
}

type EcsContainerStats struct {
	CpuUsed, CpuTotal, MemoryUsed, MemoryTotal int64
}

func (s *EcsContainerStats) Add(s1 *EcsContainerStats) {
	s.CpuUsed += s1.CpuUsed
	s.CpuTotal += s1.CpuTotal
	s.MemoryUsed += s1.MemoryUsed
	s.MemoryTotal += s1.MemoryTotal
}

func (i *EcsContainer) GetStats() *EcsContainerStats {
	cpuRemaining := i.GetRemainingResourceValue("CPU")
	cpuTotal := i.GetRegisteredResourceValue("CPU")
	memoryRemaining := i.GetRemainingResourceValue("MEMORY")
	memoryTotal := i.GetRegisteredResourceValue("MEMORY")

	if cpuRemaining != nil && cpuTotal != nil && memoryRemaining != nil && memoryTotal != nil {
		cpuUsed := *cpuTotal - *cpuRemaining
		memoryUsed := *memoryTotal - *memoryRemaining
		return &EcsContainerStats{cpuUsed, *cpuTotal, memoryUsed, *memoryTotal}
	} else {
		return nil
	}

}

func (i *EcsContainer) GetAttribute(name string) *string {
	for _, att := range i.Attributes {
		if att.Name != nil && *att.Name == name {
			return att.Value
		}
	}
	return nil
}
