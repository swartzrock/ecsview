package pages

import (
	"github.com/rivo/tview"

	"github.com/swartzrock/ecsview/cmd/ecsview"
	"github.com/swartzrock/ecsview/cmd/ui"
)

// Represents a page that displays details about an AWS cluster
type ClusterDetailsPage struct {
	Name      string
	TableInfo *ui.TableInfo
	Render    func(ecsData *ecsview.ClusterData)
}

func (p *ClusterDetailsPage) GetTable() *tview.Table {
	return p.TableInfo.Table
}
