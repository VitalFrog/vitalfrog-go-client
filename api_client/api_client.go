package api_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	hClient  *http.Client
	baseUrl  string
	apiToken string
}

func New(baseUrl, apiToken string) Client {
	return Client{
		hClient:  &http.Client{},
		baseUrl:  baseUrl,
		apiToken: apiToken,
	}
}

func (c Client) CreateSyncReport(config ReportConfig) (*SyncReport, error) {
	report := &SyncReport{}
	err := c.postJSON("/reports", CreateNewReportBody{
		Async:  false,
		Config: config,
	}, report)
	if err != nil {
		return nil, fmt.Errorf("could not postJSON: %w", err)
	}
	return report, nil
}

func (c Client) CreateAsyncReport(config ReportConfig) (*AsyncReport, error) {
	report := &AsyncReport{}
	err := c.postJSON("/reports", CreateNewReportBody{
		Async:  true,
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
		return fmt.Errorf("invalid statuscode. got %q, but expected a 2xx. Body %q", resp.StatusCode, string(body))
	}

	err = json.Unmarshal(body, out)
	if err != nil {
		return fmt.Errorf("could not marshal into out type: %w", err)
	}
	return nil
}
