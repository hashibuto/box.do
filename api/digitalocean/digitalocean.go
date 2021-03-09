package digitalocean

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Service represents the underlying data structure for all DigitalOcean services.  This should be instantiated once and
// reused for all API calls
type Service struct {
	APIKey     string
	HTTPClient *http.Client
}

// ResponseBody represents the bare minimum response body to be unmarshaled from JSON in failure cases, in order to faciliate error reporting
type ResponseBody struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

const baseURL = "https://api.digitalocean.com/v2"

// NewService returns a new DigitalOcean API service structure
func NewService(apiKey string) *Service {
	return &Service{
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

// getFullURL returns a complete URL consisting of the urlSuffix, appended to the DigitalOcean base URL
func getFullURL(urlSuffix string) string {
	return fmt.Sprintf("%v%v", baseURL, urlSuffix)
}

// doRequest executes an authenticated HTTP request.  If there is no failure or error response returned, a
// byte array consisting of the response body is returned.
func (svc *Service) doRequest(req *http.Request) ([]byte, error) {
	if req.Method == "PUT" || req.Method == "POST" || req.Method == "PATCH" {
		req.Header.Add("Content-Type", "application/json")
	}
	req.Header.Add("Authorization", fmt.Sprint("Bearer ", svc.APIKey))

	resp, err := svc.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body := &ResponseBody{}
		err = json.Unmarshal(data, body)
		if err != nil {
			return nil, err
		}

		var message string
		if len(body.Message) > 0 {
			message = body.Message
		} else {
			message = fmt.Sprint("An API error occurred at ", req.URL)
		}
		return nil, errors.New(message)
	}

	return data, nil
}

// Get executes an authenticated GET request against the provided URL suffix and returns an http.Response pointer if successful
func (svc *Service) Get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", getFullURL(url), nil)
	if err != nil {
		return nil, err
	}
	return svc.doRequest(req)
}

// Post executes an authenticated POST request against the provided URL suffix, passing a byte array for the body (JSON marshaled), and
// returns an http.Response pointer if successful
func (svc *Service) Post(url string, body []byte) ([]byte, error) {
	reader := bytes.NewReader(body)

	req, err := http.NewRequest("POST", getFullURL(url), reader)
	if err != nil {
		return nil, err
	}
	return svc.doRequest(req)
}
