package main

import (
	"log"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

type DomElement struct {
	*js.Object
}

func toString(o *js.Object) string {
	if o == nil || o == js.Undefined {
		return ""
	}
	return o.String()
}

func GetElementByID(id string) *DomElement {
	return &DomElement{js.Global.Get(id)}
}

func (de *DomElement) SetTextContent(content string) {
	de.Set("textContent", content)
}

func (de *DomElement) NodeValue() string {
	return toString(de.Get("nodeValue"))
}

func (de *DomElement) ToggleClass(old, new string) {
	oldclasses := strings.Split(toString(de.Get("classList")), " ")
	log.Printf("Oldclasses: %v", oldclasses)
	newclasses := make([]string, 0, len(oldclasses)+1)
	var replaced bool
	for _, class := range oldclasses {
		if class == old {
			replaced = true
			if new != "" {
				newclasses = append(newclasses, new)
			}
		} else {
			newclasses = append(newclasses, class)
		}
	}
	if !replaced {
		newclasses = append(newclasses, new)
	}

	log.Printf("New classes: %v", newclasses)
	de.Set("classList", strings.Join(newclasses, " "))
}

type HTMLTableBody struct {
	*DomElement
}
