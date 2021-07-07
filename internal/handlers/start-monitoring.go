package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

// job is the unit of work to be performed
type job struct {
	HostServiceID int
}

// Run runs the scheduled jobs
func (j job) Run() {
	Repo.ScheduledCheck(j.HostServiceID)
}

func (repo *DBRepo) StartMonitoring() {
	if app.PreferenceMap["monitoring_live"] == "1" {
		// Trigger a message to broadcast to all clients that app is startting to monitor
		data := make(map[string]string)
		data["message"] = "Monitoring is on!"
		err := app.WsClient.Trigger("public-channel", "app-starting", data)
		if err != nil {
			log.Println(err)
		}

		// Get all of the services that we want to monitor from the database
		servicesToMonitor, err := repo.DB.GetServicesToMonitor()
		if err != nil {
			log.Println(err)
		}

		// Range through the services
		for _, x := range servicesToMonitor {
			// Get the schedule unit and number
			var sch string
			if x.ScheduleUnit == "d" {
				sch = fmt.Sprintf("@every %d%s", x.ScheduleNumber*24, "h")
			} else {
				sch = fmt.Sprintf("@every %d%s", x.ScheduleNumber, x.ScheduleUnit)
			}

			// Create a job
			var j job
			j.HostServiceID = x.ID
			scheduleID, err := app.Scheduler.AddJob(sch, j)
			if err != nil {
				log.Println(err)
			}

			// Save the ID of the job to start/stop it
			app.MonitorMap[x.ID] = scheduleID

			// Broadcast over websockets that the service is scheduled
			payload := make(map[string]string)
			payload["message"] = "scheduling"
			payload["host_service_id"] = strconv.Itoa(x.ID)
			yearOne := time.Date(0001, 11, 17, 20, 34, 58, 65138737, time.UTC)
			if app.Scheduler.Entry(app.MonitorMap[x.ID]).Next.After(yearOne) {
				payload["next_run"] = app.Scheduler.Entry(app.MonitorMap[x.ID]).Next.Format("2006-01-02 3:04:05 PM")
			} else {
				payload["next_run"] = "Pending..."
			}
			payload["host"] = x.HostName
			payload["service"] = x.Service.ServiceName
			if x.LastCheck.After(yearOne) {
				payload["last_run"] = x.LastCheck.Format("2006-01-02 3:04:05 PM")
			} else {
				payload["last_run"] = "Pending..."
			}
			payload["schedule"] = fmt.Sprintf("@every %d%s", x.ScheduleNumber, x.ScheduleUnit)

			err = app.WsClient.Trigger("public-channel", "next-run-event", payload)
			if err != nil {
				log.Println(err)
			}

			err = app.WsClient.Trigger("public-channel", "schedule-changed-event", payload)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
