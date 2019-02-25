package apis

import (
	"encoding/xml"
	"errors"
	"fmt"
)

type AccessType int

// enumerate values for AccessType
const (
	UnknownAccessType AccessType = iota
	ReadAccessType
	WriteAccessType
)

var accessTypeValues = []string{"Unknown", "READ", "WRITE"}

type AccessLevel int

// enumerate values for AccessLevel
const (
	UnknownAccessLevel AccessLevel = iota
	UserAccessLevel
	GroupAccessLevel
	PublicAccessLevel
)

var accessLevelValues = []string{"Unknown", "USER", "GROUP", "PUBLIC"}

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
	XMLName     xml.Name
	ID          int64       `xml:"id,omitempty"`
	GroupID     int64       `xml:"group-id,omitempty"`
	ReportID    int64       `xml:"report-id,omitempty"`
	UserID      int64       `xml:"user-id,omitempty"`
	AccessType  AccessType  `xml:"access-type,omitempty"`
	AccessLevel AccessLevel `xml:"access-level,omitempty"`
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

func (p *AccessType) String() string {
	return accessTypeValues[*p]
}

// ParseAccessType converts string to its enum representation
func ParseAccessType(value string) (AccessType, error) {
	for i, v := range accessTypeValues {
		if v == value {
			return AccessType(i), nil
		}
	}
	return AccessType(0), errors.New("Unable to parse " + value)
}

// MarshalXML implements custom marshalling for the AccessType enumeration
func (p *AccessType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(p.String(), start)
}

// UnmarshalXML implements unmarshalling for the AccessType enumeration
func (p *AccessType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var accessType string
	err := d.DecodeElement(&accessType, &start)
	parsed, err := ParseAccessType(accessType)
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}

func (l *AccessLevel) String() string {
	return accessLevelValues[*l]
}

// ParseAccessLevel converts string to its enum representation
func ParseAccessLevel(value string) (AccessLevel, error) {
	for i, v := range accessLevelValues {
		if v == value {
			return AccessLevel(i), nil
		}
	}
	return AccessLevel(0), errors.New("Unable to parse " + value)
}

// MarshalXML implements custom marshalling for the AccessLevel enumeration
func (l *AccessLevel) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(l.String(), start)
}

// UnmarshalXML implements unmarshalling for the AccessLevel enumeration
func (l *AccessLevel) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var accessLevel string
	err := d.DecodeElement(&accessLevel, &start)
	parsed, err := ParseAccessLevel(accessLevel)
	if err != nil {
		return err
	}
	*l = parsed
	return nil
}
