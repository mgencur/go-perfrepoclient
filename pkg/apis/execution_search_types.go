package apis

import (
	"encoding/xml"
	"errors"
)

type GroupFilter int

// enumerate values for GroupFilter
const (
	UnknownGroupFilter GroupFilter = iota
	MyGroupFilter
	AllGroupFilter
)

var groupFilterValues = []string{"Unknown", "MY_GROUPS", "ALL_GROUPS"}

type OrderBy int

// enumerate values for OrderBy
const (
	UnknownOrderBy OrderBy = iota
	DateAscOrderBy
	DateDescOrderBy
	ParameterAscOrderBy
	ParameterDescOrderBy
	VersionAscOrderBy
	VersionDescOrderBy
	NameAscOrderBy
	NameDescOrderBy
	UIDAscOrderBy
	UIDDescOrderBy
	GroupIDAscOrderBy
	GroupIDDescOrderBy
)

var orderByValues = []string{"Unknown", "DATE_ASC", "DATE_DESC", "PARAMETER_ASC",
	"PARAMETER_DESC", "VERSION_ASC", "VERSION_DESC",
	"NAME_ASC", "NAME_DESC", "UID_ASC", "UID_DESC", "GROUP_ID_ASC",
	"GROUP_ID_DESC"}

type TestExecutionSearch struct {
	XMLName          xml.Name            `xml:"test-execution-search"`
	GroupFilter      GroupFilter         `xml:"group-filter,omitempty"`
	IDS              *[]int64            `xml:"ids>id,omitempty"` //use pointer to array so that the parent ids element can be ommitted if empty/nil
	LabelParameter   string              `xml:"labelParameter,omitempty"`
	LimitFrom        int                 `xml:"limit-from,omitempty"`
	HowMany          int                 `xml:"how-many,omitempty"`
	OrderBy          OrderBy             `xml:"order-by,omitempty"`
	OrderByParameter string              `xml:"orderByParameter,omitempty"`
	Parameters       []CriteriaParameter `xml:"parameters>parameter,omitempty"`
	ExecutedAfter    *JaxbTime           `xml:"executed-after,omitempty"`
	ExecutedBefore   *JaxbTime           `xml:"executed-before,omitempty"`
	Tags             string              `xml:"tags,omitempty"`
	TestName         string              `xml:"test-name,omitempty"`
	TestUID          string              `xml:"test-uid,omitempty"`
}

type CriteriaParameter struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

// TestExecutions type holds results of SearchTestExecutions operation
type TestExecutions struct {
	XMLName        xml.Name        `xml:"testExecutions"`
	TestExecutions []TestExecution `xml:"testExecution"`
}

func (p *GroupFilter) String() string {
	return groupFilterValues[*p]
}

// ParseGroupFilter converts string to its enum representation
func ParseGroupFilter(value string) (GroupFilter, error) {
	for i, v := range groupFilterValues {
		if v == value {
			return GroupFilter(i), nil
		}
	}
	return GroupFilter(0), errors.New("Unable to parse " + value)
}

// MarshalXML implements custom marshalling for the GroupFilter enumeration
func (p *GroupFilter) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(p.String(), start)
}

// UnmarshalXML implements unmarshalling for the GroupFilter enumeration
func (p *GroupFilter) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var groupFilter string
	err := d.DecodeElement(&groupFilter, &start)
	parsed, err := ParseGroupFilter(groupFilter)
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}

func (p *OrderBy) String() string {
	return orderByValues[*p]
}

// ParseOrderBy converts string to its enum representation
func ParseOrderBy(value string) (OrderBy, error) {
	for i, v := range orderByValues {
		if v == value {
			return OrderBy(i), nil
		}
	}
	return OrderBy(0), errors.New("Unable to parse " + value)
}

// MarshalXML implements custom marshalling for the OrderBy enumeration
func (p *OrderBy) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(p.String(), start)
}

// UnmarshalXML implements unmarshalling for the OrderBy enumeration
func (p *OrderBy) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var orderBy string
	err := d.DecodeElement(&orderBy, &start)
	parsed, err := ParseOrderBy(orderBy)
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}
