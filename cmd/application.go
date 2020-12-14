package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/swartzrock/ecsview/cmd/ecsview"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/thoas/go-funk"

	"github.com/swartzrock/ecsview/cmd/aws"
	"github.com/swartzrock/ecsview/cmd/pages"
	"github.com/swartzrock/ecsview/cmd/ui"
	"github.com/swartzrock/ecsview/cmd/utils"
)

var tviewApp *tview.Application
var clusterTable *tview.Table
var clusterDetailsPages *tview.Pages
var clusterDetailsPageMap = make(map[int32]*pages.ClusterDetailsPage)
var commandFooterBar *tview.TextView
var progressFooterBar *tview.TextView

// Entrypoint for the ecsview application
func Entrypoint() {
	buildUIElements()
}

// Select a cluster details page with a single key shortcut
func selectClusterDetailsPageByKey(key int32) bool {
	if page, found := clusterDetailsPageMap[key]; found {
		showPage(page, key)
		return true
	}
	return false
}

// Show the selected cluster detail page
func showPage(selectedPage *pages.ClusterDetailsPage, key int32) {

	_, frontPageView := clusterDetailsPages.GetFrontPage()

	cluster := getCurrentlySelectedCluster()
	if cluster == nil {
		return
	}
	ecsData := ecsview.GetClusterData(cluster)
	commandFooterBar.Highlight(string(key)).ScrollToHighlight()
	selectedPage.Render(ecsData)
	clusterDetailsPages.SwitchToPage(selectedPage.Name)
	showRefreshTime(*ecsData.Cluster.ClusterName, ecsData.Refreshed)

	// If the page about to be hidden has focus, switch focus to the new page
	if frontPageView != nil && frontPageView.HasFocus() {
		frontPageView.Blur()
		selectedPage.GetTable().Focus(nil)
	}
}

// Render the currently viewed cluster details page (eg in case the user selects a different cluster to view)
func renderCurrentClusterDetailsPage() {
	highlights := commandFooterBar.GetHighlights()
	if len(highlights) > 0 {
		var key = int32(highlights[0][0])
		selectClusterDetailsPageByKey(key)
	}
}

// Change focus between the cluster table and the cluster details page
func changeFocus() {
	_, pageView := clusterDetailsPages.GetFrontPage()
	if pageView == nil {
		return
	}
	if pageView.HasFocus() {
		pageView.Blur()
		clusterTable.Focus(nil)
	} else {
		clusterTable.Blur()
		pageView.Focus(nil)
	}
}

// Handle a user input event
func handleAppInput(event *tcell.EventKey) *tcell.EventKey {

	if event.Key() == tcell.KeyTab {
		changeFocus()
	}

	if event.Key() == tcell.KeyRune {
		key := event.Rune()

		if selectClusterDetailsPageByKey(key) {
			return event
		}

		if key == 'r' || key == 'R' {
			cluster := getCurrentlySelectedCluster()
			if cluster != nil {
				ecsview.RefreshClusterData(getCurrentlySelectedCluster())
				renderCurrentClusterDetailsPage()
			}
		}
	}

	return event
}

// Get the ECS cluster currently selected in the cluster table
func getCurrentlySelectedCluster() *aws.EcsCluster {

	// Check if there are no clusters found
	if clusterTable.GetRowCount() == 1 {
		return nil
	}

	selectedRow, _ := clusterTable.GetSelection()
	if selectedRow < 1 {
		selectedRow = 1
	}
	cell := clusterTable.GetCell(selectedRow, 0)
	cluster := cell.GetReference().(*aws.EcsCluster)
	return cluster
}

func showRefreshTime(what string, when time.Time) {
	progressFooterBar.Clear()
	fmt.Fprintf(progressFooterBar, "%s refreshed at %s", what, utils.FormatLocalTimeAmPmSecs(when))
}

