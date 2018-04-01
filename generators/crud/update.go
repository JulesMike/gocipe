package crud

import (
	"bytes"
	"strconv"
	"strings"
	"text/template"

	"github.com/fluxynet/gocipe/generators"
)

var tmplUpdate, _ = template.New("GenerateUpdate").Parse(`
//Update Will execute an SQLUpdate Statement for {{.Name}} in the database. Prefer using Save instead of Update directly.
func (entity *{{.Name}}) Update() error {
	_, err := db.Exec("UPDATE {{.TableName}} SET {{.SQLFields}} WHERE id = $1", {{.StructFields}})

	return err
}
`)

//GenerateUpdate returns code to update an entity in database
func GenerateUpdate(structInfo generators.StructureInfo) (string, error) {
	var output bytes.Buffer
	data := new(struct {
		Name         string
		TableName    string
		SQLFields    string
		StructFields string
	})

	data.Name = structInfo.Name
	data.TableName = structInfo.TableName
	data.SQLFields = ""
	data.StructFields = "entity.ID, "

	for i, field := range structInfo.Fields {
		if field.Name == "ID" {
			continue
		}

		data.SQLFields += strings.ToLower(field.Name) + " = $" + strconv.Itoa(i+1) + ", "
		data.StructFields += "entity." + field.Name + ", "
	}
	data.SQLFields = strings.TrimSuffix(data.SQLFields, ", ")
	data.StructFields = strings.TrimSuffix(data.StructFields, ", ")

	err := tmplUpdate.Execute(&output, data)

	if err != nil {
		return "", err
	}

	return output.String(), nil
}
