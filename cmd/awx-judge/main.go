package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/siw36/awx-judge/internal/awx"
	"github.com/siw36/awx-judge/internal/bg"
	"github.com/siw36/awx-judge/internal/db"
	"github.com/siw36/awx-judge/internal/model"
	"github.com/siw36/awx-judge/internal/utils"
	"github.com/siw36/awx-judge/internal/web"
	"go.mongodb.org/mongo-driver/mongo"
)

var Client *mongo.Client

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	// Parse config
	log.Info("Parsing configuration")
	var Config model.Config
	utils.ReadConfigFile(&Config)
	utils.ReadConfigEnv(&Config)

	// Inject in other packages
	awx.Config = Config
	db.Config = Config
	web.Config = Config
	//oidcConnector.OIDConnection = Config.OIDC

	// Establish MongoDB connection
	var Client *mongo.Client
	Client = db.Connect(Config.Mongo.ConnectionString, Config.Mongo.Database)
	db.Client = Client

	// Download all icons
	go bg.DownloadAllIcons()
}

func main() {
	// // testing
	// request, _ := db.GetRequest("admin", guuid.MustParse("0997457f-59da-4faa-a1b0-7cf4aae38ad4"))
	go bg.JobLaunch()
	web.Serve()
}
