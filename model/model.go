package model

import (
	"time"

	guuid "github.com/google/uuid"
)

type MongoConnection struct {
	ConnectionString string `yaml:"connectionString" envconfig:"MONGO_CONNECTION_STRING"`
	Database         string `yaml:"database" envconfig:"MONGO_DATABASE"`
}

type AWXConnection struct {
	Host     string `yaml:"host" envconfig:"AWX_HOST"`
	User     string `yaml:"user" envconfig:"AWX_USER"`
	Password string `yaml:"password" envconfig:"AWX_PASSWORD"`
}

type OIDConnection struct {
	Name              string `yaml:"name" envconfig:"OIDC_NAME"`
	DiscoveryEndpoint string `yaml:"discoveryEndpoint" envconfig:"OIDC_DISCOVERY_ENDPOINT"`
	ClientID          string `yaml:"clientID" envconfig:"OIDC_CLIENT_ID"`
	ClientSecret      string `yaml:"clientSecret" envconfig:"OIDC_CLIENT_SECRET"`
	RedirectURL       string `yaml:"redirectURL" envconfig:"OIDC_REDIRECT_URL"`
}

type Config struct {
	Mongo         MongoConnection `yaml:"mongoConnection"`
	AWX           AWXConnection   `yaml:"awxConnection"`
	OIDC          OIDConnection   `yaml:"oidConnection"`
	AdminPassword string          `yaml:"adminPassword"`
}

type Request struct {
	ID            guuid.UUID `json:"id" bson:"id"`
	UserID        string     `json:"user_id" bson:"user_id"`
	RequestReason string     `json:"request_reason" bson:"request_reason"`
	Reason        string     `json:"reason" bson:"reason"`
	State         string     `json:"state" bson:"state"`
	LastMessage   string     `json:"last_message" bson:"last_message"`
	Messages      []string   `json:"messages" bson:"messages"`
	JudgeID       string     `json:"judge_id" bson:"judge_id"`
	TemplateID    int        `json:"template_id" bson:"template_id"`
	Template      Template   `json:"template" bson:"template"`
	CreatedAt     time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" bson:"updated_at"`
}

type Pagination struct {
	Count    int         `json:"count"`
	Next     interface{} `json:"next"`
	Previous interface{} `json:"previous"`
}

type ListTemplatesResponse struct {
	Pagination
	Results []Template `json:"results"`
}

type Template struct {
	ID          int       `json:"id" bson:"id"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	IconLink    string    `json:"icon_link" bson:"icon_link"`
	Icon        string    `json:"icon" bson:"icon"`
	Survey      []Survey  `json:"spec" bson:"spec"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

type Survey struct {
	Choices             string `json:"choices" bson:"choices"`
	Default             string `json:"default" bson:"default"`
	Max                 string `json:"max" bson:"max"`
	Min                 string `json:"min" bson:"min"`
	NewQuestion         string `json:"new_question" bson:"new_question"`
	QuestionDescription string `json:"question_description" bson:"question_description"`
	QuestionName        string `json:"question_name" bson:"question_name"`
	Type                string `json:"type" bson:"type"`
	Variable            string `json:"variable" bson:"variable"`
	Required            bool   `json:"required" bson:"required"`
	RegEx               string `json:"regex" bson:"regex"`
	Value               string `json:"value" bson:"value"`
}

type Cart struct {
	UserID    string    `json:"user_id" bson:"user_id"`
	Requests  []Request `json:"requests" bson:"requests"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
