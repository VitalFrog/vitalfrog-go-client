package vfrogapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vitalfrog/jsonl"
	"io/ioutil"
	"net/http"
)

// Client is a http client to access the VitalFrog api with a few convenience functions
type Client struct {
	hClient  *http.Client
	baseUrl  string
	apiToken string
}

// New creates a new VitalFrog api client
func New(baseUrl, apiToken string) Client {
	return Client{
		hClient:  &http.Client{},
		baseUrl:  baseUrl,
		apiToken: apiToken,
	}
}

// GetPerformanceBudgets GETs the performance budgets with the given id
func (c Client) GetPerformanceBudgets(performanceBudgetsId int32) (*PerformanceBudgets, error) {
	budgets := &PerformanceBudgets{}
	err := c.getJSON(fmt.Sprintf("/performance_budgets/%d", performanceBudgetsId), budgets)
	if err != nil {
		return nil, fmt.Errorf("could not getJSON: %w", err)
	}
	return budgets, nil
}

// CreateSyncReport start a new performance report via the VitalFrog api. Loading it partially the results
// are returned into the output channel
func (c Client) CreateSyncReport(config ReportConfig) (*ReportMetadata, chan PerformanceReport, error) {
	jsonlReader, err := c.postJSONLReader("/reports", CreateNewReportBody{
		Async:  false,
		Jsonl:  true,
		Config: config,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not postBody: %w", err)
	}

	// We expect the first line to be the metadata
	metadata := &ReportMetadata{}
	err = jsonlReader.ReadSingleLine(metadata)
	if err != nil {
		return nil, nil, fmt.Errorf("could not ReadSingleLine of metadata: %w", err)
	}

	reportsChan := make(chan PerformanceReport)
	go func() {
		defer close(reportsChan)
		err := jsonlReader.ReadLines(func(data []byte) error {
			report := PerformanceReport{}

			err := json.Unmarshal(data, &report)
			if err != nil {
				return fmt.Errorf("could not unmarshal report: %w. Data: %q", err, string(data))
			}

			reportsChan <- report
			return nil
		})
		if err != nil {
			log.Error("could not ReadLines", err)
		}
	}()

	return metadata, reportsChan, nil
}

// CreateAsyncReport starts a new report via the VitalFrog api. Not returning any performance reports
func (c Client) CreateAsyncReport(config ReportConfig) (*AsyncReport, error) {
	report := &AsyncReport{}
	err := c.postJSON("/reports", CreateNewReportBody{
		Async:  true,
		Jsonl:  false,
		Config: config,
	}, report)
	if err != nil {
		return nil, fmt.Errorf("could not postJSON: %w", err)
	}
	return report, nil
}

func (c Client) postJSON(path string, in interface{}, out interface{}) error {
	jsonBody, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("could not marshal in type: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseUrl+path, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("could not create new http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := c.hClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not POST to %q: %w", c.baseUrl+path, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("invalid statuscode. got %d, but expected a 2xx. Body %q", resp.StatusCode, string(body))
	}

	err = json.Unmarshal(body, out)
	if err != nil {
		return fmt.Errorf("could not marshal into out type: %w", err)
	}
	return nil
}

func (c Client) getJSON(path string, out interface{}) error {
	req, err := http.NewRequest("GET", c.baseUrl+path, nil)
	if err != nil {
		return fmt.Errorf("could not create new http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := c.hClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not POST to %q: %w", c.baseUrl+path, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("invalid statuscode. got %d, but expected a 2xx. Body %q", resp.StatusCode, string(body))
	}

	err = json.Unmarshal(body, out)
	if err != nil {
		return fmt.Errorf("could not marshal into out type: %w", err)
	}
	return nil
}

func (c Client) postJSONLReader(path string, in interface{}) (*jsonl.Reader, error) {
	jsonBody, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("could not marshal in type: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseUrl+path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("could not create new http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := c.hClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not POST to %q: %w", c.baseUrl+path, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("invalid statuscode. got %d, but expected a 2xx.", resp.StatusCode)
	}

	jsonlReader := jsonl.NewReader(resp.Body)
	return &jsonlReader, nil
}
