package apis

import (
	"encoding/xml"
	"io"
	"sort"
	"time"
)

const (
	jaxbDateFormat = "2006-01-02T15:04:05.999-07:00"
)

type TestExecution struct {
	XMLName    xml.Name                 `xml:"testExecution"`
	Name       string                   `xml:"name,attr"`
	ID         int64                    `xml:"id,attr,omitempty"`
	Comment    string                   `xml:"comment,omitempty"`
	TestID     int64                    `xml:"testId,attr"`
	TestUID    string                   `xml:"testUid,attr"`
	Started    *JaxbTime                `xml:"started,attr"`
	Parameters []TestExecutionParameter `xml:"parameters>parameter,omitempty"`
	Tags       []Tag                    `xml:"tags>tag,omitempty"`
	Values     []Value                  `xml:"values>value,omitempty"`
}

type TestExecutionParameter struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type Tag struct {
	XMLName xml.Name `xml:"tag"`
	ID      int64    `xml:"id,attr,omitempty"`
	Name    string   `xml:"name,attr"`
}

type Value struct {
	MetricComparator Comparator       `xml:"metricComparator,attr,omitempty"`
	MetricName       string           `xml:"metricName,attr"`
	Result           float64          `xml:"result,attr"`
	Parameters       []ValueParameter `xml:"parameters>parameter,omitempty"`
}

type ValueParameter struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// Holds data related to an attachment for TestExecution
type Attachment struct {
	File           io.Reader // data
	ContentType    string    // MimeType of the data
	TargetFileName string    // name under which the attachment will be stored in PerfRepo
}

type JaxbTime struct {
	time.Time
}

// SortedTags returns a sorted copy of the tags slice
func (t *TestExecution) SortedTags() []Tag {
	sortedTags := append([]Tag(nil), t.Tags...)
	sort.Slice(sortedTags, func(i, j int) bool {
		return sortedTags[i].Name < sortedTags[j].Name
	})
	return sortedTags
}

// SortedParameters returns a sorted copy of the parameters slice
func (t *TestExecution) SortedParameters() []TestExecutionParameter {
	sortedParams := append([]TestExecutionParameter(nil), t.Parameters...)
	sort.Slice(sortedParams, func(i, j int) bool {
		return sortedParams[i].Name < sortedParams[j].Name
	})
	return sortedParams
}

// ParametersMap returns test execution parameters map where key is the t.Name and
// value is the t.Value
func (t *TestExecution) ParametersMap() map[string]string {
	paramsMap := make(map[string]string)
	for _, p := range t.Parameters {
		paramsMap[p.Name] = p.Value
	}
	return paramsMap
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
