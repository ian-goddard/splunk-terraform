package appinspect

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	AppInspectURL = "https://appinspect.splunk.com/v1/app/validate"
)

type Client struct {
	SplunkLoginToken string
}

type ClientInterface interface {
	CheckValidationStatus(validationID string) (*http.Response, error)
}

type WaitPrivateAppValidationRead struct {
	RequestID string  `json:"request_id"`
	Status    string  `json:"status"`
	Links     []Links `json:"links"`
	Info      Info    `json:"info"`
}

type Info struct {
	Error         int `json:"error"`
	Failure       int `json:"failure"`
	Skipped       int `json:"skipped"`
	NotApplicable int `json:"not_applicable"`
	Warning       int `json:"warning"`
	ManualCheck   int `json:"manual_check"`
	Success       int `json:"success"`
}

type Links struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

func GetAppInspectClient(splunkLoginToken string) ClientInterface {
	return &Client{
		SplunkLoginToken: strings.TrimSpace(splunkLoginToken),
	}
}

func (c *Client) NewCheckValidationStatusRequest(validationID string) (*http.Request, error) {
	validationID = strings.TrimSpace(validationID)
	if validationID == "" {
		return nil, fmt.Errorf("validation ID cannot be empty")
	}

	serverURL, err := url.Parse(fmt.Sprintf("%s/status/%s", AppInspectURL, validationID))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", serverURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Authorization", "bearer "+c.SplunkLoginToken)

	return req, nil
}

func (c *Client) CheckValidationStatus(validationID string) (*http.Response, error) {
	if c.SplunkLoginToken == "" {
		return nil, fmt.Errorf("authorization token cannot be empty")
	}

	req, err := c.NewCheckValidationStatusRequest(validationID)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
