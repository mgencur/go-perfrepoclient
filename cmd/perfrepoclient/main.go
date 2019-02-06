package main

import (
	"fmt"

	"github.com/PerfCake/go-perfrepoclient/pkg/apis"
	"github.com/PerfCake/go-perfrepoclient/pkg/client"
)

func main() {
	v := &apis.Test{
		Name:        "Martin Gencur",
		GroupID:     "perfrepouser",
		UID:         "uiddddddddddddddd",
		Description: "My description",
		Metrics: []apis.Metric{
			apis.Metric{
				Comparator:  apis.MetricComparatorLB,
				Name:        "metric1",
				Description: "my description",
			},
		},
	}

	client := client.NewPerfRepoClient("http://localhost:8080/testing-repo/rest", "perfrepouser", "perfrepouser1.")

	id, err := client.CreateTest(v)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("ID:", id)

	test, err := client.GetTest(id)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Test: %+v", test)
}
