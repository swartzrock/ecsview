package pages

import (
	"fmt"
	"strings"

	"github.com/swartzrock/ecsview/cmd/ecsview"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/thoas/go-funk"

	"github.com/swartzrock/ecsview/cmd/aws"
	"github.com/swartzrock/ecsview/cmd/ui"
	"github.com/swartzrock/ecsview/cmd/utils"
)

// Returns a page that displays the running tasks in a cluster
func NewTasksPage() *ClusterDetailsPage {

	tasksTable := tview.NewTable()
	tasksTable.
		SetBorders(true).
		SetBorder(true).
		SetBorderColor(tcell.ColorDimGray).
		SetTitle(" 🐳 ECS Tasks ")

	tasksTableInfo := &ui.TableInfo{
		Table:      tasksTable,
		Alignment:  []int{ui.L, ui.L, ui.L, ui.L, ui.L, ui.L, ui.R},
		Expansions: []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		Selectable: true,
	}
	ui.AddTableConfigData(tasksTableInfo, 0, [][]string{
		{"TaskDef ▾", "Images", "Status", "Created", "EC2 Instance", "Arn", "Version"},
	}, tcell.ColorYellow)

	return &ClusterDetailsPage{
		"Tasks",
		tasksTableInfo,
		taskPageRenderer(tasksTableInfo),
	}
}

func taskPageRenderer(tableInfo *ui.TableInfo) func(*ecsview.ClusterData) {
	return func(e *ecsview.ClusterData) {
		renderTasksPage(tableInfo, e)
	}
}

func renderTasksPage(tableInfo *ui.TableInfo, ecsData *ecsview.ClusterData) {

	ui.TruncTableRows(tableInfo.Table, 1)

	if len(ecsData.Tasks) == 0 {
		return
	}

	arnToEc2InstanceIdMap := make(map[string]string)
	for _, instance := range ecsData.Containers {
		arnToEc2InstanceIdMap[*instance.ContainerInstanceArn] = *instance.Ec2InstanceId
	}

	connectedToEmojiMap := map[string]string{"CONNECTED": "🔗", "DISCONNECTED": "🚫"}

	data := funk.Map(ecsData.Tasks, func(task *ecs.Task) []string {

		connected := connectedToEmojiMap["DISCONNECTED"]
		if task.Connectivity != nil {
			connected = connectedToEmojiMap[*task.Connectivity]
		}
		status := fmt.Sprintf("%s %s", utils.LowerTitle(*task.LastStatus), connected)

		taskImages := "n/a"
		maxImageWidth := 20
		if taskDef, found := ecsData.TaskDefArnLookup[*task.TaskDefinitionArn]; found {
			images := funk.Map(taskDef.ContainerDefinitions, func(d *ecs.ContainerDefinition) string {
				return utils.TakeLeft(utils.RemoveAllRegex(`.*/`, *d.Image), maxImageWidth)
			}).([]string)
			taskImages = strings.Join(images, ",")
		}

		ec2InstanceId := "n/a"
		if task.ContainerInstanceArn != nil {
			ec2InstanceId = arnToEc2InstanceIdMap[*task.ContainerInstanceArn]
		}

		return []string{
			aws.ShortenTaskDefArn(task.TaskDefinitionArn),
			taskImages,
			status,
			utils.FormatLocalDateTimeAmPmZone(*task.CreatedAt),
			ec2InstanceId,
			utils.TakeRight(utils.RemoveAllRegex(`.*/`, *task.TaskArn), 8),
			utils.I64ToString(*task.Version),
		}
	}).([][]string)

	ui.AddTableConfigData(tableInfo, 1, data, tcell.ColorWhite)
	taskArnColumnStyle := tcell.StyleDefault.Bold(true).Foreground(tcell.ColorWhite)
	ui.SetColumnStyle(tableInfo.Table, 0, 1, taskArnColumnStyle)

}
