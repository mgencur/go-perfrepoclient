package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/mgencur/go-perfrepoclient/pkg/apis"
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
	Client *http.Client
	URL    string
	Auth   string
}

// NewClient creates a new PerfRepoClient
func NewClient(url, username, password string) *PerfRepoClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return newClientWithTransport(transport, url, username, password)
}

// NewSecuredClient creates a new PerfRepoClient vith CA (certification authority) setup for TLS.
func NewSecuredClient(url, username, password, caFile string) (*PerfRepoClient, error) {
	// Load CA cert
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read CA file")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	return newClientWithTransport(transport, url, username, password), nil
}

func newClientWithTransport(transport *http.Transport, url, username, password string) *PerfRepoClient {
	client := &PerfRepoClient{
		Client: &http.Client{
			Transport: transport,
		},
		URL:  url + "/rest",
		Auth: "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password)),
	}
	return client
}

// CreateTest creates a new Test object in PerfRepo with subobjects. Returns
// the ID of the Test record in database or returns 0 when there was an error.
func (c *PerfRepoClient) CreateTest(test *apis.Test) (id int64, err error) {
	createTestURL := c.URL + "/test/create"
	if id, err = c.postEntity(test, createTestURL); err != nil {
		return 0, errors.Wrap(err, "Failed to create test")
	}
	return id, nil
}

func errMsg(req *http.Request, resp *http.Response) error {
	body, _ := ioutil.ReadAll(resp.Body)
	return fmt.Errorf("URL: %s, Status: %v, Response: %v", req.URL.String(), resp.Status, string(body))
}

// AddMetric adds a new Metric to an existing Test. Returns
// the ID of the Metric or returns 0 when there was an error.
func (c *PerfRepoClient) AddMetric(testID int64, metric *apis.Metric) (id int64, err error) {
	addMetricURL := fmt.Sprintf("%s/test/id/%d/addMetric", c.URL, testID)
	if id, err = c.postEntity(metric, addMetricURL); err != nil {
		return 0, errors.Wrap(err, "Failed to add metric")
	}
	return id, nil
}

// GetMetric returns an existing Metric by its identifier or nil if there's an error
func (c *PerfRepoClient) GetMetric(id int64) (*apis.Metric, error) {
	URL := fmt.Sprintf("%s/metric/%d", c.URL, id)
	entity, err := c.getEntity(URL)
	if err != nil {
		return nil, err
	}
	var m apis.Metric
	err = xml.Unmarshal(entity, &m)
	return &m, errors.Wrap(err, "Failed to get metric")
}

// GetTest returns an existing test by its identifier or nil if there's an error
func (c *PerfRepoClient) GetTest(id int64) (*apis.Test, error) {
	URL := fmt.Sprintf("%s/test/id/%d", c.URL, id)
	test, err := c.getTest(URL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get test by id")
	}
	return test, nil
}

// GetTestByUID returns an existing test by UID identifier or nil if there's an error
func (c *PerfRepoClient) GetTestByUID(uid string) (*apis.Test, error) {
	URL := fmt.Sprintf("%s/test/uid/%s", c.URL, uid)
	test, err := c.getTest(URL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get test by uid")
	}
	return test, nil
}

func (c *PerfRepoClient) getTest(URL string) (*apis.Test, error) {
	entity, err := c.getEntity(URL)
	if err != nil {
		return nil, err
	}
	var test apis.Test
	err = xml.Unmarshal(entity, &test)
	return &test, err
}

func (c *PerfRepoClient) getEntity(URL string) ([]byte, error) {
	req, err := c.httpGet(URL)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if resp.ContentLength == 0 {
			return nil, fmt.Errorf("Entity with given location %s doesn't exist", URL)
		}
		return ioutil.ReadAll(resp.Body)
	default:
		return nil, errors.Wrap(errMsg(req, resp), "Error while getting entity")
	}
}

// DeleteTest deletes the given test from the PerfRepo database. Returns nil when the request
// succeeds.
func (c *PerfRepoClient) DeleteTest(id int64) error {
	deleteTestURL := fmt.Sprintf("%s/test/id/%d", c.URL, id)
	if err := c.delete(deleteTestURL); err != nil {
		errors.Wrap(err, fmt.Sprintf("Failed to delete test with id %d", id))
	}
	return nil
}

