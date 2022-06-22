package main

import (
	"fmt"
	"github.com/VitalFrog/vitalfrog-go-client/vfrogapi"
	"github.com/alecthomas/kong"
	log "github.com/sirupsen/logrus"
	"strings"
)

type config struct {
	APIBaseUrl string `kong:"default='https://api.vitalfrog.com/v2',env='API_BASE_URL',help='API Address of the VitalFrog api'"`
	APIToken   string `kong:"required,env='API_TOKEN',help='Your VitalFrog api token'"`

	AllowedCountries []string `kong:"env='ALLOWED_COUNTRIES',help='Which countries to test from. Either ALLOWED_COUNTRIES or BlOCKED_COUNTRIES can be set, not both.'"`
	BlockedCountries []string `kong:"env='BlOCKED_COUNTRIES',help='Which countries NOT to test from. Either ALLOWED_COUNTRIES or BlOCKED_COUNTRIES can be set, not both.'"`

	PerformanceBudgetsId int32    `kong:"env='PERFORMANCE_BUDGETS_ID',help='Performance budgets to use. If not defined falls back to VitalFrog default.'"`
	Devices              []string `kong:"env='DEVICES',help='Which devices you want to test for. If not set falls back to desktop & mobile'"`

	TargetHost       string   `kong:"required,env='TARGET_HOST',help='Host of url you want to test'"`
	TargetSchemeHost string   `kong:"default='https',enum='https,http',env='TARGET_SCHEMA',help='What schema (http|https) to use on target host'"`
	TargetPaths      []string `kong:"required,env='TARGET_PATHS',help='Paths to test'"`

	Version       string `kong:"env='VERSION',help='Version of the given code. Good for later tracing'"`
	ComponentName string `kong:"env='COMPONENT_NAME',help='Name of the component we are testing. Helps to figure out cross repo problems. Good for later tracing'"`

	BasicAuthUsername string `kong:"env='BASIC_AUTH_USERNAME',help='Username to use for basic auth. If configured, then BASIC_AUTH_PASSWORD must also be set'"`
	BasicAuthPassword string `kong:"env='BASIC_AUTH_PASSWORD',help='Password to use for basic auth. If configured, then BASIC_AUTH_USERNAME must also be set'"`

	ExtraHeaders map[string]string `kong:"env='EXTRA_HEADERS',help='Additional headers to set on the request. Mostly used for auth reasons'"`

	RunAsync bool   `kong:"env='RUN_ASYNC',help='Configure if the request should run async, to not block execution. Report must be checked in browser then later'"`
	LogLevel string `kong:"default='info',enum='error,info,debug',env='LOG_LEVEL',help='Log level'"`
}

func parseEnvironmentToConfig() (*config, error) {
	cfg := config{}

	kong.Parse(&cfg)
	err := cfg.check()
	if err != nil {
		return nil, fmt.Errorf("configCheck failed: %w", err)
	}
	return &cfg, nil
}

func (c config) check() error {
	switch c.LogLevel {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		return fmt.Errorf("invalid LOG_LEVEL: %q", c.LogLevel)
	}

	if (c.BasicAuthUsername == "" && c.BasicAuthPassword != "") || (c.BasicAuthUsername != "" && c.BasicAuthPassword == "") {
		return fmt.Errorf("both BASIC_AUTH_PASSWORD and BASIC_AUTH_USERNAME must be configure if one of them is set")
	}
	if len(c.TargetPaths) == 0 {
		return fmt.Errorf("at least 1 TARGET_PATH must be set")
	}

	if c.TargetHost == "" {
		return fmt.Errorf("env var TARGET_HOST must be set")
	}

	if len(c.AllowedCountries) > 0 && len(c.BlockedCountries) > 0 {
		return fmt.Errorf("either ALLOWED_COUNTRIES or BlOCKED_COUNTRIES can be set, not both")
	}

	if strings.HasSuffix(c.APIBaseUrl, "/") {
		return fmt.Errorf("API_BASE_URL must not have '/' suffix")
	}

	return nil
}

func (c config) ToReportConfig() vfrogapi.ReportConfig {
	// Create new performance report
	reportConfig := vfrogapi.ReportConfig{
		Countries:            nil,
		Devices:              nil,
		PerformanceBudgetsId: nil,
		Http:                 nil,
		Target: vfrogapi.Target{
			Host:   c.TargetHost,
			Scheme: &c.TargetSchemeHost,
			Paths: vfrogapi.ManualPathSelection{
				Mode:  "manual",
				Paths: c.TargetPaths,
			},
		},
	}

	if c.ComponentName != "" {
		reportConfig.Component = &c.ComponentName
	}
	if c.Version != "" {
		reportConfig.Version = &c.Version
	}

	// Configure allowed countries
	if len(c.AllowedCountries) > 0 || len(c.BlockedCountries) > 0 {
		newCountries := vfrogapi.Countries{
			List: make([]vfrogapi.Country, 0),
			Mode: vfrogapi.AllowList,
		}
		if len(c.BlockedCountries) > 0 {
			newCountries.Mode = vfrogapi.BlockList
		}
		for _, c := range c.AllowedCountries {
			newCountries.List = append(newCountries.List, vfrogapi.Country{
				Code: c,
			})
		}
		reportConfig.Countries = &newCountries
	}
	// Configure devices (if default is not enough)
	if len(c.Devices) > 0 {
		newDevices := make([]vfrogapi.Device, 0)
		for _, d := range c.Devices {
			switch d {
			case string(vfrogapi.Mobile):
				newDevices = append(newDevices, vfrogapi.Device{Name: vfrogapi.Mobile})
			case string(vfrogapi.Desktop):
				newDevices = append(newDevices, vfrogapi.Device{Name: vfrogapi.Desktop})
			}
		}
		reportConfig.Devices = &newDevices
	}
	// Set dedicated performance budget
	if c.PerformanceBudgetsId != 0 {
		reportConfig.PerformanceBudgetsId = &c.PerformanceBudgetsId
	}
	// Configure http data
	if c.BasicAuthPassword != "" || len(c.ExtraHeaders) > 0 {
		newHttp := vfrogapi.HttpConfig{}
		if c.BasicAuthPassword != "" {
			newHttp.BasicAuth = &vfrogapi.BasicAuth{
				Password: c.BasicAuthPassword,
				Username: c.BasicAuthUsername,
			}
		}
		if len(c.ExtraHeaders) > 0 {
			extraHeaders := make(vfrogapi.ExtraHeadersConfig, 0)
			for header, value := range c.ExtraHeaders {
				extraHeaders = append(extraHeaders, vfrogapi.Header{
					Header: header,
					Value:  value,
				})
			}
			newHttp.ExtraHeaders = &extraHeaders
		}
		reportConfig.Http = &newHttp
	}

	return reportConfig

}
