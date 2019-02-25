package apis

import (
	"encoding/xml"
	"errors"
)

type Comparator int

// enumerate values for Comparator
const (
	UnknownComparator Comparator = iota
	LBComparator
	HBComparator
)

var ComparatorValues = []string{"Unknown", "LB", "HB"}

type Metric struct {
	XMLName     xml.Name   `xml:"metric"`
	Comparator  Comparator `xml:"comparator,attr,omitempty"`
	Name        string     `xml:"name,attr"`
	ID          int64      `xml:"id,attr,omitempty"`
	Description string     `xml:"description,omitempty"`
}

type Test struct {
	XMLName     xml.Name `xml:"test"`
	Name        string   `xml:"name,attr"`
	GroupID     string   `xml:"groupId,attr"`
	ID          int64    `xml:"id,attr,omitempty"`
	UID         string   `xml:"uid,attr"`
	Description string   `xml:"description,omitempty"`
	Metrics     []Metric `xml:"metrics>metric,omitempty"`
}

func (c *Comparator) String() string {
	return ComparatorValues[*c]
}

// ParseComparator converts string to its enum representation
func ParseComparator(value string) (Comparator, error) {
	for i, v := range ComparatorValues {
		if v == value {
			return Comparator(i), nil
		}
	}
	return Comparator(0), errors.New("Unable to parse " + value)
}

// UnmarshalXMLAttr implements unmarshalling for the Comparator enumeration
func (c *Comparator) UnmarshalXMLAttr(attr xml.Attr) error {
	parsed, err := ParseComparator(attr.Value)
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}

// MarshalXMLAttr implements custom marshalling for the Comparator enumeration
func (c *Comparator) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{
		Name:  name,
		Value: c.String(),
	}, nil
}
