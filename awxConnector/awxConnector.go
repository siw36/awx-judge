package awxConnector

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
