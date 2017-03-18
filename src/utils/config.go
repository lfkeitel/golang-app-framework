package utils

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
)

var packageName = "netmon"
var configEnvVar = "NETMON_CONFIG"
var configFileLocations = []string{
	"config.toml",
	"config/config.toml",
	os.ExpandEnv("$HOME/.pg/config.toml"),
	"/etc/" + packageName + "/config.toml",
}

// Config defines the configuration struct for the application
type Config struct {
	sourceFile string
	Core       struct {
		SiteTitle       string
		SiteCompanyName string
		SiteDomainName  string
		SiteFooterText  string
	}
	Logging struct {
		Enabled    bool
		EnableHTTP bool
		Level      string
		Path       string
	}
	Database struct {
		Type         string
		Address      string
		Port         int
		Username     string
		Password     string
		Name         string
		Retry        int
		RetryTimeout string
	}
	Webserver struct {
		Address             string
		HttpPort            int
		HttpsPort           int
		TLSCertFile         string
		TLSKeyFile          string
		RedirectHttpToHttps bool
		SessionStore        string
		SessionName         string
		SessionsDir         string
		SessionsAuthKey     string
		SessionsEncryptKey  string
	}
}

func FindConfigFile() string {
	if os.Getenv(configEnvVar) != "" && FileExists(os.Getenv(configEnvVar)) {
		return os.Getenv(configEnvVar)
	}

	for _, path := range configFileLocations {
		if FileExists(path) {
			return path
		}
	}
	return ""
}

func NewEmptyConfig() *Config {
	return &Config{}
}

func NewConfig(configFile string) (conf *Config, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()

	if configFile == "" {
		configFile = "config.toml"
	}

	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var con Config
	if err := toml.Unmarshal(buf, &con); err != nil {
		return nil, err
	}
	con.sourceFile = configFile
	return setSensibleDefaults(&con)
}

func setSensibleDefaults(c *Config) (*Config, error) {
	// Anything not set here implies its zero value is the default

	// Core
	c.Core.SiteTitle = setStringOrDefault(c.Core.SiteTitle, "My App")
	c.Core.SiteFooterText = setStringOrDefault(c.Core.SiteFooterText, "My app footer")

	// Logging
	c.Logging.Level = setStringOrDefault(c.Logging.Level, "notice")

	// Database
	c.Database.Type = setStringOrDefault(c.Database.Type, "sqlite")
	c.Database.Address = setStringOrDefault(c.Database.Address, "database.sqlite3")
	c.Database.RetryTimeout = setStringOrDefault(c.Database.RetryTimeout, "1m")

	// Webserver
	c.Webserver.HttpPort = setIntOrDefault(c.Webserver.HttpPort, 8080)
	c.Webserver.HttpsPort = setIntOrDefault(c.Webserver.HttpsPort, 1443)
	c.Webserver.SessionName = setStringOrDefault(c.Webserver.SessionName, "my-app")
	c.Webserver.SessionsDir = setStringOrDefault(c.Webserver.SessionsDir, "sessions")
	c.Webserver.SessionStore = setStringOrDefault(c.Webserver.SessionStore, "filesystem")
	return c, nil
}

// Given string s, if it is empty, return v else return s.
func setStringOrDefault(s, v string) string {
	if s == "" {
		return v
	}
	return s
}

// Given integer s, if it is 0, return v else return s.
func setIntOrDefault(s, v int) int {
	if s == 0 {
		return v
	}
	return s
}

func (c *Config) Reload() error {
	con, err := NewConfig(c.sourceFile)
	if err != nil {
		return err
	}
	c = con
	return nil
}
