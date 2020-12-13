package pages

import (
	"fmt"
	"strings"

	"github.com/swartzrock/ecsview/cmd/ecsview"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/thoas/go-funk"

	"github.com/swartzrock/ecsview/cmd/ui"
	"github.com/swartzrock/ecsview/cmd/utils"
)

// Returns a page that displays the services in a cluster
func NewServicesPage() *ClusterDetailsPage {

	servicesTable := tview.NewTable()

	servicesTable.
		SetBorders(true).
		SetBorder(true).
		SetBorderColor(tcell.ColorDimGray).
		SetTitle(" 📋 ECS Services ")

	servicesTableInfo := &ui.TableInfo{
		Table:      servicesTable,
		Alignment:  []int{ui.L, ui.L, ui.L, ui.L, ui.L, ui.R, ui.R, ui.R},
		Expansions: []int{1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1},
		Selectable: true,
	}
	ui.AddTableConfigData(servicesTableInfo, 0, [][]string{
		{"Name ▾", "TaskDef", "Images", "Status", "Deployed", "Tasks"},
	}, tcell.ColorYellow)

	return &ClusterDetailsPage{
		"Services",
		servicesTableInfo,
		servicesPageRenderer(servicesTableInfo),
	}
}

func servicesPageRenderer(tableInfo *ui.TableInfo) func(*ecsview.ClusterData) {
	return func(e *ecsview.ClusterData) {
		renderServicesTable(tableInfo, e)
	}
}

func renderServicesTable(tableInfo *ui.TableInfo, ecsData *ecsview.ClusterData) {
	ui.TruncTableRows(tableInfo.Table, 1)

	if len(ecsData.Services) == 0 {
		return
	}

	data := funk.Map(ecsData.Services, func(service *ecs.Service) []string {

		serviceImages := "n/a"
		maxImageWidth := 50
		if taskDef, found := ecsData.TaskDefArnLookup[*service.TaskDefinition]; found {
			images := funk.Map(taskDef.ContainerDefinitions, func(d *ecs.ContainerDefinition) string {
				return utils.TakeLeft(utils.RemoveAllRegex(`.*/`, *d.Image), maxImageWidth)
			}).([]string)
			serviceImages = strings.Join(images, ",")
		}

		deployTimeTxt := "n/a"
		if len(service.Deployments) > 0 {
			deployTimeTxt = utils.FormatLocalDateTimeAmPmZone(*service.Deployments[0].CreatedAt)
		}

		taskCount := utils.I64ToString(*service.RunningCount)
		if *service.PendingCount > 0 {
			taskCount = fmt.Sprintf("%s (%d pending)", taskCount, *service.PendingCount)
		}
		if *service.DesiredCount != *service.RunningCount {
			taskCount = fmt.Sprintf("%s (%d desired)", taskCount, *service.DesiredCount)
		}

		return []string{
			*service.ServiceName,
			utils.RemoveAllRegex(`.*/`, *service.TaskDefinition),
			serviceImages,
			utils.LowerTitle(*service.Status),
			deployTimeTxt,
			taskCount,
		}
	}).([][]string)

	ui.AddTableConfigData(tableInfo, 1, data, tcell.ColorWhite)
	servicesColumnStyle := tcell.StyleDefault.Bold(true).Foreground(tcell.ColorWhite)
	ui.SetColumnStyle(tableInfo.Table, 0, 1, servicesColumnStyle)
}
