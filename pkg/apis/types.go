package apis

import "encoding/xml"

const (
	MetricComparatorHB = "HB"
	MetricComparatorLB = "LB"
)

type Metric struct {
	Comparator  string `xml:"comparator,attr"`
	Name        string `xml:"name,attr"`
	ID          string `xml:"id,attr,omitempty"`
	Description string `xml:"description"`
}

type Test struct {
	XMLName     xml.Name `xml:"test"`
	Name        string   `xml:"name,attr"`
	GroupID     string   `xml:"groupId,attr"`
	ID          string   `xml:"id,attr,omitempty"`
	UID         string   `xml:"uid,attr"`
	Description string   `xml:"description"`
	Metrics     []Metric `xml:"metrics>metric"`
}
