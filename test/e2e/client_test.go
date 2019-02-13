// +build e2e

package e2e

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/PerfCake/go-perfrepoclient/pkg/apis"
	"github.com/PerfCake/go-perfrepoclient/pkg/client"
	"github.com/PerfCake/go-perfrepoclient/test"
)

const (
	perfRepoURL  = "http://localhost:8080/testing-repo/rest"
	perfRepoUser = "perfrepouser"
	perfRepoPass = "perfrepouser1."
)

var testClient *client.PerfRepoClient

func TestMain(m *testing.M) {
	testClient = client.NewPerfRepoClient(perfRepoURL, perfRepoUser, perfRepoPass)
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
		//TODO: Verify testOut.TestExecutions are nil
		t.Fatalf("The returned test: %+v does not match the original test %+v", testOut, testIn)
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

	testExecIn := test.TestExecution(testID)

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
		!metricsEqual(testExecOut, testExecIn, "metric1", "metric2") ||
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

	testExecIn := test.TestExecution(testID)

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
		!metricsEqual(testExecOut, testExecIn, "metric1", "metric2") ||
		firstMetricByParam(testExecOut, "multimetric",
			apis.ValueParameter{Name: "client", Value: "1"}) != 20.0 ||
		firstMetricByParam(testExecOut, "multimetric",
			apis.ValueParameter{Name: "client", Value: "2"}) != 40.0 {
		t.Fatalf("The returned test execution: %+v does not match the original %+v",
			testExecOut, testExecIn)
	}
}

func TestCreateAttachment(t *testing.T) {
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

	testExecIn := test.TestExecution(testID)

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
	att := apis.Attachment{
		File:           strings.NewReader(attachmentText),
		ContentType:    "text/plain",
		TargetFileName: "attachment1.txt",
	}
	attID, err := testClient.CreateAttachment(testExecID, att)
	if err != nil {
		t.Fatal("Failed to create Attachment", err.Error())
	}
	bodyReader, err := testClient.GetAttachment(attID)
	if err != nil {
		t.Fatal("Failed to get Attachment", err.Error())
	}
	defer bodyReader.Close()
	bodyBytes, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		t.Fatal("Unable to read attachment", err.Error())
	}
	if string(bodyBytes) != attachmentText {
		t.Fatal("Unexpected attachment body")
	}
}

func TestCreateGetDeleteReport(t *testing.T) {
	reportIn := test.Report("report", perfRepoUser)

	id, err := testClient.CreateReport(reportIn)

	if err != nil {
		t.Fatal("Failed to create Report", err.Error())
	}
	defer func() {
		if err := testClient.DeleteReport(id); err != nil {
			t.Fatal(err.Error())
		}

		if _, err = testClient.GetReport(id); err == nil {
			// https://github.com/PerfCake/PerfRepo/issues/93
			//!strings.Contains(err.Error(), "doesn't exist") {
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

func paramsEqual(actual, expected *apis.TestExecution) bool {
	for i, p := range expected.Parameters {
		if actual.Parameters[i].Name != p.Name || actual.Parameters[i].Value != p.Value {
			return false
		}
	}
	return true
}

func tagsEqual(actual, expected *apis.TestExecution) bool {
	for i, tag := range expected.Tags {
		if actual.Tags[i].Name != tag.Name {
			return false
		}
	}
	return true
}

func metricsEqual(actual, expected *apis.TestExecution, metricsToCompare ...string) bool {
	for i, v := range expected.Values {
		if isIncluded(v.MetricName, metricsToCompare...) {
			if actual.Values[i].MetricName != v.MetricName ||
				actual.Values[i].Result != v.Result ||
				!valueParamsEqual(actual.Values[i], v) {
				return false
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

func valuesEqual(actual, expected *apis.TestExecution) bool {
	for i, v := range expected.Values {
		if actual.Values[i].MetricName != v.MetricName ||
			actual.Values[i].Result != v.Result ||
			!valueParamsEqual(actual.Values[i], v) {
			return false
		}
	}
	return true
}

func valueParamsEqual(actual, expected apis.Value) bool {
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
	for i, p := range expected.Properties {
		if isIncluded(p.Key, propsToCompare...) {
			if actual.Properties[i].Value.Name != p.Value.Name ||
				actual.Properties[i].Value.Value != p.Value.Value {
				return false
			}
		}
	}
	return true
}
