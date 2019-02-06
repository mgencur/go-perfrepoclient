package client

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/PerfCake/go-perfrepoclient/pkg/apis"
)

const (
	authHeader        = "Authorization"
	contentTypeHeader = "Content-Type"
)

// PerfRepoClient has methods for communicating with a remote PerfRepo instance via
// REST interface
type PerfRepoClient struct {
	client *http.Client
	url    string
	auth   string
}

// NewPerfRepoClient creates a new PerfRepoClient
func NewPerfRepoClient(url, username, password string) *PerfRepoClient {
	client := &PerfRepoClient{
		client: &http.Client{},
		url:    url,
		auth:   base64.StdEncoding.EncodeToString([]byte(username + ":" + password)),
	}
	return client
}

// CreateTest creates a new Test object in PerfRepo with subobjects. Returns
// the ID of the Test record in database or returns 0 when there was an error.
func (c *PerfRepoClient) CreateTest(test *apis.Test) (int64, error) {
	createTestURL := c.url + "/test/create"

	marshalled, err := xml.MarshalIndent(test, "", "    ")
	if err != nil {
		return 0, err
	}

	req, err := c.httpPost(createTestURL, marshalled)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Request: %+v\n", req)

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	fmt.Printf("Response: %+v\n", resp)

	if resp.StatusCode != http.StatusCreated {
		return 0, errors.New(resp.Status)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	res, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return 0, err
	}
	return res, nil
}

// GetTest returns an existing test by its identifier or nil if there's an error
func (c *PerfRepoClient) GetTest(id int64) (*apis.Test, error) {
	getTestURL := fmt.Sprintf("%s/test/id/%d", c.url, id)

	req, err := c.httpGet(getTestURL)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Request: %+v\n", req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("Response: %+v\n", resp)

	switch resp.StatusCode {
	case http.StatusOK:
		if resp.ContentLength == 0 {
			return nil, fmt.Errorf("Test with given id %d doesn't exist", id)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var t apis.Test
		err = xml.Unmarshal(body, &t)
		return &t, err
	default:
		return nil, errors.New(resp.Status)
	}
}

func (c *PerfRepoClient) httpGet(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(authHeader, "Basic "+c.auth)
	return req, nil
}

func (c *PerfRepoClient) httpPost(url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add(authHeader, "Basic "+c.auth)
	req.Header.Add(contentTypeHeader, "text/xml")
	return req, nil
}
