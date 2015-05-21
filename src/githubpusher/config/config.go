package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

type ServerConfig struct {
	Port        string `json:"port"`
	WebhookPort string `json:"webhookPort"`

	CertFilename string `json:"certFilename"`
	KeyFilename  string `json:"keyFilename"`

	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	OAuthState   string `json:"oauthState"`

	CookieSecret string `json:"cookieSecret"`

	PackagePath string `json:"packagePath"`

	MongoAddress string `json:"mongoAddress"`
	GCPApiKey    string `json:"gcmAPIKey"`
}

var (
	configFile   *string = flag.String("config", "config.json", "Configuration File")
	serverConfig *ServerConfig
)

func ReadConfig() *ServerConfig {
	if serverConfig != nil {
		return serverConfig
	}

	var data []byte
	var err error

	data, err = ioutil.ReadFile(*configFile)
	if err != nil {
		log.Println("Not configured.  Could not find config.json")
		os.Exit(-1)
	}

	serverConfig = new(ServerConfig)

	err = json.Unmarshal(data, serverConfig)
	if err != nil {
		log.Println("Could not unmarshal config.json", err)
		os.Exit(-1)
	}
	return serverConfig
}
