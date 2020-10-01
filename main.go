package main

import (
	awx "./awx"
	helper "./helper"
	model "./model"
	db "./db"
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
	helper.ReadConfigFile(&Config)
	helper.ReadConfigEnv(&Config)

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
	// awx.LaunchJob(request)

	web.Serve()
}