// Build the UI elements for this application
func buildUIElements() {

	clusterTable = buildClusterTable()

	// Build the cluster detail pages and add their view shortcuts
	clusterDetailsPageMap['1'] = pages.NewServicesPage()
	clusterDetailsPageMap['2'] = pages.NewTasksPage()
	clusterDetailsPageMap['3'] = pages.NewInstancesPage()
	clusterDetailsPages = tview.NewPages()
	for _, page := range clusterDetailsPageMap {
		clusterDetailsPages.AddPage(page.Name, page.GetTable(), true, false)
	}

	commandFooterBar = buildCommandFooterBar()

	progressFooterBar = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetTextAlign(ui.R)
	progressFooterBar.SetBorderPadding(0, 0, 1, 2)

	if clusterTable.GetRowCount() == 1 {
		commandFooterBar.Clear()
		fmt.Fprint(commandFooterBar, "No clusters found")
	}

	footer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(commandFooterBar, 0, 6, false).
		AddItem(progressFooterBar, 0, 4, false)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(clusterTable, 10, 1, true).
		AddItem(clusterDetailsPages, 0, 1, false).
		AddItem(footer, 1, 1, false)

	tviewApp = tview.NewApplication().
		SetInputCapture(handleAppInput)

	// Show the services page
	selectClusterDetailsPageByKey('1')

	if err := tviewApp.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}

}

// Load the ECS clusters and create the cluster table
func buildClusterTable() *tview.Table {

	table := tview.NewTable().
		SetFixed(5, 6).
		SetSelectable(true, false)
	table.
		SetBorder(true).
		SetBorderColor(tcell.ColorDimGray).
		SetTitle(" ✨ ECS Clusters ").
		SetBorderPadding(0, 0, 1, 1)

	table.SetSelectionChangedFunc(func(row, column int) {
		renderCurrentClusterDetailsPage()
	})

	expansions := []int{2, 1, 1, 1, 1, 1, 1, 1}
	alignment := []int{ui.L, ui.L, ui.L, ui.R, ui.R, ui.R, ui.C, ui.C}

	headers := []string{"Name", "Status", "Type", "Instances", "Services", "Tasks", "CPU", "Memory"}
	ui.AddTableData(table, 0, [][]string{headers}, alignment, expansions, tcell.ColorYellow, false)

	ecsClusters := ecsview.GetClusters()
	if len(ecsClusters) == 0 {
		return table
	}

	data := funk.Map(ecsClusters, func(cluster *aws.EcsCluster) []string {

		containers := ecsview.GetClusterContainers(cluster)
		usage := aws.EcsContainerStats{}
		funk.ForEach(containers, func(c *aws.EcsContainer) { usage.Add(c.GetStats()) })
		meterWidth := 10
		cpuMeter := utils.BuildAsciiMeterCurrentTotal(usage.CpuUsed, usage.CpuTotal, meterWidth)
		memoryMeter := utils.BuildAsciiMeterCurrentTotal(usage.MemoryUsed, usage.MemoryTotal, meterWidth)

		return []string{
			*cluster.ClusterName,
			utils.LowerTitle(*cluster.Status),
			cluster.GetClusterType(),
			utils.I64ToString(*cluster.RegisteredContainerInstancesCount),
			utils.I64ToString(*cluster.ActiveServicesCount),
			utils.I64ToString(*cluster.RunningTasksCount),
			cpuMeter,
			memoryMeter,
		}
	}).([][]string)
	ui.AddTableData(table, 1, data, alignment, expansions, tcell.ColorWhite, true)

	// Mark the cpu and memory columns with dark cyan color and make them non-selectable
	usageMeterStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkCyan)
	for _, columnName := range []string{"CPU", "Memory"} {
		column := funk.IndexOfString(headers, columnName)
		ui.SetColumnStyle(table, column, 1, usageMeterStyle)
		for row := 1; row < table.GetRowCount(); row++ {
			table.GetCell(row, column).SetSelectable(false)
		}
	}

	// Add a reference to the Cluster to column 0 in each row for easy access later on
	for row, cluster := range ecsClusters {
		table.GetCell(row+1, 0).SetReference(cluster)
	}

	return table
}

// Build the command bar with detail page shortcuts that appears in the footer
func buildCommandFooterBar() *tview.TextView {

	footerBar := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	pageCommands := make([]string, 0)
	for key, page := range clusterDetailsPageMap {
		pageCommands = append(pageCommands, fmt.Sprintf(`[bold]%c ["%c"][darkcyan]%s[white][""]`, key, key, page.Name))
	}
	sort.Strings(pageCommands)

	footerPageText := strings.Join(pageCommands, " ")
	footerPageText = fmt.Sprintf(`%s %c [white::b]R[darkcyan::-] Refresh-Data`, footerPageText, tcell.RuneVLine)
	footerPageText = fmt.Sprintf(`%s [white::b]Tab[darkcyan::-] Next-View`, footerPageText)

	fmt.Fprint(footerBar, footerPageText)

	return footerBar
}
