// +build e2e

package e2e

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/PerfCake/go-perfrepoclient/pkg/apis"
	"github.com/PerfCake/go-perfrepoclient/pkg/client"
	"github.com/PerfCake/go-perfrepoclient/test"
)

var testClient *client.PerfRepoClient

func TestMain(m *testing.M) {
	testClient = client.NewPerfRepoClient(test.Flags.URL, test.Flags.User, test.Flags.Pass)
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
func TestCreateGetDeleteTest(t *testing.T) {
	testIn := test.Test("test1")

	id, err := testClient.CreateTest(testIn)

	if err != nil {
		t.Fatal("Failed to create Test", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTest(id); err != nil {
			t.Fatal(err.Error())
		}

		if _, err = testClient.GetTest(id); err == nil || !strings.Contains(err.Error(), "doesn't exist") {
			t.Fatalf("Test not deleted")
		}
	}()

	testOut, err := testClient.GetTest(id)
	if err != nil {
		t.Fatal("Failed to get Test", err.Error())
	}

	if testIn.Name != testOut.Name ||
		testIn.Description != testOut.Description ||
		testIn.GroupID != testOut.GroupID ||
		id != testOut.ID ||
		testIn.UID != testOut.UID {
		//TODO: Verify testOut.TestExecutions are nil
		t.Fatalf("The returned test: %+v does not match the original test %+v", testOut, testIn)
	}
}

func TestGetTestByUID(t *testing.T) {
	testIn := test.Test("test1")

	id, err := testClient.CreateTest(testIn)

	if err != nil {
		t.Fatal("Failed to create Test", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTest(id); err != nil {
			t.Fatal(err.Error())
		}
	}()

	testOut, err := testClient.GetTestByUID(testIn.UID)
	if err != nil {
		t.Fatal("Failed to get Test", err.Error())
	}

	if testIn.Name != testOut.Name ||
		testIn.Description != testOut.Description ||
		testIn.GroupID != testOut.GroupID ||
		id != testOut.ID ||
		testIn.UID != testOut.UID {
		t.Fatalf("The returned test: %+v does not match the original test %+v", testOut, testIn)
	}
}

func TestAddGetMetric(t *testing.T) {
	t.Skip("https://github.com/PerfCake/PerfRepo/issues/94")
	testIn := test.Test("test1")

	id, err := testClient.CreateTest(testIn)

	if err != nil {
		t.Fatal("Failed to create Test", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTest(id); err != nil {
			t.Fatal(err.Error())
		}

		if _, err = testClient.GetTest(id); err == nil || !strings.Contains(err.Error(), "doesn't exist") {
			t.Fatalf("Test not deleted")
		}
	}()

	testOut, err := testClient.GetTest(id)
	if err != nil {
		t.Fatal("Failed to get Test", err.Error())
	}

	if !metricsEqual(testOut, testIn, "metric1", "metric2") {
		t.Fatalf("The returned metrics: %+v do not match the original metrics%+v", testOut.Metrics, testIn.Metrics)
	}

	newMetric := &apis.Metric{
		Comparator:  "LB",
		Name:        "metric3",
		Description: "this is a test metric 3",
	}

	metricID, err := testClient.AddMetric(id, newMetric)
	if err != nil || metricID == 0 {
		t.Fatal("Failed to add metric", err.Error())
	}

	updatedTest, err := testClient.GetTest(id)
	if err != nil {
		t.Fatal("Failed to get Test", err.Error())
	}

	testIn.Metrics = append(testIn.Metrics, *newMetric)
	if !metricsEqual(testOut, testIn, "metric1", "metric2", "metric3") {
		t.Fatalf("The returned metrics: %+v do not match the original metrics%+v", updatedTest.Metrics, testIn.Metrics)
	}
}

func TestCreateGetDeleteTestExecution(t *testing.T) {
	testIn := test.Test("test1")

	testID, err := testClient.CreateTest(testIn)

	if err != nil {
		t.Fatal("Failed to create Test", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTest(testID); err != nil {
			t.Fatal(err.Error())
		}
	}()

	testExecIn := test.ExecutionDefault(testID)

	testExecID, err := testClient.CreateTestExecution(testExecIn)

	if err != nil {
		t.Fatal("Failed to create TestExecution", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTestExecution(testExecID); err != nil {
			t.Fatal(err.Error())
		}
		if _, err = testClient.GetTestExecution(testExecID); err == nil || !strings.Contains(err.Error(), "doesn't exist") {
			t.Fatalf("Test execution not deleted")
		}
	}()

	testExecOut, err := testClient.GetTestExecution(testExecID)

	if err != nil {
		t.Fatal("Failed to get TestExecution", err.Error())
	}

	if testExecOut.ID != testExecID ||
		testExecOut.Name != testExecIn.Name ||
		testExecOut.Started.String() != testExecIn.Started.String() ||
		!paramsEqual(testExecOut, testExecIn) ||
		!tagsEqual(testExecOut, testExecIn) ||
		!valuesEqual(testExecOut, testExecIn, "metric1", "metric2") ||
		firstMetricByParam(testExecOut, "multimetric",
			apis.ValueParameter{Name: "client", Value: "1"}) != 20.0 ||
		firstMetricByParam(testExecOut, "multimetric",
			apis.ValueParameter{Name: "client", Value: "2"}) != 40.0 {
		t.Fatalf("The returned test execution: %+v does not match the original %+v",
			testExecOut, testExecIn)
	}
}

func TestCreateInvalidTestExecution(t *testing.T) {
	testIn := test.Test("test1")

	testID, err := testClient.CreateTest(testIn)

	if err != nil {
		t.Fatal("Failed to create Test", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTest(testID); err != nil {
			t.Fatal(err.Error())
		}
	}()

	testExecIn := test.InvalidTestExecution(testID)

	testExecID, err := testClient.CreateTestExecution(testExecIn)

	if err == nil || testExecID != 0 {
		t.Fatal("Invalid test execution accepted")
	}
}

//TODO: Change this to actually update test execution
func TestUpdateTestExecution(t *testing.T) {
	testIn := test.Test("test1")

	testID, err := testClient.CreateTest(testIn)

	if err != nil {
		t.Fatal("Failed to create Test", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTest(testID); err != nil {
			t.Fatal(err.Error())
		}
	}()

	testExecIn := test.ExecutionDefault(testID)

	testExecID, err := testClient.CreateTestExecution(testExecIn)

	if err != nil {
		t.Fatal("Failed to create TestExecution", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTestExecution(testExecID); err != nil {
			t.Fatal(err.Error())
		}
		if _, err = testClient.GetTestExecution(testExecID); err == nil ||
			!strings.Contains(err.Error(), "doesn't exist") {
			t.Fatalf("Test execution not deleted")
		}
	}()

	testExecOut, err := testClient.GetTestExecution(testExecID)

	if err != nil {
		t.Fatal("Failed to get TestExecution", err.Error())
	}

	if testExecOut.ID != testExecID ||
		testExecOut.Name != testExecIn.Name ||
		testExecOut.Started.String() != testExecIn.Started.String() ||
		!paramsEqual(testExecOut, testExecIn) ||
		!tagsEqual(testExecOut, testExecIn) ||
		!valuesEqual(testExecOut, testExecIn, "metric1", "metric2") ||
		firstMetricByParam(testExecOut, "multimetric",
			apis.ValueParameter{Name: "client", Value: "1"}) != 20.0 ||
		firstMetricByParam(testExecOut, "multimetric",
			apis.ValueParameter{Name: "client", Value: "2"}) != 40.0 {
		t.Fatalf("The returned test execution: %+v does not match the original %+v",
			testExecOut, testExecIn)
	}
}

func TestSearchTestExecutions(t *testing.T) {
	test1In := test.Test("test1")
	test1ID, err := testClient.CreateTest(test1In)
	if err != nil {
		t.Fatal("Failed to create Test", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTest(test1ID); err != nil {
			t.Fatal(err.Error())
		}
	}()

	test2In := test.Test("test2")
	test2ID, err := testClient.CreateTest(test2In)
	if err != nil {
		t.Fatal("Failed to create Test", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTest(test2ID); err != nil {
			t.Fatal(err.Error())
		}
	}()

	// create 1. test execution
	params := []apis.TestExecutionParameter{
		{Name: "param1", Value: "value1"},
		{Name: "param2", Value: "value2"},
	}
	tags := []apis.Tag{{Name: "tag1"}, {Name: "tag2"}}
	datetime := time.Date(2016, time.July, 7, 0, 0, 0, 0, time.UTC)
	testExec1 := test.Execution(test1ID, &apis.JaxbTime{datetime}, params, tags)
	testExec1ID, err := testClient.CreateTestExecution(testExec1)
	if err != nil {
		t.Fatal("Failed to create TestExecution", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTestExecution(testExec1ID); err != nil {
			t.Fatal(err.Error())
		}
	}()

	// create 2. test execution
	params = []apis.TestExecutionParameter{
		{Name: "param1", Value: "value3"},
		{Name: "param2", Value: "value4"},
	}
	tags = []apis.Tag{{Name: "tag2"}, {Name: "tag3"}}
	datetime = time.Date(2016, time.July, 10, 0, 0, 0, 0, time.UTC)
	testExec2 := test.Execution(test1ID, &apis.JaxbTime{datetime}, params, tags)
	testExec2ID, err := testClient.CreateTestExecution(testExec2)
	if err != nil {
		t.Fatal("Failed to create TestExecution", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTestExecution(testExec2ID); err != nil {
			t.Fatal(err.Error())
		}
	}()

	// create 3. test execution
	params = []apis.TestExecutionParameter{
		{Name: "param1", Value: "value1"},
		{Name: "param2", Value: "value3"},
	}
	tags = []apis.Tag{{Name: "tag3"}, {Name: "tag4"}}
	datetime = time.Date(2016, time.July, 13, 0, 0, 0, 0, time.UTC)
	testExec3 := test.Execution(test2ID, &apis.JaxbTime{datetime}, params, tags)
	testExec3ID, err := testClient.CreateTestExecution(testExec3)
	if err != nil {
		t.Fatal("Failed to create TestExecution", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTestExecution(testExec3ID); err != nil {
			t.Fatal(err.Error())
		}
	}()

	// create 1. search
	ids := []int64{testExec1ID, testExec2ID}
	criteria := &apis.TestExecutionSearch{
		IDS: &ids,
	}
	executions, err := testClient.SearchTestExecutions(criteria)

	if len(ids) != len(executions) ||
		!idsIncluded(executions, ids...) {
		t.Fatalf("The returned test executions do not match the search criteria. Executions: %+v, Criteria: %+v",
			executions, criteria)
	}

	// create 2. search
	criteria = &apis.TestExecutionSearch{
		Tags: "tag2",
	}
	executions, err = testClient.SearchTestExecutions(criteria)

	if len(ids) != len(executions) ||
		!idsIncluded(executions, ids...) {
		t.Fatalf("The returned test executions do not match the search criteria. Executions: %+v, Criteria: %+v",
			executions, criteria)
	}

	// create 3. search
	ids = []int64{testExec1ID, testExec3ID}
	criteria = &apis.TestExecutionSearch{
		Parameters: []apis.CriteriaParameter{
			{Name: "param1", Value: "value1"},
		},
	}
	executions, err = testClient.SearchTestExecutions(criteria)

	if len(ids) != len(executions) ||
		!idsIncluded(executions, ids...) {
		t.Fatalf("The returned test executions do not match the search criteria. Executions: %+v, Criteria: %+v",
			executions, criteria)
	}
}

func TestCreateGetAttachment(t *testing.T) {
	testIn := test.Test("test1")

	testID, err := testClient.CreateTest(testIn)

	if err != nil {
		t.Fatal("Failed to create Test", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTest(testID); err != nil {
			t.Fatal(err.Error())
		}
	}()

	testExecIn := test.ExecutionDefault(testID)

	testExecID, err := testClient.CreateTestExecution(testExecIn)

	if err != nil {
		t.Fatal("Failed to create TestExecution", err.Error())
	}
	defer func() {
		if err := testClient.DeleteTestExecution(testExecID); err != nil {
			t.Fatal(err.Error())
		}
		if _, err = testClient.GetTestExecution(testExecID); err == nil || !strings.Contains(err.Error(), "doesn't exist") {
			t.Fatalf("Test execution not deleted")
		}
	}()

	const attachmentText = "this is a juicy test file"
	attIn := apis.Attachment{
		File:           strings.NewReader(attachmentText),
		ContentType:    "text/plain",
		TargetFileName: "attachment1.txt",
	}
	attID, err := testClient.CreateAttachment(testExecID, attIn)
	if err != nil {
		t.Fatal("Failed to create Attachment", err.Error())
	}
	attOut, err := testClient.GetAttachment(attID)
	if err != nil {
		t.Fatal("Failed to get Attachment", err.Error())
	}
	bodyBytes, err := ioutil.ReadAll(attOut.File)
	if err != nil {
		t.Fatal("Unable to read attachment", err.Error())
	}
	if string(bodyBytes) != attachmentText ||
		attOut.ContentType != attIn.ContentType ||
		attOut.TargetFileName != attIn.TargetFileName {
		t.Fatalf("The returned attachment: %+v does not match the original %+v",
			attOut, attIn)
	}
}

func TestCreateGetDeleteReport(t *testing.T) {
	reportIn := test.Report("report", test.Flags.User)

	id, err := testClient.CreateReport(reportIn)

	if err != nil {
		t.Fatal("Failed to create Report", err.Error())
	}
	defer func() {
		if err := testClient.DeleteReport(id); err != nil {
			t.Fatal(err.Error())
		}
		if _, err = testClient.GetReport(id); err == nil {
			t.Fatalf("Report not deleted: %v", err)
		}
	}()

	reportOut, err := testClient.GetReport(id)
	if err != nil {
		t.Fatal("Failed to get Report", err.Error())
	}

	if reportOut.Name != reportIn.Name ||
		reportOut.Type != reportIn.Type ||
		len(reportOut.Properties) != len(reportIn.Properties) ||
		!propertiesEqual(reportOut, reportIn, "property1") {
		t.Fatalf("The returned report: %+v does not match the original %+v", reportOut, reportIn)
	}
}

func TestUpdateReport(t *testing.T) {
	orig := test.Report("report", test.Flags.User)

	origID, err := testClient.CreateReport(orig)

	if err != nil {
		t.Fatal("Failed to create Report", err.Error())
	}
	defer func() {
		if err := testClient.DeleteReport(origID); err != nil {
			t.Fatal(err.Error())
		}
		if _, err = testClient.GetReport(origID); err == nil {
			t.Fatalf("Report not deleted: %v", err)
		}
	}()

	update := test.Report("updated report", test.Flags.User)
	update.ID = origID //use the same id, we're updating the original report
	update.Type = "ReportUpdate"
	update.Properties["property2"] = "value"

	updateID, err := testClient.UpdateReport(update)

	updateOut, err := testClient.GetReport(updateID)
	if err != nil {
		t.Fatal("Failed to get Report", err.Error())
	}

	if updateOut.Name != update.Name ||
		updateOut.Type != update.Type ||
		len(updateOut.Properties) != len(update.Properties) ||
		!propertiesEqual(updateOut, update, "property1", "property2") {
		t.Fatalf("The returned report: %+v does not match the original %+v", updateOut, update)
	}
}

func TestCreateDeleteReportPermission(t *testing.T) {
	report := test.Report("report", test.Flags.User)

	reportID, err := testClient.CreateReport(report)
	if err != nil {
		t.Fatal("Failed to create Report", err.Error())
	}

	defer func() {
		if err := testClient.DeleteReport(reportID); err != nil {
			t.Fatal(err.Error())
		}
	}()

	reportOut, err := testClient.GetReport(reportID)
	if err != nil {
		t.Fatal("Failed to get Report", err.Error())
	}

	if len(reportOut.Permissions) != 1 ||
		reportOut.Permissions[0].AccessLevel != "GROUP" ||
		reportOut.Permissions[0].AccessType != "WRITE" {
		t.Fatal("Default permissions not applied")
	}

	permission := &apis.Permission{
		//need to specify this name because PerfRepo expects different name when
		//the permission is a standalone message and when it's part of a report
		XMLName:     xml.Name{"", "report-permission"},
		ReportID:    reportOut.ID,
		AccessLevel: "PUBLIC",
		AccessType:  "READ",
	}

	err = testClient.CreateReportPermission(permission)
	if err != nil {
		t.Fatal("Failed to create permission", err.Error())
	}

	defer func() {
		if err := testClient.DeleteReportPermission(permission); err != nil {
			t.Fatal("Failed to delete permission", err.Error())
		}
		reportOut, err := testClient.GetReport(reportID)
		if err != nil {
			t.Fatal("Failed to get Report", err.Error())
		}
		if len(reportOut.Permissions) != 1 ||
			containsPermission(reportOut.Permissions, permission) {
			t.Fatal("Permission not deleted")
		}
	}()

	reportOut, err = testClient.GetReport(reportID)
	if err != nil {
		t.Fatal("Failed to get Report", err.Error())
	}

	if len(reportOut.Permissions) != 2 ||
		!containsPermission(reportOut.Permissions, permission) {
		t.Fatal("Default permissions not applied")
	}
}

func paramsEqual(actual, expected *apis.TestExecution) bool {
	actualSorted := actual.SortedParameters()
	for i, p := range expected.SortedParameters() {
		if actualSorted[i].Name != p.Name || actualSorted[i].Value != p.Value {
			return false
		}
	}
	return true
}

func tagsEqual(actual, expected *apis.TestExecution) bool {
	actualSorted := actual.SortedTags()
	for i, tag := range expected.SortedTags() {
		if actualSorted[i].Name != tag.Name {
			return false
		}
	}
	return true
}

func valuesEqual(actual, expected *apis.TestExecution, metricsToCompare ...string) bool {
	actualMetricNames := make([]string, 0)
	for _, v := range actual.Values {
		actualMetricNames = append(actualMetricNames, v.MetricName)
	}
	for _, m := range metricsToCompare {
		if !isIncluded(m, actualMetricNames...) {
			//the requested metric doesn't exist
			return false
		}
		for _, av := range actual.Values {
			for _, ev := range expected.Values {
				if av.MetricName == m && ev.MetricName == m {
					if av.Result != ev.Result || !valueParamsEqual(ev, av) {
						return false
					}
				}
			}
		}
	}
	return true
}

func metricsEqual(actual, expected *apis.Test, metricsToCompare ...string) bool {
	actualMetricNames := make([]string, 0)
	for _, m := range actual.Metrics {
		actualMetricNames = append(actualMetricNames, m.Name)
	}
	for _, m := range metricsToCompare {
		if !isIncluded(m, actualMetricNames...) {
			//the requested metric doesn't exist
			return false
		}
		for _, am := range actual.Metrics {
			for _, em := range expected.Metrics {
				if am.Name == m && em.Name == m && am.Description != em.Description {
					return false
				}
			}
		}
	}
	return true
}

func isIncluded(element string, inList ...string) bool {
	for _, m := range inList {
		if m == element {
			return true
		}
	}
	return false
}

func valueParamsEqual(expected, actual apis.Value) bool {
	for i, p := range expected.Parameters {
		if actual.Parameters[i].Name != p.Name || actual.Parameters[i].Value != p.Value {
			return false
		}
	}
	return true
}

func firstMetricByParam(testExec *apis.TestExecution, metricName string, param apis.ValueParameter) float64 {
	for _, v := range testExec.Values {
		if v.MetricName == metricName {
			for _, p := range v.Parameters {
				if p == param {
					return v.Result
				}
			}
		}
	}
	return 0.0
}

func propertiesEqual(actual, expected *apis.Report, propsToCompare ...string) bool {
	for _, p := range propsToCompare {
		if actual.Properties[p] == "" || actual.Properties[p] != expected.Properties[p] {
			return false
		}
	}
	return true
}

func idsIncluded(executions []apis.TestExecution, ids ...int64) bool {
	for _, id := range ids {
		var found bool
		for _, execution := range executions {
			if execution.ID == id {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func containsPermission(perms []apis.Permission, permission *apis.Permission) bool {
	for _, p := range perms {
		if p.AccessLevel == permission.AccessLevel &&
			p.AccessType == permission.AccessType {
			return true
		}
	}
	return false
}
