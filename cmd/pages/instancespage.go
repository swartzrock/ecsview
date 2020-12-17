package pages

import (
	"fmt"
	"strings"

	"github.com/swartzrock/ecsview/cmd/ecsview"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/thoas/go-funk"

	"github.com/swartzrock/ecsview/cmd/aws"
	"github.com/swartzrock/ecsview/cmd/ui"
	"github.com/swartzrock/ecsview/cmd/utils"
)

var latestEcsAgentVersion *string

// Returns a page that displays the container instances in a cluster
func NewInstancesPage() *ClusterDetailsPage {

	instancesTable := tview.NewTable()
	instancesTable.
		SetBorders(true).
		SetBorder(true).
		SetBorderColor(tcell.ColorDimGray).
		SetTitle(" 📦 ECS Instances ")

	instancesTableInfo := &ui.TableInfo{
		Table:      instancesTable,
		Alignment:  []int{ui.L, ui.L, ui.L, ui.L, ui.L, ui.L, ui.L, ui.L, ui.L},
		Expansions: []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		Selectable: true,
	}
	ui.AddTableConfigData(instancesTableInfo, 0, [][]string{
		{"Instance Id ▾", "Status", "Type", "ECS Agent", "Registered", "Tasks", "CPU", "Memory"},
	}, tcell.ColorYellow)

	return &ClusterDetailsPage{
		"Instances",
		instancesTableInfo,
		instancesPageRenderer(instancesTableInfo),
	}

}

func instancesPageRenderer(tableInfo *ui.TableInfo) func(*ecsview.ClusterData) {
	return func(e *ecsview.ClusterData) {
		renderInstancesTable(tableInfo, e)
	}
}

func renderInstancesTable(tableInfo *ui.TableInfo, ecsData *ecsview.ClusterData) {

	if latestEcsAgentVersion == nil {
		latestEcsAgentVersion, _ = aws.GetLatestECSAgentVersion()
	}

	ui.TruncTableRows(tableInfo.Table, 1)
	if len(ecsData.Containers) == 0 {
		return
	}

	data := funk.Map(ecsData.Containers, func(instance *aws.EcsContainer) []string {

		agentVersion := *instance.VersionInfo.AgentVersion
		if latestEcsAgentVersion == nil {
			agentVersion = agentVersion + " ❓"
		} else if agentVersion == *latestEcsAgentVersion {
			agentVersion = agentVersion + " ✅"
		} else {
			agentVersion = agentVersion + " ⚠️"
		}

		meterWidth := 5
		cpuMeter := ""
		memoryMeter := ""
		usage := instance.GetStats()
		if usage != nil {
			cpuMeter = utils.BuildAsciiMeterCurrentTotal(usage.CpuUsed, usage.CpuTotal, meterWidth)
			memoryMeter = utils.BuildAsciiMeterCurrentTotal(usage.MemoryUsed, usage.MemoryTotal, meterWidth)
		}

		instanceType := "n/a"
		instanceTypeAttribute := instance.GetAttribute("ecs.instance-type")
		if instanceTypeAttribute != nil {
			instanceType = *instanceTypeAttribute
		}

		taskCount := utils.I64ToString(*instance.RunningTasksCount)
		if *instance.PendingTasksCount > 0 {
			taskCount = fmt.Sprintf("%s (%d pending)", taskCount, *instance.PendingTasksCount)
		}
		if taskCount != "0" {
			tasks := make([]string, 0)
			for _, task := range ecsData.Tasks {
				if task.ContainerInstanceArn != nil && *task.ContainerInstanceArn == *instance.ContainerInstanceArn {
					tasks = append(tasks, aws.ShortenTaskDefArn(task.TaskDefinitionArn))
				}
			}
			instanceTasks := strings.Join(tasks, ",")
			taskCount = fmt.Sprintf("%s: %s", taskCount, utils.TakeLeft(instanceTasks, 40))
		}

		return []string{
			*instance.Ec2InstanceId,
			utils.LowerTitle(*instance.Status),
			instanceType,
			agentVersion,
			utils.FormatLocalDate(*instance.RegisteredAt),
			taskCount,
			cpuMeter,
			memoryMeter,
		}
	}).([][]string)

	ui.AddTableConfigData(tableInfo, 1, data, tcell.ColorWhite)

	instanceIdStyle := tcell.StyleDefault.Bold(true).Foreground(tcell.ColorWhite)
	ui.SetColumnStyle(tableInfo.Table, 0, 1, instanceIdStyle)

	usageMeterStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkCyan)
	ui.SetColumnStyle(tableInfo.Table, 6, 1, usageMeterStyle)
	ui.SetColumnStyle(tableInfo.Table, 7, 1, usageMeterStyle)

}
