package svgparser

import (
	"bytes"
	"encoding/xml"
	"golang.org/x/net/html/charset"
	"io"
	"io/ioutil"
	"strings"
)

// ValidationError contains errors which have occured when parsing svg input.
type ValidationError struct {
	msg string
}

func (err ValidationError) Error() string {
	return err.msg
}

// Element is a representation of an SVG element.
type Element struct {
	Name       string
	Attributes map[string]string
	Children   []*Element
	Parent     *Element
	Content    string
}

// NewElement creates element from decoder token.
func NewElement(token xml.StartElement, parent *Element) *Element {
	element := &Element{}
	attributes := make(map[string]string)
	for _, attr := range token.Attr {
		attributes[attr.Name.Local] = attr.Value
	}
	element.Name = token.Name.Local
	element.Attributes = attributes
	element.Parent = parent

	return element
}

// Compare compares two elements.
func (e *Element) Compare(o *Element) bool {
	if e.Name != o.Name || e.Content != o.Content ||
		len(e.Attributes) != len(o.Attributes) ||
		len(e.Children) != len(o.Children) {
		return false
	}

	for k, v := range e.Attributes {
		if v != o.Attributes[k] {
			return false
		}
	}

	for i, child := range e.Children {
		if !child.Compare(o.Children[i]) {
			return false
		}
	}
	return true
}

// DecodeFirst creates the first element from the decoder.
func DecodeFirst(decoder *xml.Decoder) (*Element, error) {
	for {
		token, err := decoder.Token()
		if token == nil && err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		switch element := token.(type) {
		case xml.StartElement:
			return NewElement(element, nil), nil
		}
	}
	return &Element{}, nil
}

// Decode decodes the child elements of element.
func (e *Element) Decode(decoder *xml.Decoder) error {
	for {
		token, err := decoder.Token()
		if token == nil && err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch element := token.(type) {
		case xml.StartElement:
			nextElement := NewElement(element, e)
			err := nextElement.Decode(decoder)
			if err != nil {
				return err
			}

			e.Children = append(e.Children, nextElement)

		case xml.CharData:
			data := strings.TrimSpace(string(element))
			if data != "" {
				e.Content = string(element)
			}

		case xml.EndElement:
			if element.Name.Local == e.Name {
				return nil
			}
		}
	}
	return nil
}

// Parse creates an Element instance from an SVG input.
func Parse(source io.Reader, validate bool) (*Element, error) {
	raw, err := ioutil.ReadAll(source)
	if err != nil {
		return nil, err
	}
	decoder := xml.NewDecoder(bytes.NewReader(raw))
	decoder.CharsetReader = charset.NewReaderLabel
	element, err := DecodeFirst(decoder)
	if err != nil {
		return nil, err
	}
	if err := element.Decode(decoder); err != nil && err != io.EOF {
		return nil, err
	}
	return element, nil
}

func (el Element) MarshalXML(e *xml.Encoder, start xml.StartElement) error {

	openToken := xml.StartElement{
		Name: xml.Name{
			Local: el.Name,
		},
	}

	if len(el.Attributes) > 0 {
		openToken.Attr = []xml.Attr{}

		for key, value := range el.Attributes {
			openToken.Attr = append(openToken.Attr, xml.Attr{
				Name: xml.Name{
					Local: key,
				},
				Value: value,
			})
		}
	}

	if err := e.EncodeToken(openToken); err != nil {
		return err
	}

	if len(el.Content) > 0 {
		if err := e.EncodeToken(xml.CharData(el.Content)); err != nil {
			return err
		}
	}

	for _, c := range el.Children {
		if c != nil {
			if err := c.MarshalXML(e, openToken); err != nil {
				return err
			}
		}
	}

	closeToken := xml.EndElement{openToken.Name}

	if err := e.EncodeToken(closeToken); err != nil {
		return err
	}

	return e.Flush()
}
