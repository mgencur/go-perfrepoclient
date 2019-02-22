package test

import (
	"math/rand"
	"sync"
	"time"

	"github.com/PerfCake/go-perfrepoclient/pkg/apis"
)

const (
	clientGroup   = "perfrepouser"
	letterBytes   = "abcdefghijklmnopqrstuvwxyz"
	randSuffixLen = 8
)

// r is used by AppendRandomString to generate a random string. It is seeded with the time
// at import so the strings will be different between test runs.
var (
	r        *rand.Rand
	rndMutex *sync.Mutex
)

// Test creates a new Test object with the given name
func Test(name string) *apis.Test {
	salt := RandomString()
	return &apis.Test{
		Name:        name + salt,
		GroupID:     clientGroup,
		UID:         name + "uid" + salt,
		Description: "This is a test object",
		Metrics: []apis.Metric{
			apis.Metric{
				Comparator:  "LB",
				Name:        "metric1",
				Description: "this is a test metric 1",
			},
			apis.Metric{
				Comparator:  "LB",
				Name:        "metric2",
				Description: "this is a test metric 2",
			},
			apis.Metric{
				Comparator:  "HB",
				Name:        "multimetric",
				Description: "this is a metric with multiple values",
			},
		},
	}
}

// once is used to initialize r
var once sync.Once

// RandomString will generate a random string.
func RandomString() string {
	once.Do(initSeed())
	result := make([]byte, randSuffixLen)
	rndMutex.Lock()
	for i := range result {
		result[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	rndMutex.Unlock()
	return string(result)
}

func initSeed() func() {
	return func() {
		seed := time.Now().UTC().UnixNano()
		r = rand.New(rand.NewSource(seed))
		rndMutex = &sync.Mutex{}
	}
}

// Execution creates a new TestExecution object referencing given test via testId
func DefaultExecution(testID int64) *apis.TestExecution {
	started := &apis.JaxbTime{time.Now()}
	params := []apis.TestExecutionParameter{
		{
			Name:  "param1",
			Value: "value1",
		},
		{
			Name:  "param2",
			Value: "value2",
		},
	}
	tags := []apis.Tag{
		{
			Name: "tag1",
		},
		{
			Name: "tag2",
		},
	}
	return Execution(testID, started, params, tags)
}

func Execution(testID int64, started *apis.JaxbTime, params []apis.TestExecutionParameter, tags []apis.Tag) *apis.TestExecution {
	salt := RandomString()
	testExec := apis.TestExecution{
		TestID:  testID,
		Name:    "execution" + salt,
		Started: started,
		Values: []apis.Value{
			{
				MetricName: "metric1",
				Result:     12.0,
			},
			{
				MetricName: "metric2",
				Result:     8.0,
			},
			{
				MetricName: "multimetric",
				Result:     20.0,
				Parameters: []apis.ValueParameter{
					{
						Name:  "client",
						Value: "1",
					},
				},
			},
			{
				MetricName: "multimetric",
				Result:     40.0,
				Parameters: []apis.ValueParameter{
					{
						Name:  "client",
						Value: "2",
					},
				},
			},
		},
	}
	for _, param := range params {
		testExec.Parameters = append(testExec.Parameters, param)
	}
	for _, tag := range tags {
		testExec.Tags = append(testExec.Tags, tag)
	}
	return &testExec
}

func ReducedExecution(testID int64) *apis.TestExecution {
	salt := RandomString()
	testExec := apis.TestExecution{
		TestID:  testID,
		Name:    "reduced execution" + salt,
		Started: &apis.JaxbTime{time.Now()},
		Values: []apis.Value{
			{
				MetricName: "metric1",
				Result:     7.0,
			},
			{
				MetricName: "multimetric",
				Result:     77.0,
				Parameters: []apis.ValueParameter{
					{
						Name:  "client",
						Value: "30",
					},
				},
			},
		},
		Parameters: []apis.TestExecutionParameter{
			{
				Name:  "param1",
				Value: "differentValue",
			},
		},
		Tags: []apis.Tag{
			{
				Name: "differentTag",
			},
		},
	}
	return &testExec
}

// InvalidTestExecution creates a TestExecution that has multiple metrics with same
// name but without any parameters that would differentiate them
func InvalidTestExecution(testId int64) *apis.TestExecution {
	salt := RandomString()
	return &apis.TestExecution{
		TestID:  testId,
		Name:    "execution" + salt,
		Started: &apis.JaxbTime{time.Now()},
		Values: []apis.Value{
			{
				MetricName: "multimetric",
				Result:     20.0,
			},
			{
				MetricName: "multimetric",
				Result:     40.0,
			},
		},
	}
}

// Report creates a new Report object with the given name
func Report(name string, username string) *apis.Report {
	salt := RandomString()
	return &apis.Report{
		Name: name + salt,
		Type: "TestReport",
		User: username,
		Properties: map[string]string{
			"property1": "value",
		},
	}
}