func (c *PerfRepoClient) delete(URL string) error {
	req, err := c.httpDelete(URL)
	if err != nil {
		return err
	}

	resp, err := c.Client.Do(req)
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
	createTestExecURL := c.URL + "/testExecution/create"
	if id, err = c.postEntity(testExec, createTestExecURL); err != nil {
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
	updateTestExecURL := fmt.Sprintf("%s/testExecution/update/%d", c.URL, testExec.ID)
	if id, err = c.postEntity(testExec, updateTestExecURL); err != nil {
		err = errors.Wrap(err, "Failed to update test execution")
	}
	return
}

// postEntity sends a HTTP post with the given entity masrhalled as a body of the request.
// Returns the id of the entity record in database or 0 in the event of error
func (c *PerfRepoClient) postEntity(entity interface{}, URL string) (int64, error) {
	marshalled, err := xml.MarshalIndent(entity, "", "    ")
	if err != nil {
		return 0, err
	}

	req, err := c.httpPost(URL, marshalled)
	if err != nil {
		return 0, err
	}

	resp, err := c.Client.Do(req)
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
	URL := fmt.Sprintf("%s/testExecution/%d", c.URL, id)
	entity, err := c.getEntity(URL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get metric")
	}
	var t apis.TestExecution
	err = xml.Unmarshal(entity, &t)
	return &t, err
}

// DeleteTestExecution deletes the given test execution from the PerfRepo database.
// Returns nil when the request succeeds.
func (c *PerfRepoClient) DeleteTestExecution(id int64) error {
	deleteTestExecURL := fmt.Sprintf("%s/testExecution/%d", c.URL, id)
	if err := c.delete(deleteTestExecURL); err != nil {
		errors.Wrap(err, fmt.Sprintf("Failed to delete test execution with id %d", id))
	}
	return nil
}

// SearchTestExecutions searches for test executions based on criteria passed as the argument.
func (c *PerfRepoClient) SearchTestExecutions(criteria *apis.TestExecutionSearch) ([]apis.TestExecution, error) {
	searchTestExecURL := c.URL + "/testExecution/search"

	marshalled, err := xml.MarshalIndent(criteria, "", "    ")
	if err != nil {
		return nil, err
	}

	req, err := c.httpPost(searchTestExecURL, marshalled)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
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
	createAttachmentURL := fmt.Sprintf("%s/testExecution/%d/addAttachment", c.URL, testExecutionID)

	req, err := http.NewRequest(http.MethodPost, createAttachmentURL, attachment.File)
	if err != nil {
		return 0, err
	}
	req.Header.Add(authHeader, c.Auth)
	req.Header.Add(contentTypeHeader, attachment.ContentType)
	req.Header.Add(targetFileHeader, attachment.TargetFileName)

	resp, err := c.Client.Do(req)
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
	URL := fmt.Sprintf("%s/testExecution/attachment/%d", c.URL, id)
	req, err := c.httpGet(URL)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if resp.ContentLength == 0 {
			return nil, fmt.Errorf("Attachment with given location %s doesn't exist", URL)
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
func (c *PerfRepoClient) CreateReport(report *apis.Report) (id int64, err error) {
	createReportURL := c.URL + "/report/create"
	if id, err = c.postEntity(report, createReportURL); err != nil {
		return 0, errors.Wrap(err, "Failed to create report")
	}
	return id, nil
}

// UpdateReport updates existing Report in PerfRepo. Returns
// the ID of the Report record in database or returns 0 when there was an error.
func (c *PerfRepoClient) UpdateReport(report *apis.Report) (id int64, err error) {
	if report == nil {
		return 0, errors.New("Invalid Report")
	}
	updateReportURL := fmt.Sprintf("%s/report/update/%d", c.URL, report.ID)
	if id, err = c.postEntity(report, updateReportURL); err != nil {
		return 0, errors.Wrap(err, "Failed to udpate report")
	}
	return id, nil
}

// DeleteReport deletes the given Report from the PerfRepo database.
// Returns nil when the request succeeds.
func (c *PerfRepoClient) DeleteReport(id int64) error {
	deleteReportURL := fmt.Sprintf("%s/report/id/%d", c.URL, id)
	if err := c.delete(deleteReportURL); err != nil {
		errors.Wrap(err, fmt.Sprintf("Failed to delete report with id %d", id))
	}
	return nil
}

// GetReport returns an existing Report by its identifier or nil if there's an error
func (c *PerfRepoClient) GetReport(id int64) (*apis.Report, error) {
	URL := fmt.Sprintf("%s/report/id/%d", c.URL, id)
	entity, err := c.getEntity(URL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get report")
	}
	var r apis.Report
	err = xml.Unmarshal(entity, &r)
	return &r, err
}

// CreateReportPermission adds a new permission to an existing report. Returns
// nil if the operation was successful.
func (c *PerfRepoClient) CreateReportPermission(permission *apis.Permission) error {
	URL := fmt.Sprintf("%s/report/id/%d/addPermission", c.URL, permission.ReportID)

	marshalled, err := xml.MarshalIndent(permission, "", "    ")
	if err != nil {
		return err
	}

	req, err := c.httpPost(URL, marshalled)
	if err != nil {
		return err
	}

	resp, err := c.Client.Do(req)
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
	deletePermissionURL := fmt.Sprintf("%s/report/id/%d/deletePermission", c.URL, permission.ReportID)

	marshalled, err := xml.MarshalIndent(permission, "", "    ")
	if err != nil {
		return err
	}

	req, err := c.httpPost(deletePermissionURL, marshalled)
	if err != nil {
		return err
	}

	resp, err := c.Client.Do(req)
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

// GetServerVersion returns the server version
func (c *PerfRepoClient) GetServerVersion() (string, error) {
	URL := c.URL + "/info/version"
	entity, err := c.getEntity(URL)
	if err != nil {
		return "", errors.Wrap(err, "Failed to get server version")
	}
	version := string(entity)
	return version, nil
}

func (c *PerfRepoClient) httpGet(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(authHeader, c.Auth)
	return req, nil
}

func (c *PerfRepoClient) httpPost(url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add(authHeader, c.Auth)
	req.Header.Add(contentTypeHeader, "text/xml")
	return req, nil
}

func (c *PerfRepoClient) httpDelete(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(authHeader, c.Auth)
	return req, nil
}
