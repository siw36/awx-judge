package awx

import (
	"encoding/json"
	"net/http"
	"strconv"

	model "../model"
	log "github.com/Sirupsen/logrus"
)

var (
	Config model.Config
)

func GetTemplates() ([]model.Template, error) {
	client := &http.Client{}
	var jobTemplates model.ListTemplatesResponse
	request, err := http.NewRequest("GET", Config.AWX.Host+"/api/v2/job_templates/?page_size=10000", nil)
	if err != nil {
		return jobTemplates.Results, err
	}
	request.SetBasicAuth(Config.AWX.User, Config.AWX.Password)
	log.Info("Getting all job templates from AWX")
	response, err := client.Do(request)
	if err != nil {
		return jobTemplates.Results, err
	}
	json.NewDecoder(response.Body).Decode(&jobTemplates)
	defer response.Body.Close()
	return jobTemplates.Results, err
}

func GetTemplate(ID int) (model.Template, error) {
	client := &http.Client{}
	var jobTemplate model.Template
	request, err := http.NewRequest("GET", Config.AWX.Host+"/api/v2/job_templates/"+strconv.Itoa(ID), nil)
	if err != nil {
		return jobTemplate, err
	}
	request.SetBasicAuth(Config.AWX.User, Config.AWX.Password)
	log.Info("Getting job template " + strconv.Itoa(ID) + " from AWX")
	response, err := client.Do(request)
	if err != nil {
		return jobTemplate, err
	}
	json.NewDecoder(response.Body).Decode(&jobTemplate)
	defer response.Body.Close()
	return jobTemplate, err
}

func GetSurvey(ID int) ([]model.Survey, error) {
	client := &http.Client{}
	var template model.Template
	request, err := http.NewRequest("GET", Config.AWX.Host+"/api/v2/job_templates/"+strconv.Itoa(ID)+"/survey_spec/", nil)
	if err != nil {
		return template.Survey, err
	}
	request.SetBasicAuth(Config.AWX.User, Config.AWX.Password)
	log.Info("Getting survey spec from AWX for job template " + strconv.Itoa(ID))
	response, err := client.Do(request)
	if err != nil {
		return template.Survey, err
	}
	json.NewDecoder(response.Body).Decode(&template)
	defer response.Body.Close()
	return template.Survey, err
}

func JobTemplateLaunch(id int, extraVars []byte) (jobID int, err error) {
	log.Info("Launching job template  " + strconv.Itoa(id))
	var response model.JobTemplateLaunchResponse
	client := &http.Client{}
	request, err := http.NewRequest("POST", Config.AWX.Host+"/api/v2/job_templates/"+strconv.Itoa(id)+"/launch/", nil)
	request.SetBasicAuth(Config.AWX.User, Config.AWX.Password)
	if err != nil {
		log.Error("Failed to construct request for launch job template with error: ", err)
		return 0, err
	}
	responseRaw, err := client.Do(request)
	if err != nil {
		log.Error("Failed to launch job template with error: ", err)
		return 0, err
	}
	if responseRaw.StatusCode == 201 {
		json.NewDecoder(responseRaw.Body).Decode(&response)
		log.Info("Successfully launched job template  " + strconv.Itoa(id))
		defer responseRaw.Body.Close()
		return response.Job, err
	} else {
		log.Error("Failed to launch job template with error: ", responseRaw.Body)
		return 0, err
	}
}

// func JobGet(request model.Request) error {
//
// }
