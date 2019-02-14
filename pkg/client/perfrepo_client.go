package client

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/PerfCake/go-perfrepoclient/pkg/apis"
)

const (
	authHeader        = "Authorization"
	contentTypeHeader = "Content-Type"
	targetFileHeader  = "filename"
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
		return 0, errors.Wrap(errMsg(req, resp), "Error while creating Test")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	res, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func errMsg(req *http.Request, resp *http.Response) error {
	body, _ := ioutil.ReadAll(resp.Body)
	return fmt.Errorf("URL: %s, Status: %v, Response: %v", req.URL.String(), resp.Status, string(body))
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
		return nil, errors.Wrap(errMsg(req, resp), "Error while getting Test")
	}
}

// DeleteTest deletes the given test from the PerfRepo database. Returns nil when the request
// succeeds.
func (c *PerfRepoClient) DeleteTest(id int64) error {
	deleteTestURL := fmt.Sprintf("%s/test/id/%d", c.url, id)
	if err := c.deleteByURL(deleteTestURL); err != nil {
		errors.Wrap(err, fmt.Sprintf("Failed to delete test with id %d", id))
	}
	return nil
}

func (c *PerfRepoClient) deleteByURL(url string) error {
	req, err := c.httpDelete(url)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return errMsg(req, resp)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errMsg(req, resp)
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
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, errors.Wrap(errMsg(req, resp), "Error while creating TestExecution")
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
		return nil, errors.Wrap(errMsg(req, resp), "Error while getting TestExecution")
	}
}

// DeleteTestExecution deletes the given test execution from the PerfRepo database.
// Returns nil when the request succeeds.
func (c *PerfRepoClient) DeleteTestExecution(id int64) error {
	deleteTestExecURL := fmt.Sprintf("%s/testExecution/%d", c.url, id)
	if err := c.deleteByURL(deleteTestExecURL); err != nil {
		errors.Wrap(err, fmt.Sprintf("Failed to delete test execution with id %d", id))
	}
	return nil
}

// CreateAttachment creates a new attachment for a TestExecution identified by its ID.
// Returns an ID of the attachment itself or error when the operation failed
func (c *PerfRepoClient) CreateAttachment(testExecutionID int64, attachment apis.Attachment) (int64, error) {
	createAttachmentURL := fmt.Sprintf("%s/testExecution/%d/addAttachment", c.url, testExecutionID)

	req, err := http.NewRequest(http.MethodPost, createAttachmentURL, attachment.File)
	if err != nil {
		return 0, err
	}
	req.Header.Add(authHeader, "Basic "+c.auth)
	req.Header.Add(contentTypeHeader, attachment.ContentType)
	req.Header.Add(targetFileHeader, attachment.TargetFileName)

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, errors.Wrap(errMsg(req, resp), "Error while creating Attachment")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	res, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return 0, err
	}
	return res, nil
}

// GetAttachment returns the attachment with given ID in the form of io.Reader or
// error when the operation failed.
func (c *PerfRepoClient) GetAttachment(id int64) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/testExecution/attachment/%d", c.url, id)
	req, err := c.httpGet(url)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		if resp.ContentLength == 0 {
			return nil, fmt.Errorf("Attachment with given location %s doesn't exist", url)
		}
		return resp.Body, err
	default:
		return nil, errors.Wrap(errMsg(req, resp), "Error while getting Attachment")
	}
}

// CreateReport creates a new Report object in PerfRepo. Returns
// the ID of the Report record in database or returns 0 when there was an error.
func (c *PerfRepoClient) CreateReport(report *apis.Report) (int64, error) {
	createReportURL := c.url + "/report/create"

	marshalled, err := xml.MarshalIndent(report, "", "    ")
	if err != nil {
		return 0, err
	}

	req, err := c.httpPost(createReportURL, marshalled)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, errors.Wrap(errMsg(req, resp), "Error while creating Report")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	res, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return 0, err
	}

	return res, nil
}

// UpdateReport updates existing Report in PerfRepo. Returns
// the ID of the Report record in database or returns 0 when there was an error.
func (c *PerfRepoClient) UpdateReport(report *apis.Report) (int64, error) {
	if report == nil {
		return 0, errors.New("Invalid Report")
	}
	createReportURL := fmt.Sprintf("%s/report/update/%d", c.url, report.ID)

	marshalled, err := xml.MarshalIndent(report, "", "    ")
	if err != nil {
		return 0, err
	}

	req, err := c.httpPost(createReportURL, marshalled)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, errors.Wrap(errMsg(req, resp), "Error while updating Report")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	res, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return 0, err
	}
	return res, nil
}

// DeleteReport deletes the given Report from the PerfRepo database.
// Returns nil when the request succeeds.
func (c *PerfRepoClient) DeleteReport(id int64) error {
	deleteReportURL := fmt.Sprintf("%s/report/id/%d", c.url, id)
	if err := c.deleteByURL(deleteReportURL); err != nil {
		errors.Wrap(err, fmt.Sprintf("Failed to delete report with id %d", id))
	}
	return nil
}

// GetReport returns an existing Report by its identifier or nil if there's an error
func (c *PerfRepoClient) GetReport(id int64) (*apis.Report, error) {
	url := fmt.Sprintf("%s/report/id/%d", c.url, id)
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
			return nil, fmt.Errorf("Report with given location %s doesn't exist", url)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var t apis.Report
		err = xml.Unmarshal(body, &t)
		return &t, err
	default:
		return nil, errors.Wrap(errMsg(req, resp), "Error while getting Report")
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

func (c *PerfRepoClient) httpDelete(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(authHeader, "Basic "+c.auth)
	return req, nil
}
