package crud

var tmplModels = `
package models

//ListFilter represents a filter to apply during listing (crud)
type ListFilter struct {
	Field     string
	Operation string
	Value     interface{}
}
`

func generateModels() string {
	return tmplModels
}
