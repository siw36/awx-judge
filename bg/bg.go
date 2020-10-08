package bg

import (
	"time"

	awx "../awx"
	db "../db"
	internal "../internal"
	model "../model"
	log "github.com/Sirupsen/logrus"
)

func JobLaunch() {
	log.Info("Starting background job launcher")
	// Get all approved requests
	requests, err := db.GetRequestsByState("approved")
	if err != nil {
		log.Info("No approved requests found")
	}
	for _, request := range requests {
		// Update the status and message
		request.State = "Running"
		request.LastMessage = "Launched job template on Tower/AWX"
		request.Messages = append(request.Messages, request.LastMessage)
		request.UpdatedAt = time.Now()
		err = db.UpdateRequest(request)
		if err != nil {
			log.Error("Failed to update request ", request.ID)
			continue
		}
		// Construct extra vars
		log.Info("Launching request ", request.ID)
		extraVars, err := internal.ExtraVars(request)
		if err != nil {
			request.State = "Error"
			request.LastMessage = "Failed to construct extra vars"
			request.Messages = append(request.Messages, request.LastMessage)
			request.UpdatedAt = time.Now()
			err = db.UpdateRequest(request)
			if err != nil {
				log.Error("Failed to update request ", request.ID)
			}
			continue
		}
		// Start the job on Tower/AWX
		jobID, err := awx.JobTemplateLaunch(request.TemplateID, extraVars)
		if err != nil {
			request.State = "Error"
			request.LastMessage = "Failed to launch job template"
			request.Messages = append(request.Messages, request.LastMessage)
			request.UpdatedAt = time.Now()
			err = db.UpdateRequest(request)
			if err != nil {
				log.Error("Failed to update request ", request.ID)
			}
			continue
		}
		// Init JobWatch for each launched job
		go JobWatch(jobID, request)
	}
	log.Info("Background job launcher is now waiting")
	time.Sleep(1 * time.Minute)
	JobLaunch()
}

func JobWatch(id int, request model.Request) {
	log.Info("Starting watch on job ", id)
	var job model.Job
	// null timestamp
	unfinished := time.Time{}
	for job.Finished == unfinished {
		job, _ := awx.JobGet(id)
		if job.Status == "successful" {
			request.State = "Success"
			request.LastMessage = "Job template run completed successfully"
			request.Messages = append(request.Messages, request.LastMessage)
			request.UpdatedAt = time.Now()
			err := db.UpdateRequest(request)
			if err != nil {
				log.Error("Failed to update request ", request.ID)
			}
			return
		} else if job.Status == "running" || job.Status == "pending" {
			log.Info("Job template is still running")
			time.Sleep(20 * time.Second)
			continue
		} else {
			request.State = "Failed"
			request.LastMessage = "Job template run failed"
			request.Messages = append(request.Messages, request.LastMessage)
			request.UpdatedAt = time.Now()
			err := db.UpdateRequest(request)
			if err != nil {
				log.Error("Failed to update request ", request.ID)
			}
			log.Error(request.LastMessage)
			return
		}
	}
	log.Info("Finished watch on job ", id)
}