package apis

import (
	"encoding/xml"
	"time"
)

const (
	MetricComparatorHB = "HB"
	MetricComparatorLB = "LB"
	jaxbDateFormat     = "2006-01-02T15:04:05.999-07:00"
)

type Metric struct {
	Comparator  string `xml:"comparator,attr,omitempty"`
	Name        string `xml:"name,attr"`
	ID          int64  `xml:"id,attr,omitempty"`
	Description string `xml:"description,omitempty"`
}

type Test struct {
	XMLName     xml.Name `xml:"test"`
	Name        string   `xml:"name,attr"`
	GroupID     string   `xml:"groupId,attr"`
	ID          int64    `xml:"id,attr,omitempty"`
	UID         string   `xml:"uid,attr"`
	Description string   `xml:"description,omitempty"`
	Metrics     []Metric `xml:"metrics>metric,omitempty"`
	//TODO: Add TestExecutions
}

//TODO: Implement helper function ToMap
type TestExecution struct {
	XMLName    xml.Name                 `xml:"testExecution"`
	Name       string                   `xml:"name,attr"`
	ID         int64                    `xml:"id,attr,omitempty"`
	Comment    string                   `xml:"comment,omitempty"`
	TestID     int64                    `xml:"testId,attr"`
	TestUID    string                   `xml:"testUid,attr"`
	Started    JaxbTime                 `xml:"started,attr"`
	Parameters []TestExecutionParameter `xml:"parameters>parameter,omitempty"`
	Tags       []Tag                    `xml:"tags>tag,omitempty"`
	Values     []Value                  `xml:"values>value,omitempty"`
}

type TestExecutionParameter struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type Tag struct {
	ID   int64  `xml:"id,attr,omitempty"`
	Name string `xml:"name,attr"`
}

type Value struct {
	MetricComparator string           `xml:"metricComparator,attr,omitempty"`
	MetricName       string           `xml:"metricName,attr"`
	Result           float64          `xml:"result,attr"`
	Parameters       []ValueParameter `xml:"parameters>parameter,omitempty"`
}

type ValueParameter struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type JaxbTime struct {
	time.Time
}

// UnmarshalXMLAttr implements custom unmarshalling of date/time attribute compatible with default JAXB format
func (c *JaxbTime) UnmarshalXMLAttr(attr xml.Attr) error {
	parsed, err := time.Parse(jaxbDateFormat, attr.Value)
	if err != nil {
		return err
	}
	*c = JaxbTime{parsed}
	return nil
}

// MarshalXMLAttr implements custom marshalling of date/time attribute
// compatible with default JAXB format
func (c *JaxbTime) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{
		Name:  name,
		Value: c.String(),
	}, nil
}

func (c *JaxbTime) String() string {
	return c.Format(jaxbDateFormat)
}
