package rest

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/fluxynet/gocipe/generators"
)

var tmplDelete, _ = template.New("GenerateDelete").Parse(`
//RestDelete is a REST endpoint for DELETE /{{.Endpoint}}/{id}
func RestDelete(w http.ResponseWriter, r *http.Request) {
	var (
		id       int64
		err      error
		response responseSingle
		tx       *sql.Tx
		{{if .Hooks}}stop     bool{{end}}
	)

	vars := mux.Vars(r)
	valid := false
	if _, ok := vars["id"]; ok {
		id, err = strconv.ParseInt(vars["id"], 10, 64)
		valid = err == nil && id > 0
	}

	if !valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, ` + "`" + `{"status": false, "messages": [{"type": "error", "text": "Invalid ID"}]}` + "`" + `)
		return
	}

	response.Entity, err = Get(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, ` + "`" + `{"status": false, "messages": [{"type": "error", "text": "An error occurred"}]}` + "`" + `)
		return
	}

	if response.Entity == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, ` + "`" + `{"status": false, "messages": [{"type": "error", "text": "Entity not found"}]}` + "`" + `)
		return
	}

	tx, err = db.Begin()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, ` + "`" + `{"status": false, "messages": [{"type": "error", "text": "Failed to process"}]}` + "`" + `)
		return
	}
	{{if .PreExecHook}}
	if stop, err = restPreDelete(w, r, id, tx); err != nil {
		tx.Rollback()
		return
	} else if stop {
		return
	}
    {{end}}
	err = response.Entity.Delete(tx, false)
	if err != nil {
		tx.Rollback()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, ` + "`" + `{"status": false, "messages": [{"type": "error", "text": "Delete failed"}]}` + "`" + `)
		return
	}
	{{if .PostExecHook}}
	if stop, err = restPostDelete(w, r, id, tx); err != nil {
		tx.Rollback()
		return
	} else if stop {
		return
	}
	{{end}}
	if err = tx.Commit(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, ` + "`" + `{"status": false, "messages": [{"type": "E", "message": "RestDelete could not commit transaction"}]}` + "`" + `)
		return
	}

	output, err := json.Marshal(response)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, ` + "`" + `{"status": false, "messages": [{"type": "error", "text": "JSON encoding failed"}]}` + "`" + `)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(output))
}
`)

var tmplDeleteHook, _ = template.New("GenerateDeleteHook").Parse(`
{{if .PreExecHook }}
func restPreDelete(w http.ResponseWriter, r *http.Request, id int64, tx *sql.Tx) (bool, error) {
	return false, nil
}
{{end}}
{{if .PostExecHook }}
func restPostDelete(w http.ResponseWriter, r *http.Request, id int64, tx *sql.Tx) (bool, error) {
	return false, nil
}
{{end}}
`)

//GenerateDelete will generate a REST handler function for Delete
func GenerateDelete(structInfo generators.StructureInfo, preExecHook bool, postExecHook bool) (string, error) {
	var output bytes.Buffer
	data := struct {
		Endpoint     string
		PreExecHook  bool
		PostExecHook bool
		Hooks        bool
	}{strings.ToLower(structInfo.Name), preExecHook, postExecHook, preExecHook || postExecHook}

	err := tmplDelete.Execute(&output, data)

	if err != nil {
		return "", err
	}

	return output.String(), nil
}

// GenerateDeleteHook will generate 2 functions: restPreDelete() and restPostDelete()
func GenerateDeleteHook(structInfo generators.StructureInfo, preExecHook bool, postExecHook bool) (string, error) {
	var output bytes.Buffer

	data := new(struct {
		Name         string
		PreExecHook  bool
		PostExecHook bool
	})

	data.Name = structInfo.Name
	data.PreExecHook = preExecHook
	data.PostExecHook = postExecHook

	err := tmplDeleteHook.Execute(&output, data)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
