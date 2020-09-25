package main

import (
	awxConnector "./awxConnector"
	helper "./helper"
	model "./model"
	mongoConnector "./mongoConnector"
	oidcConnector "./oidcConnector"
	webServer "./webServer"
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
	awxConnector.Config = Config
	mongoConnector.Config = Config
	webServer.Config = Config
	oidcConnector.OIDConnection = Config.OIDC

	// Establish MongoDB connection
	var Client *mongo.Client
	Client = mongoConnector.DBConnect(Config.Mongo.ConnectionString, Config.Mongo.Database)
	mongoConnector.Client = Client
}

func main() {
	webServer.Serve()
}
