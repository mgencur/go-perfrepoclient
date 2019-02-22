package client

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/PerfCake/go-perfrepoclient/pkg/apis"
)

const (
	authHeader               = "Authorization"
	contentTypeHeader        = "Content-Type"
	contentDispositionHeader = "Content-Disposition"
	targetFileHeader         = "filename"
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
		url:    url + "/rest",
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

	return responseBodyAsInt(resp)
}

func errMsg(req *http.Request, resp *http.Response) error {
	body, _ := ioutil.ReadAll(resp.Body)
	return fmt.Errorf("URL: %s, Status: %v, Response: %v", req.URL.String(), resp.Status, string(body))
}

// AddMetric adds a new Metric to an existing Test. Returns
// the ID of the Metric or returns 0 when there was an error.
func (c *PerfRepoClient) AddMetric(testID int64, metric *apis.Metric) (int64, error) {
	url := fmt.Sprintf("%s/test/id/%d/addMetric", c.url, testID)

	marshalled, err := xml.MarshalIndent(metric, "", "    ")
	if err != nil {
		return 0, err
	}

	req, err := c.httpPost(url, marshalled)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, errors.Wrap(errMsg(req, resp), "Error while adding Metric")
	}

	return responseBodyAsInt(resp)
}

// GetMetric returns an existing Metric by its identifier or nil if there's an error
func (c *PerfRepoClient) GetMetric(id int64) (*apis.Metric, error) {
	url := fmt.Sprintf("%s/metric/%d", c.url, id)
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
			return nil, fmt.Errorf("Metric with given location %s doesn't exist", url)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var m apis.Metric
		err = xml.Unmarshal(body, &m)
		return &m, err
	default:
		return nil, errors.Wrap(errMsg(req, resp), "Error while getting Metric")
	}
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
// the ID of the TestExecution record in database or 0 in the event of error
func (c *PerfRepoClient) CreateTestExecution(testExec *apis.TestExecution) (id int64, err error) {
	createTestExecURL := c.url + "/testExecution/create"
	if id, err = c.postTestExecution(testExec, createTestExecURL); err != nil {
		err = errors.Wrap(err, "Failed to create test execution")
	}
	return
}

// UpdateTestExecution updates a given TestExecution object in PerfRepo. Returns
// the ID of the TestExecution record in database or 0 in the event of error
func (c *PerfRepoClient) UpdateTestExecution(testExec *apis.TestExecution) (id int64, err error) {
	if testExec == nil || testExec.ID == 0 {
		id, err = 0, errors.New("Invalid test execution for update")
	}
	updateTestExecURL := fmt.Sprintf("%s/testExecution/update/%d", c.url, testExec.ID)
	if id, err = c.postTestExecution(testExec, updateTestExecURL); err != nil {
		err = errors.Wrap(err, "Failed to update test execution")
	}
	return
}

func (c *PerfRepoClient) postTestExecution(testExec *apis.TestExecution, url string) (int64, error) {
	marshalled, err := xml.MarshalIndent(testExec, "", "    ")
	if err != nil {
		return 0, err
	}

	req, err := c.httpPost(url, marshalled)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, errMsg(req, resp)
	}

	return responseBodyAsInt(resp)
}

func responseBodyAsInt(resp *http.Response) (int64, error) {
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

// SearchTestExecutions searches for test executions based on criteria passed as the argument.
func (c *PerfRepoClient) SearchTestExecutions(criteria *apis.TestExecutionSearch) ([]apis.TestExecution, error) {
	searchTestExecURL := c.url + "/testExecution/search"

	marshalled, err := xml.MarshalIndent(criteria, "", "    ")
	if err != nil {
		return nil, err
	}

	req, err := c.httpPost(searchTestExecURL, marshalled)
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
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var t apis.TestExecutions
		err = xml.Unmarshal(body, &t)
		return t.TestExecutions, err
	default:
		return nil, errors.Wrap(errMsg(req, resp), "Error while searching TestExecutions")
	}
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

	return responseBodyAsInt(resp)
}

// GetAttachment returns an existing attachment with given ID or
// error when the operation failed.
func (c *PerfRepoClient) GetAttachment(id int64) (*apis.Attachment, error) {
	url := fmt.Sprintf("%s/testExecution/attachment/%d", c.url, id)
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
			return nil, fmt.Errorf("Attachment with given location %s doesn't exist", url)
		}
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return &apis.Attachment{
			File:           bytes.NewReader(bodyBytes),
			ContentType:    resp.Header.Get(contentTypeHeader),
			TargetFileName: parseFileName(resp.Header.Get(contentDispositionHeader)),
		}, nil
	default:
		return nil, errors.Wrap(errMsg(req, resp), "Error while getting Attachment")
	}
}

// The header value is in this format: attachment; filename=attachment1.txt
func parseFileName(headerValue string) string {
	const fileSep = "filename="
	parts := strings.Split(headerValue, ";")
	if len(parts) != 2 || !strings.Contains(parts[1], fileSep) {
		return ""
	}
	return strings.Split(parts[1], fileSep)[1]
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

	return responseBodyAsInt(resp)
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

	return responseBodyAsInt(resp)
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

// CreateReportPermission adds a new permission to an existing report. Returns
// nil if the operation was successful.
func (c *PerfRepoClient) CreateReportPermission(permission *apis.Permission) error {
	url := fmt.Sprintf("%s/report/id/%d/addPermission", c.url, permission.ReportID)

	marshalled, err := xml.MarshalIndent(permission, "", "    ")
	if err != nil {
		return err
	}

	req, err := c.httpPost(url, marshalled)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//This is inconsistent with other "Create" API methods where PerfRepo returns StatusCreated
	if resp.StatusCode != http.StatusOK {
		return errors.Wrap(errMsg(req, resp), "Error while adding Permission to Report")
	}
	//The return type is inconsistent with other "Create" API methods where PerfRepo returns id
	//of the object
	return nil
}

// DeleteReportPermission deletes the given permission from the PerfRepo database.
// Returns nil when the request succeeds
func (c *PerfRepoClient) DeleteReportPermission(permission *apis.Permission) error {
	deletePermissionURL := fmt.Sprintf("%s/report/id/%d/deletePermission", c.url, permission.ReportID)

	marshalled, err := xml.MarshalIndent(permission, "", "    ")
	if err != nil {
		return err
	}

	req, err := c.httpPost(deletePermissionURL, marshalled)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return errMsg(req, resp)
	}
	defer resp.Body.Close()

	//This is inconsistent with other "Delete" API methods where PerfRepo returns StatusNoContent
	if resp.StatusCode != http.StatusOK {
		return errors.Wrap(errMsg(req, resp), "Error while deleting permission")
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
