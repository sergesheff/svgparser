package svgparser

import "strings"

// FindID finds the first child with the specified ID.
func (e *Element) FindID(id string) *Element {
	for _, child := range e.Children {
		if childID, ok := child.Attributes["id"]; ok && childID == id {
			return child
		}
		if element := child.FindID(id); element != nil {
			return element
		}
	}
	return nil
}

// FindAll finds all children with the given name.
func (e *Element) FindAll(name string) []*Element {
	var elements []*Element
	for _, child := range e.Children {
		if child.Name == name {
			elements = append(elements, child)
		}
		elements = append(elements, child.FindAll(name)...)
	}
	return elements
}

// FindByCharData finds all children with the given chardata.
func (e *Element) FindByCharData(text string) []*Element {
	if e == nil {
		return nil
	}

	text = strings.ToLower(text)

	var elements []*Element
	for _, child := range e.Children {
		if child == nil {
			continue
		}

		if strings.Contains(strings.ToLower(child.Content), text) {
			elements = append(elements, child)
		}
		elements = append(elements, child.FindByCharData(text)...)
	}
	return elements
}
