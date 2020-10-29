package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/siw36/awx-judge/internal/model"

	log "github.com/sirupsen/logrus"
	envconfig "github.com/kelseyhightower/envconfig"
	yaml "gopkg.in/yaml.v2"
)

var Root string = RootDir()

func ReadConfigFile(cfg *model.Config) {
	f, err := os.Open("../../configs/config.yaml")
	if err != nil {
		log.Error(err)
		return
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		log.Error(err)
		return
	}
}

func ReadConfigEnv(cfg *model.Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Error(err)
		return
	}
}

func HttpClient() *http.Client {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	return &client
}

func DownloadIcon(id int, link string) (err error, icon string) {
	var pv string = "web/static/icons"
	if _, err := os.Stat(pv); os.IsNotExist(err) {
		return err, ""
	}
	fileURL, err := url.Parse(link)
	if err != nil {
		return err, ""
	}

	path := fileURL.Path
	segments := strings.Split(path, ".")
	fileName := fmt.Sprintf("%s/%s.%s", pv, strconv.Itoa(id), segments[len(segments)-1])
	file, err := os.Create(fileName)
	if err != nil {
		return err, ""
	}
	client := HttpClient()

	// Download the file to PV
	resp, err := client.Get(link)
	if err != nil {
		return err, ""
	}

	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)
	defer file.Close()
	if err != nil {
		return err, ""
	}

	log.Infof("Downloaded icon %s with size %s", fileName, strconv.FormatInt(int64(size), 10))

	return err, strings.Replace(fileName, "www/", "", -1)
}

func JsonResponse(w http.ResponseWriter, data interface{}) {
	// Construt json response
	log.Debug("Constructing json response")
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Write json response
	log.Debug("Sending json response")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ValidUrl tests a string to determine if it is a well-structured url or not.
func ValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}
	return true
}

func ExtraVars(request model.Request) ([]byte, error) {
	// Format survey variables
	log.Info("Building extra_vars json from request")
	var surveyVars model.SurveyVars
	m := make(map[string]interface{})
	for _, item := range request.Template.Survey {
		m[item.Variable] = item.Value
	}
	surveyVars.ExtraVars = m
	jsonString, err := json.Marshal(surveyVars)
	if err != nil {
		log.Error(err)
		return jsonString, err
	}
	return jsonString, err
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}
