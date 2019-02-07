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
				Comparator:  apis.MetricComparatorLB,
				Name:        "metric1",
				Description: "this is a test metric 1",
			},
			apis.Metric{
				Comparator:  apis.MetricComparatorLB,
				Name:        "metric2",
				Description: "this is a test metric 2",
			},
			apis.Metric{
				Comparator:  apis.MetricComparatorHB,
				Name:        "multimetric",
				Description: "this is a metric with multiple values",
			},
		},
	}
}

// once is used to initialize r
var once sync.Once

func initSeed() func() {
	return func() {
		seed := time.Now().UTC().UnixNano()
		r = rand.New(rand.NewSource(seed))
		rndMutex = &sync.Mutex{}
	}
}

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
