// +build e2e

package e2e

import (
	"fmt"
	"os"
	"testing"

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
func TestCreateGetDeleteTestObject(t *testing.T) {
	testIn := test.Test("test1")

	id, err := testClient.CreateTest(testIn)

	if err != nil {
		t.Fatal("Failed to create Test", err.Error())
	}

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
	fmt.Printf("Test: %+v", testOut)
}
