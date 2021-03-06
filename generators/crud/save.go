package crud

import (
	"bytes"
	"text/template"

	"github.com/fluxynet/gocipe/generators"
)

var tmplSave, _ = template.New("GenerateSave").Parse(`
// Save either inserts or updates a {{.Name}} record based on whether or not id is nil
func (entity *{{.Name}}) Save(tx *sql.Tx, autocommit bool) error {
	if entity.ID == nil {
		return entity.Insert(tx, autocommit)
	}
	return entity.Update(tx, autocommit)
}
`)

var tmplSaveHook, _ = template.New("GenerateSaveHook").Parse(`
{{if .PreExecHook }}
func crudPreSave(entity *{{.Name}}, tx *sql.Tx) error {
	return nil
}
{{end}}
{{if .PostExecHook }}
func crudPostSave(entity *{{.Name}}, tx *sql.Tx) error {
	return nil
}
{{end}}
`)

//GenerateSave return code to save entity in database
func GenerateSave(structInfo generators.StructureInfo) (string, error) {
	var output bytes.Buffer

	err := tmplSave.Execute(&output, struct{ Name string }{structInfo.Name})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

// GenerateSaveHook will generate 2 functions: crudSavePreExecHook() and crudSavePostExecHook()
func GenerateSaveHook(structInfo generators.StructureInfo, preExecHook bool, postExecHook bool) (string, error) {
	var output bytes.Buffer

	data := new(struct {
		Name         string
		PreExecHook  bool
		PostExecHook bool
	})

	data.Name = structInfo.Name
	data.PreExecHook = preExecHook
	data.PostExecHook = postExecHook

	err := tmplSaveHook.Execute(&output, data)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
