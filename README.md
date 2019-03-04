# PerfRepo Go Client

This is a client library for [PerfRepo](https://github.com/PerfCake/PerfRepo) written in Go.
The client has operations for manipulating with Tests, TestExecutions, Reports, 
Report Permissions and more.

# How to use the library

1) Download the client

    `go get github.com/PerfCake/go-perfrepoclient`

2) Create an object of type `*client.PerfRepoClient`:

    ```go
    import "github.com/PerfCake/go-perfrepoclient/pkg/client"

    testClient := client.NewClient("http://perf.repo.url", "username", "password")
    ```

3) Call the API:

    * Create a Test object:

        ```go
        import "github.com/PerfCake/go-perfrepoclient/pkg/apis"

        //create a Test object
        perfRepoTest := &apis.Test{
            Name:        "Product XYZ Performance",
            GroupID:     "my_testing_group",
            UID:         "product_xyz_performance",
            Description: "Holds performance numbers for product XYZ",
            Metrics: []apis.Metric{
                apis.Metric{
                    Comparator:  "LB",
                    Name:        "avg_requests_per_second",
                    Description: "Contains the average number of requests per second",
                },
            },
        }

        //call the API to actually send HTTP request to PerfRepo and create the Test
        id, _ := testClient.CreateTest(perfRepoTest)

        //print the id of the created test
        fmt.Println("ID of the test:", id)

        //retrieve the Test object by id
        testBack, _ := testClient.GetTest(id)

        //print the whole object including names of fields
        fmt.Printf("Test object: %+v", testBack)
        ```

    * Create a TestExecution object:

        ```go
        //create a TestExecution object
        testExec := &apis.TestExecution{
            TestID:  id, // the id that was returned by testClient.CreateTest() function
            Name:    "Distributed Mode",
            Started: &apis.JaxbTime{time.Now()},
            // add parameters metadata (detailed information about the test execution)
            Parameters: []apis.TestExecutionParameter{
                {
                    Name:  "git_branch",
                    Value: "master",
                },
                {
                    Name:  "git_commit",
                    Value: "88159a3b498760e0d637b0720401e593cc1f1d5d",
                },
            },
            // add tags metadata (test executions will be searchable through them)
            Tags: []apis.Tag{
                {
                    Name: "distributed",
                },
                {
                    Name: "size4",
                },
            },
            // add actual values
            Values: []apis.Value{
                {
                    MetricName: "avg_requests_per_second",
                    Result:     120.0,
                },
            },
        }

        //call the API to actually send HTTP request to PerfRepo and create the TestExecution
        testExecID, err := testClient.CreateTestExecution(testExec)

        if err != nil {
		    t.Fatal("Failed to create TestExecution", err.Error())
	    }
        ```

    Note: More examples in the `test/e2e` package.

# How to run e2e tests

1) Make sure [PerfRepo is up and running](https://github.com/PerfCake/PerfRepo#set-up-the-application-server) as the tests require it

2) Run the E2E tests

    * Run all E2E tests with PerfRepo running at default location (default
    location is `http://localhost:8080/testing-repo` with username/password 
    `perfrepouser/perfrepouser1.`):

        `make test-e2e`

    * Run all E2E tests with PerfRepo running at specific location 

        `go test -tags=e2e -v -count=1 ./test/e2e --url http://perf.repo.url --user username --pass password`






