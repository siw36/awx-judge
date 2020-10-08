package main

import (
	awx "./awx"
	bg "./bg"
	db "./db"
	internal "./internal"
	model "./model"
	oidcConnector "./oidcConnector"
	web "./web"
	log "github.com/Sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

var Client *mongo.Client

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	// Parse config
	log.Info("Parsing configuration")
	var Config model.Config
	internal.ReadConfigFile(&Config)
	internal.ReadConfigEnv(&Config)

	// Inject in other packages
	awx.Config = Config
	db.Config = Config
	web.Config = Config
	oidcConnector.OIDConnection = Config.OIDC

	// Establish MongoDB connection
	var Client *mongo.Client
	Client = db.Connect(Config.Mongo.ConnectionString, Config.Mongo.Database)
	db.Client = Client
}

func main() {
	// // testing
	// request, _ := db.GetRequest("admin", guuid.MustParse("0997457f-59da-4faa-a1b0-7cf4aae38ad4"))
	go bg.JobLaunch()
	web.Serve()
}
