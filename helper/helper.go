package helper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	model "../model"

	log "github.com/Sirupsen/logrus"
	envconfig "github.com/kelseyhightower/envconfig"
	yaml "gopkg.in/yaml.v2"
)

func ReadConfigFile(cfg *model.Config) {
	f, err := os.Open("config.yaml")
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
	var pv string = "www/static/icons"
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
	log.Info(toTest)
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	return true
}
