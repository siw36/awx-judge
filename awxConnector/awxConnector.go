package awxConnector

import (
	"encoding/json"
	"net/http"

	model "../model"

	awxGo "github.com/Colstuwjx/awx-go"
	log "github.com/Sirupsen/logrus"
)

var (
	Config model.Config
)

func GetTemplates() []*awxGo.JobTemplate {
	log.Info("Get job templates from AWX")
	var awx = awxGo.NewAWX(Config.AWX.Host, Config.AWX.User, Config.AWX.Password, nil)
	result, _, err := awx.JobTemplateService.ListJobTemplates(map[string]string{"page_size": "10000"})
	if err != nil {
		log.Error(err)
		return nil
	}
	// for _, template := range result {
	// 	log.Info(template.Name)
	// }
	return result
}

func GetSurveySpec(ID string) (model.Survey, error) {
	client := &http.Client{}
	var surveySpec model.Survey
	request, err := http.NewRequest("GET", Config.AWX.Host+"/api/v2/job_templates/"+ID+"/survey_spec/", nil)
	if err != nil {
		return surveySpec, err
	}
	request.SetBasicAuth(Config.AWX.User, Config.AWX.Password)
	log.Info("Get survey spec for job template " + ID)
	response, err := client.Do(request)
	if err != nil {
		return surveySpec, err
	}
	surveySpec.ID = ID
	json.NewDecoder(response.Body).Decode(&surveySpec)
	defer response.Body.Close()
	return surveySpec, err
}
