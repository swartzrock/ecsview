package pages

import (
	"strconv"

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

// Prepend every slice in data with a row-number value
func PrependRowNumColumn(data [][]string) [][]string {
	for i := 0; i < len(data); i++ {
		row := strconv.Itoa(i + 1)
		data[i] = append([]string{row}, data[i]...)
	}
	return data
}
