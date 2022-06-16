package main

import (
	"fmt"
	"github.com/VitalFrog/vitalfrog-go-client/api_client"
	"github.com/alecthomas/kong"
	log "github.com/sirupsen/logrus"
	"strings"
)

var cli struct {
	APIBaseUrl string `kong:"default='https://api.vitalfrog.com/v2',env='API_BASE_URL',help='API Address of the VitalFrog api'"`
	APIToken   string `kong:"required,env='API_TOKEN',help='Your VitalFrog api token'"`

	AllowedCountries []string `kong:"env='ALLOWED_COUNTRIES',help='Which countries to test from. Either ALLOWED_COUNTRIES or BlOCKED_COUNTRIES can be set, not both.'"`
	BlockedCountries []string `kong:"env='BlOCKED_COUNTRIES',help='Which countries NOT to test from. Either ALLOWED_COUNTRIES or BlOCKED_COUNTRIES can be set, not both.'"`

	PerformanceBudgetsId int32    `kong:"env='PERFORMANCE_BUDGETS_ID',help='Performance budgets to use. If not defined falls back to VitalFrog default.'"`
	Devices              []string `kong:"env='DEVICES',help='Which devices you want to test for. If not set falls back to desktop & mobile'"`

	TargetHost       string   `kong:"required,env='TARGET_HOST',help='Host of url you want to test'"`
	TargetSchemeHost string   `kong:"default='https',enum='https,http',env='TARGET_SCHEMA',help='What schema (http|https) to use on target host'"`
	TargetPaths      []string `kong:"required,env='TARGET_PATHS',help='Paths to test'"`

	BasicAuthUsername string `kong:"env='BASIC_AUTH_USERNAME',help='Username to use for basic auth. If configured, then BASIC_AUTH_PASSWORD must also be set'"`
	BasicAuthPassword string `kong:"env='BASIC_AUTH_PASSWORD',help='Password to use for basic auth. If configured, then BASIC_AUTH_USERNAME must also be set'"`

	ExtraHeaders map[string]string `kong:"env='EXTRA_HEADERS',help='Additional headers to set on the request. Mostly used for auth reasons'"`

	RunAsync bool   `kong:"env='RUN_ASYNC',help='Configure if the request should run async, to not block execution. Report must be checked in browser then later'"`
	LogLevel string `kong:"default='info',enum='error,info,debug',env='LOG_LEVEL',help='Log level'"`
}

func main() {
	kong.Parse(&cli)
	err := cliCheck()
	if err != nil {
		log.Fatalf(err.Error())
	}

	c := api_client.New(cli.APIBaseUrl, cli.APIToken)

	reportConfig := api_client.ReportConfig{
		Countries:            nil,
		Devices:              nil,
		PerformanceBudgetsId: nil,
		Http:                 nil,
		Target: api_client.Target{
			Host:   cli.TargetHost,
			Scheme: &cli.TargetSchemeHost,
			Paths: api_client.ManualPathSelection{
				Mode:  "manual",
				Paths: cli.TargetPaths,
			},
		},
	}
	if len(cli.AllowedCountries) > 0 || len(cli.BlockedCountries) > 0 {
		// Configure allow list
		newCountries := api_client.Countries{
			List: make([]api_client.Country, 0),
			Mode: api_client.AllowList,
		}
		if len(cli.BlockedCountries) > 0 {
			newCountries.Mode = api_client.BlockList
		}
		for _, c := range cli.AllowedCountries {
			newCountries.List = append(newCountries.List, api_client.Country{
				Code: c,
			})
		}
		reportConfig.Countries = &newCountries
	}
	if len(cli.Devices) > 0 {
		newDevices := make([]api_client.Device, 0)
		for _, d := range cli.Devices {
			switch d {
			case string(api_client.Mobile):
				newDevices = append(newDevices, api_client.Device{Name: api_client.Mobile})
			case string(api_client.Desktop):
				newDevices = append(newDevices, api_client.Device{Name: api_client.Desktop})
			}
		}
		reportConfig.Devices = &newDevices
	}
	if cli.PerformanceBudgetsId != 0 {
		reportConfig.PerformanceBudgetsId = &cli.PerformanceBudgetsId
	}
	if cli.BasicAuthPassword != "" || len(cli.ExtraHeaders) > 0 {
		newHttp := api_client.HttpConfig{}
		if cli.BasicAuthPassword != "" {
			newHttp.BasicAuth = &api_client.BasicAuth{
				Password: cli.BasicAuthPassword,
				Username: cli.BasicAuthUsername,
			}
		}
		if len(cli.ExtraHeaders) > 0 {
			extraHeaders := make(api_client.ExtraHeadersConfig, 0)
			for header, value := range cli.ExtraHeaders {
				extraHeaders = append(extraHeaders, api_client.Header{
					Header: header,
					Value:  value,
				})
			}
			newHttp.ExtraHeaders = &extraHeaders
		}
		reportConfig.Http = &newHttp
	}

	if cli.RunAsync {
		// Create async report
		report, err := c.CreateAsyncReport(reportConfig)
		if err != nil {
			log.Fatalf("Could not CreateAsyncReport: ", err)
		}
		fmt.Println(report)
	} else {
		// Create sync report
		report, err := c.CreateSyncReport(reportConfig)
		if err != nil {
			log.Fatalf("Could not CreateAsyncReport: ", err)
		}
		fmt.Println(report)
	}

}

func cliCheck() error {
	switch cli.LogLevel {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		return fmt.Errorf("invalid LOG_LEVEL: %q", cli.LogLevel)
	}

	if (cli.BasicAuthUsername == "" && cli.BasicAuthPassword != "") || (cli.BasicAuthUsername != "" && cli.BasicAuthPassword == "") {
		return fmt.Errorf("both BASIC_AUTH_PASSWORD and BASIC_AUTH_USERNAME must be configure if one of them is set")
	}
	if len(cli.TargetPaths) == 0 {
		return fmt.Errorf("at least 1 TARGET_PATH must be set")
	}

	if cli.TargetHost == "" {
		return fmt.Errorf("env var TARGET_HOST must be set")
	}

	if len(cli.AllowedCountries) > 0 && len(cli.BlockedCountries) > 0 {
		return fmt.Errorf("either ALLOWED_COUNTRIES or BlOCKED_COUNTRIES can be set, not both")
	}

	if strings.HasSuffix(cli.APIBaseUrl, "/") {
		return fmt.Errorf("either API_BASE_URL must not have '/' suffix")
	}

	return nil
}
