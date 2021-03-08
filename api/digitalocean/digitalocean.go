package digitalocean

import (
	"bytes"
	"fmt"
	"net/http"
)

// Service represents the underlying data structure for all DigitalOcean services.  This should be instantiated once and
// reused for all API calls
type Service struct {
	APIKey     string
	HTTPClient *http.Client
}

const baseURL = "https://api.digitalocean.com/v2"

// NewService returns a new DigitalOcean API service structure
func NewService(apiKey string) *Service {
	return &Service{
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

func (svc *Service) addHeaders(req *http.Request) {
	if req.Method == "PUT" || req.Method == "POST" || req.Method == "PATCH" {
		req.Header.Add("Content-Type", "application/json")
	}
	req.Header.Add("Authorization", fmt.Sprint("Bearer", svc.APIKey))
}

func getFullURL(urlSuffix string) string {
	return fmt.Sprintf("%v%v", baseURL, urlSuffix)
}

// Get executes an authenticated GET request against the provided URL suffix and returns an http.Response pointer if successful
func (svc *Service) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", getFullURL(url), nil)
	if err != nil {
		return nil, err
	}
	svc.addHeaders(req)
	resp, err := svc.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Post executes an authenticated POST request against the provided URL suffix, passing a byte array for the body (JSON unmarshaled), and
// returns an http.Response pointer if successful
func (svc *Service) Post(url string, body []byte) (*http.Response, error) {
	reader := bytes.NewReader(body)

	req, err := http.NewRequest("POST", getFullURL(url), reader)
	if err != nil {
		return nil, err
	}
	svc.addHeaders(req)
	resp, err := svc.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
