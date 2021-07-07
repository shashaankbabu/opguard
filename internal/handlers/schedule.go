package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/shashaankramesh/opguard/internal/helpers"
	"github.com/shashaankramesh/opguard/internal/models"
)

type ByHost []models.Schedule

// Len is used to sort ByHost
func (a ByHost) Len() int {
	return len(a)
}

// Less is used to sort ByHost
func (a ByHost) Less(i, j int) bool {
	return a[i].Host < a[j].Host
}

// Swap is used to sort ByHost
func (a ByHost) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// ListEntries lists schedule entries
func (repo *DBRepo) ListEntries(w http.ResponseWriter, r *http.Request) {
	var items []models.Schedule

	for k, v := range repo.App.MonitorMap {
		var item models.Schedule

		item.ID = k
		item.EntryID = v
		item.Entry = repo.App.Scheduler.Entry(v)
		hs, err := repo.DB.GetHostServiceByID(k)
		if err != nil {
			log.Println(err)
			return
		}
		item.ScheduleText = fmt.Sprintf("@every %d%s", hs.ScheduleNumber, hs.ScheduleUnit)
		item.LastRunFromHS = hs.LastCheck
		item.Host = hs.HostName
		item.Service = hs.Service.ServiceName

		items = append(items, item)
	}

	// Sort the slice by Host
	sort.Sort(ByHost(items))

	data := make(jet.VarMap)
	data.Set("items", items)

	err := helpers.RenderPage(w, r, "schedule", data, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
