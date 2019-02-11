package client

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

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

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

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
	url := fmt.Sprintf("%s/test/id/%d", c.url, id)
	return c.getTestByURL(url)
}

// GetTestByUID returns an existing test by UID identifier or nil if there's an error
func (c *PerfRepoClient) GetTestByUID(uid string) (*apis.Test, error) {
	url := fmt.Sprintf("%s/test/uid/%s", c.url, uid)
	return c.getTestByURL(url)
}

func (c *PerfRepoClient) getTestByURL(url string) (*apis.Test, error) {
	req, err := c.httpGet(url)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if resp.ContentLength == 0 {
			return nil, fmt.Errorf("Test with given location %s doesn't exist", url)
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

// DeleteTest deletes the given test from the PerfRepo database. Returns nil when the request
// succeeds.
func (c *PerfRepoClient) DeleteTest(id int64) error {
	deleteTestURL := fmt.Sprintf("%s/test/id/%d", c.url, id)

	req, err := c.httpDelete(deleteTestURL)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Failed to delete test with id %d: %v", id, resp.Status)
	}
	return nil
}

// CreateTestExecution creates a new TestExecution object in PerfRepo with subobjects. Returns
// the ID of the TestExecution record in database or returns 0 when there was an error.
// TODO: remove redundant code that is common to API calls
func (c *PerfRepoClient) CreateTestExecution(testExec *apis.TestExecution) (int64, error) {
	createTestExecURL := c.url + "/testExecution/create"

	marshalled, err := xml.MarshalIndent(testExec, "", "    ")
	if err != nil {
		return 0, err
	}

	req, err := c.httpPost(createTestExecURL, marshalled)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "HTTP request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("Unexpected status code %v", errors.New(resp.Status))
	}
	body, _ := ioutil.ReadAll(resp.Body)
	res, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return 0, err
	}
	return res, nil
}

// GetTestExecution returns an existing test execution by its identifier or nil if there's an error
func (c *PerfRepoClient) GetTestExecution(id int64) (*apis.TestExecution, error) {
	url := fmt.Sprintf("%s/testExecution/%d", c.url, id)
	req, err := c.httpGet(url)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if resp.ContentLength == 0 {
			return nil, fmt.Errorf("Test execution with given location %s doesn't exist", url)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var t apis.TestExecution
		err = xml.Unmarshal(body, &t)
		return &t, err
	default:
		return nil, errors.New(resp.Status)
	}
}

// DeleteTestExecution deletes the given test execution from the PerfRepo database.
// Returns nil when the request succeeds.
func (c *PerfRepoClient) DeleteTestExecution(id int64) error {
	deleteTestURL := fmt.Sprintf("%s/testExecution/%d", c.url, id)

	req, err := c.httpDelete(deleteTestURL)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Failed to delete test execution with id %d: %v", id, resp.Status)
	}
	return nil
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

func (c *PerfRepoClient) httpDelete(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(authHeader, "Basic "+c.auth)
	return req, nil
}
