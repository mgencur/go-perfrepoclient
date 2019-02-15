package apis

import (
	"encoding/xml"
	"fmt"
	"io"
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

type Report struct {
	XMLName     xml.Name     `xml:"report"`
	ID          int64        `xml:"id,attr,omitempty"`
	Name        string       `xml:"name,attr"`
	Type        string       `xml:"type,attr"`
	User        string       `xml:"user,attr"`
	Permissions []Permission `xml:"permissions>permission,omitempty"`
	Properties  PropertyMap  `xml:"properties"`
}

type Permission struct {
	XMLName     xml.Name `xml:"permission"`
	ID          int64    `xml:"id,omitempty"`
	GroupID     int64    `xml:"group-id,omitempty"`
	ReportID    int64    `xml:"report-id,omitempty"`
	UserID      int64    `xml:"user-id,omitempty"`
	AccessType  string   `xml:"access-type,omitempty"`
	AccessLevel string   `xml:"access-level,omitempty"`
}

type PropertyMap map[string]string

// MarshalXML marshals the property map to XML.
// Go doesn't support marshalling maps out of the box
func (p *PropertyMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	tokens := []xml.Token{start}

	for key, value := range *p {
		startEntry := xml.StartElement{Name: xml.Name{"", "entry"}}
		tokens = append(tokens, startEntry)
		startKey := xml.StartElement{Name: xml.Name{"", "key"}}
		tokens = append(tokens, startKey, xml.CharData(key), startKey.End())
		startValue := xml.StartElement{
			Name: xml.Name{"", "value"},
			Attr: []xml.Attr{
				{
					Name:  xml.Name{"", "name"},
					Value: key,
				},
				{
					Name:  xml.Name{"", "value"},
					Value: value,
				},
			},
		}
		tokens = append(tokens, startValue, startValue.End())
		tokens = append(tokens, startEntry.End())
	}

	tokens = append(tokens, start.End())

	for _, t := range tokens {
		err := e.EncodeToken(t)
		if err != nil {
			return err
		}
	}

	err := e.Flush()
	if err != nil {
		return err
	}

	return nil
}

// Unmarshall provides custom unmarshalling for the PropertyMap type
func (p *PropertyMap) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	propertyMap := map[string]string{}
	for {
		t, err := d.Token()
		if err != nil {
			break
		}
		switch tt := t.(type) {
		case xml.StartElement:
			if tt.Name.Local == "entry" { //parse whole entry and sub-elements
				entryParsed := false
				for {
					if entryParsed {
						break
					}
					tEntry, err := d.Token()
					if err != nil {
						break
					}
					var key, value string
					switch ttEntry := tEntry.(type) {
					case xml.StartElement:
						if ttEntry.Name.Local == "value" { //parse value element
							for _, attr := range ttEntry.Attr {
								if attr.Name.Local == "name" {
									key = attr.Value
								}
								if attr.Name.Local == "value" {
									value = attr.Value
								}
							}
							propertyMap[key] = value
						} else if ttEntry.Name.Local == "key" { //ignore key element
							continue
						} else {
							return fmt.Errorf("Unexpected element: %v", ttEntry)
						}
					case xml.EndElement:
						if ttEntry.Name.Local == "entry" {
							entryParsed = true
						}
					}
				}
			} else {
				return fmt.Errorf("Unexpected element: %v", tt)
			}
		case xml.EndElement:
			if tt.Name == start.Name {
				break
			}
		}
	}
	*p = propertyMap
	return nil
}

type ReportProperty struct {
	ID    int64  `xml:"id,attr,omitempty"`
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
