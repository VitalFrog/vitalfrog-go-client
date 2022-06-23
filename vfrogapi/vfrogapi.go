package vfrogapi

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// CreateReport starts a new report via the VitalFrog api. Not returning any performance reports
func (c Client) CreateReport(config ReportConfig) (*ReportMetadata, error) {
	metadata := &ReportMetadata{}
	err := c.postJSON("/reports", config, metadata)
	if err != nil {
		return nil, fmt.Errorf("could not postJSON: %w", err)
	}
	return metadata, nil
}

// GetReport gets the report data by uuid
func (c Client) GetReport(uuid string) (*Report, error) {
	report := &Report{}
	err := c.getJSON(fmt.Sprintf("/reports/%s", uuid), report)
	if err != nil {
		return nil, fmt.Errorf("could not getJSON: %w", err)
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
