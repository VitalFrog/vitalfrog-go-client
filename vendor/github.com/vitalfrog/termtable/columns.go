package termtable

import (
	"github.com/fatih/color"
	"unicode/utf8"
)

// Field is are regular Field in a row
type Field interface {
	String() string
	Len() int
}

// HeaderField hold a regular field and optionally allows to configure the width of the Field
type HeaderField struct {
	Field Field
	Width *int
}

//
// StringField holds a string
type StringField struct {
	s string
}

func NewStringField(s string) StringField {
	return StringField{s: s}
}
func (sc StringField) String() string {
	return sc.s
}
func (sc StringField) Len() int {
	return utf8.RuneCountInString(sc.s)
}

// ColorField is a Field holding a string which will be colored on cli
type ColorField struct {
	s         string
	printFunc func(a ...interface{}) string
}

func NewColorField(s string, c *color.Color) ColorField {
	return ColorField{s: s, printFunc: c.SprintFunc()}
}

func (cc ColorField) String() string {
	return cc.printFunc(cc.s)
}
func (cc ColorField) Len() int {
	return utf8.RuneCountInString(cc.s)

}

// NewEmptyField returns a StringField holding an empty string ""
func NewEmptyField() StringField {
	return StringField{s: ""}
}
