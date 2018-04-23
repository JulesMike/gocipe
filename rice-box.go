package main

import (
	"github.com/GeertJohan/go.rice/embedded"
	"time"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "bootstrap.go.tmpl",
		FileModTime: time.Unix(1524249429, 0),
		Content:     string("package main\n\nimport (\n\t\"database/sql\"\n\t\"log\"\n\t\"os\"\n\n\t\"github.com/joho/godotenv\"\n\t_ \"github.com/lib/pq\"\n)\n\nconst (\n\t//EnvironmentProd represents production environment\n\tEnvironmentProd = \"PROD\"\n\n\t//EnvironmentDev represents development environment\n\tEnvironmentDev  = \"DEV\"\n)\n\nvar (\n\tenv string\n    db  *sql.DB\n\t{{range .Settings}}\n\t// {{.Name}} {{.Description}}\n\t{{.Name}} {{.Type}}\n\t{{end}}\n)\n\nfunc bootstrap() {\n\tvar err error\n\n\tgodotenv.Load()\n\n\tdsn := os.Getenv(\"DSN\")\n\tenv = os.Getenv(\"ENV\")\n\n\tif env == \"\" {\n\t\tenv = EnvironmentProd\n\t}\n\n\tif dsn == \"\" {\n\t\tlog.Fatal(\"Environment variable DSN must be defined. Example: postgres://user:pass@host/db?sslmode=disable\")\n\t}\n\n\tdb, err = sql.Open(\"postgres\", dsn)\n\tif err == nil {\n\t\tlog.Println(\"Connected to database successfully.\")\n\t} else if (env == EnvironmentDev) {\n\t\tlog.Println(\"Database connection failed: \", err)\n\t} else {\n\t\tlog.Fatal(\"Database connection failed: \", err)\n\t}\n\n\terr = db.Ping()\n\tif err == nil {\n\t\tlog.Println(\"Pinged database successfully.\")\n\t} else if (env == EnvironmentDev) {\n\t\tlog.Println(\"Database ping failed: \", err)\n\t} else {\n\t\tlog.Fatal(\"Database ping failed: \", err)\n\t}\n}"),
	}
	file3 := &embedded.EmbeddedFile{
		Filename:    "crud.go.tmpl",
		FileModTime: time.Unix(1524507050, 0),
		Content:     string("package {{.Package}}\n\nimport \"database/sql\"\n\nvar db *sql.DB\n\n// Inject allows injection of services into the package\nfunc Inject(database *sql.DB) {\n\tdb = database\n}\n\n// New return a new {{.Entity.Name}} instance\nfunc New() *{{.Entity.Name}} {\n\tentity := new({{.Entity.Name}})\n\t{{range .Entity.Fields}}entity.{{.Property.Name}} = new({{.Property.Type}})\n\t{{end}}\n\treturn entity\n}\n\n// Get returns a single {{.Entity.Name}} from database by primary key\nfunc Get(id int64) (*{{.Entity.Name}}, error) {\n\tvar entity = New()\n\t{{if .Entity.Crud.Hooks.PreRead}}\n    if err := crudPreGet(id); err != nil {\n\t\treturn nil, fmt.Errorf(\"error executing crudPreGet() in Get(%d) for entity '{{.Entity.Name}}': %s\", id, err)\n\t}\n    {{end}}\n\trows, err := db.Query(\"SELECT {{.SQLFieldsSelect}} FROM {{.Entity.Table}} t {{.Joins}}WHERE id = $1 ORDER BY t.id ASC\", id)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\n\tdefer rows.Close()\n\tfor rows.Next() {\n\t\t{{range .JoinVarsDecl}}{{.}}\n\t\t{{end}}\n\n\t\terr := rows.Scan({{.StructFieldsSelect}})\n\t\tif err != nil {\n\t\t\treturn nil, err\n\t\t} \n\t\t\n\t\t{{range .JoinVarsAssgn}}{{.}}\n\t\t{{end}}\n\t}\n\t{{if .Entity.Crud.Hooks.PostRead}}\n\tif err = crudPostGet(entity); err != nil {\n\t\treturn nil, fmt.Errorf(\"error executing crudPostGet() in Get(%d) for entity '{{.Entity.Name}}': %s\", id, err)\n\t}\n\t{{end}}\n\n\treturn entity, nil\n}\n\n// List returns a slice containing {{.Entity.Name}} records\nfunc List(filters []models.ListFilter) ([]*{{.Entity.Name}}, error) {\n\tvar (\n\t\tlist     []*{{.Entity.Name}}\n\t\tsegments []string\n\t\tvalues   []interface{}\n\t\terr      error\n\t)\n\n\tquery := \"SELECT {{.SQLFieldsSelect}} FROM {{.Entity.Table}}\"\n\t{{if .Entity.Crud.Hooks.PreList}}\n    if filters, err = crudPreList(filters); err != nil {\n\t\treturn nil, fmt.Errorf(\"error executing crudPreList() in List(filters) for entity '{{.Entity.Name}}': %s\", err)\n\t}\n    {{end}}\n\tfor i, filter := range filters {\n\t\tsegments = append(segments, filter.Field+\" \"+filter.Operation+\" $\"+strconv.Itoa(i+1))\n\t\tvalues = append(values, filter.Value)\n\t}\n\n\tif len(segments) != 0 {\n\t\tquery += \" WHERE \" + strings.Join(segments, \" AND \")\n\t}\n\n\trows, err := db.Query(query+\" ORDER BY id ASC\", values...)\n\t{{if .HasRelationshipManyMany}}\n\tindexID := make(map[int64]*%s)\n\t{{end}}\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\n\tdefer rows.Close()\n\tfor rows.Next() {\n\t\tentity := New()\n\t\terr := rows.Scan({{.StructFieldsSelect}})\n\t\tif err != nil {\n\t\t\treturn nil, err\n\t\t}\n\n\t\tlist = append(list, entity)\n\t\t{{if .HasRelationshipManyMany}}\n\t\tindexID[*entity.ID] = entity\n\t\t{{end}}\n\t}\n\t{{range .ManyManyFields}}\n\tif related, e := loadRelated(indexID, \"{{.Relationship.Target.ThisID}}\", \"{{.Relationship.Target.ThatID}}\", \"{{.Relationship.Target.Table}}\"); e == nil {\n\t\tfor i, v := range related {\n\t\t\tindexID[i].{{.Property.Name}} = append(indexID[i].{{.Property.Name}}, v)\n\t\t}\n\t} else {\n\t\treturn nil, err\n\t}\n\t{{end}}\n\n\t{{if .Entity.Crud.Hooks.PostList}}\n\tif list, err = crudPostList(list); err != nil {\n\t\treturn nil, fmt.Errorf(\"error executing crudPostList() in List(filters) for entity '{{.Entity.Name}}': %s\", err)\n\t}\n\t{{end}}\n\treturn list, nil\n}\n\n{{if .HasRelationshipManyMany}}\n// loadRelated is a helper function to load related entities\nfunc loadRelated(indexID map[int64]*{{.Entity.Name}}, thisid string, thatid string, pivot string) (map[int64]int64, error) {\n\tvar (\n\t\tplaceholder string\n\t\tvalues  []interface{}\n\t\tidthis  int64\n\t\tidthat  int64\n\t)\n\n\trelated := make(map[int64]int64)\n\n\tc := 1\n\tfor i := range indexID {\n\t\tplaceholder += \"$\" + strconv.Itoa(c) + \",\"\n\t\tvalues = append(values, i)\n\t\tc++\n\t}\n\tplaceholder = strings.TrimRight(placeholder, \",\")\n\n\trows, err := db.Query(\"SELECT \"+thisid+\", \"+thatid+\" FROM \"+pivot+\" WHERE \"+thisid+\" IN (\"+placeholder+\")\", values...)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\n\tfor rows.Next() {\n\t\terr = rows.Scan(&idthis, &idthat)\n\t\tif err != nil {\n\t\t\treturn nil, err\n\t\t}\n\t\trelated[idthis] = idthat\n\t}\n\n\treturn related, nil\n}\n{{end}}\n\n// Delete deletes a {{.Entity.Name}} record from database by id primary key\nfunc Delete(id int64, tx *sql.Tx, autocommit bool) error {\n\tvar (\n\t\terr error\n\t\t{{if .HasRelationshipManyMany}}\n\t\tstmtMmany *sql.Stmt\n\t\t{{end}}\n\t)\n\n\tif tx == nil {\n\t\ttx, err = db.Begin()\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t}\n\n\tstmt, err := tx.Prepare(\"DELETE FROM {{.Entity.Table}} WHERE id = $1\")\n\tif err != nil {\n\t\treturn err\n\t}\n\t{{if .Entity.Crud.Hooks.PreDelete}}\n\tif err := crudPreDelete(id, tx); err != nil {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing crudPreDelete() in Delete(%d) for entity '{{.Entity.Name}}': %s\", id, err)\n\t}\n\t{{end}}\n\t{{range .ManyManyFields}}\t\n\tstmtMmany, err = tx.Prepare(\"DELETE FROM {{.Relationship.Target.Table}} WHERE {{.Relationship.Target.ThisID}} = $1\")\n\n\tif err != nil {\n\t\treturn fmt.Errorf(\"error preparing transaction statement in ManyManyDelete(%d) {{.Relationship.Target.ThisID}}-{{.Relationship.Target.ThatID}} for table '{{.Relationship.Target.Table}}': %s\", id, err)\n\t}\n\n\t_, err = stmtMmany.Exec(id)\n\tif err != nil {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing transaction statement in ManyManyDelete(%d) {{.Relationship.Target.ThisID}}-{{.Relationship.Target.ThatID}} for table '{{.Relationship.Target.Table}}': %s\", id, err)\n\t}\n\t{{end}}\n\t_, err = stmt.Exec(id)\n\tif err != nil {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing transaction statement in Delete(%d) for entity '{{.Entity.Name}}': %s\", id, err)\n\t}\n\t{{if .Entity.Crud.Hooks.PostDelete}}\n\tif err := crudPostDelete(id, tx); err != nil {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"Error executing crudPostDelete() in Delete(%d) for entity '{{.Entity.Name}}': %s\", id, err)\n\t}\n\t{{end}}\n\tif autocommit {\n\t\terr = tx.Commit()\n\t\tif err != nil {\n\t\t\treturn fmt.Errorf(\"error committing transaction in Delete(%d) for '{{.Entity.Name}}': %s\", id, err)\n\t\t}\n\t}\n\n\treturn err\n}\n\n// Delete deletes a {{.Entity.Name}} record from database and sets id to nil\nfunc (entity *{{.Entity.Name}}) Delete(tx *sql.Tx, autocommit bool) error {\n\tvar (\n\t\terr error\n\t\t{{if .HasRelationshipManyMany}}\n\t\tstmtMmany *sql.Stmt\n\t\t{{end}}\n\t)\n\n\tid := *entity.ID\n\n\tif tx == nil {\n\t\ttx, err = db.Begin()\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t}\n\n\tstmt, err := tx.Prepare(\"DELETE FROM {{.Entity.Table}} WHERE id = $1\")\n\tif err != nil {\n\t\treturn err\n\t}\n\t{{if .Entity.Crud.Hooks.PreDelete}}\n\tif err := crudPreDelete(id, tx); err != nil {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing crudPreDelete() in {{.Entity.Name}}.Delete() for ID = %d : %s\", id, err)\n\t}\n\t{{end}}\n\t{{range .ManyManyFields}}\t\n\tstmtMmany, err = tx.Prepare(\"DELETE FROM {{.Relationship.Target.Table}} WHERE {{.Relationship.Target.ThisID}} = $1\")\n\n\tif err != nil {\n\t\treturn fmt.Errorf(\"error preparing transaction statement in ManyManyDelete(%d) {{.Relationship.Target.ThisID}}-{{.Relationship.Target.ThatID}} for table '{{.Relationship.Target.Table}}': %s\", *entity.ID, err)\n\t}\n\n\t_, err = stmtMmany.Exec(*entity.ID)\n\tif err != nil {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing transaction statement in ManyManyDelete(%d) {{.Relationship.Target.ThisID}}-{{.Relationship.Target.ThatID}} for table '{{.Relationship.Target.Table}}': %s\", *entity.ID, err)\n\t}\n\t{{end}}\n\t_, err = stmt.Exec(id)\n\tif err == nil {\n\t\tentity.ID = nil\n\t} else {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing transaction statement in {{.Entity.Name}}.Delete() for ID = %d : %s\", id, err)\n\t}\n\t{{if .Entity.Crud.Hooks.PostDelete}}\n\tif err = crudPostDelete(id, tx); err != nil {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing crudPostDelete() in {{.Entity.Name}}.Delete() for ID = %d : %s\", id, err)\n\t}\n\t{{end}}\n\tif autocommit {\n\t\terr = tx.Commit()\n\t\tif err != nil {\n\t\t\treturn fmt.Errorf(\"error committing transaction in {{.Entity.Name}}.Delete() for ID = %d : %s\", id, err)\n\t\t}\n\t}\n\n\treturn err\n}\n\n// Save either inserts or updates a {{.Entity.Name}} record based on whether or not id is nil\nfunc (entity *{{.Entity.Name}}) Save(tx *sql.Tx, autocommit bool) error {\n\tif entity.ID == nil {\n\t\treturn entity.Insert(tx, autocommit)\n\t}\n\treturn entity.Update(tx, autocommit)\n}\n\n// Insert performs an SQL insert for {{.Entity.Name}} record and update instance with inserted id.\n// Prefer using Save rather than Insert directly.\nfunc (entity *{{.Entity.Name}}) Insert(tx *sql.Tx, autocommit bool) error {\n\tvar (\n\t\tid  int64\n\t\terr error\n\t\t{{if .HasRelationshipManyMany}}\n\t\tstmtMmany *sql.Stmt\n\t\t{{end}}\n\t)\n\n\tif tx == nil {\n\t\ttx, err = db.Begin()\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t}\n\t{{range .BeforeInsert}}{{.}}\n\t{{end}}\n\tstmt, err := tx.Prepare(\"INSERT INTO {{.Entity.Table}} ({{.SQLFieldsInsert}}) VALUES ({{.SQLPlaceholders}}) RETURNING id\")\n\tif err != nil {\n\t\treturn err\n\t}\n\t{{if .Entity.Crud.Hooks.PreCreate}}\n    if err := crudPreCreate(entity, tx); err != nil {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing crudPreCreate() in {{.Entity.Name}}.Insert(): %s\", err)\n\t}\n    {{end}}\n\terr = stmt.QueryRow({{.StructFieldsInsert}}).Scan(&id)\n\tif err == nil {\n\t\tentity.ID = &id\n\t} else {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing transaction statement in {{.Entity.Name}}: %s\", err)\n\t}\n\t{{range .ManyManyFields}}\n\tstmtMmany, err = tx.Prepare(\"INSERT INTO {{.Relationship.Target.Table}} ({{.Relationship.Target.ThisID}}, {{.Relationship.Target.ThatID}}) VALUES ($1, $2)\")\n\t\n\tif err != nil {\n\t\treturn fmt.Errorf(\"error preparing transaction statement in ManyManyInsert(%d) {{.Relationship.Target.ThisID}}-{{.Relationship.Target.ThatID}} for table '{{.Relationship.Target.Table}}': %s\", *entity.ID, err)\n\t}\n\n\tfor _, relatedID := range entity.{{.Property.Name}} {\n\t\t_, err = stmtMmany.Exec(entity.ID, relatedID)\n\t\tif err != nil {\n\t\t\ttx.Rollback()\n\t\t\treturn fmt.Errorf(\"error executing transaction statement in ManyManyInsert(%d) {{.Relationship.Target.ThisID}}-{{.Relationship.Target.ThatID}} for table '{{.Relationship.Target.Table}}': %s\", *entity.ID, err)\n\t\t}\n\t}\n\t{{end}}\n\t{{if .Entity.Crud.Hooks.PostCreate}}\n\tif err := crudPostCreate(entity, tx); err != nil {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing crudPostCreate() in {{.Entity.Name}}.Insert(): %s\", err)\n\t}\n\t{{end}}\n\tif autocommit {\n\t\terr = tx.Commit()\n\t\tif err != nil {\n\t\t\treturn fmt.Errorf(\"error committing transaction in {{.Entity.Name}}.Insert(): %s\", err)\n\t\t}\n\t}\n\n\treturn nil\n}\n\n// Update Will execute an SQLUpdate Statement for {{.Entity.Name}} in the database. Prefer using Save instead of Update directly.\nfunc (entity *{{.Entity.Name}}) Update(tx *sql.Tx, autocommit bool) error {\n\tvar (\n\t\terr error\n\t\t{{if .HasRelationshipManyMany}}\n\t\tstmtMmany *sql.Stmt\n\t\t{{end}}\n\t)\n\n\tif tx == nil {\n\t\ttx, err = db.Begin()\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t}\n\t{{range .BeforeUpdate}}{{.}}\n\t{{end}}\n\tstmt, err := tx.Prepare(\"UPDATE {{.Entity.Table}} SET {{.SQLFieldsUpdate}} WHERE id = $1\")\n\tif err != nil {\n\t\treturn err\n\t}\n\n\t{{if .Entity.Crud.Hooks.PreUpdate}}\n    if err := crudPreUpdate(entity, tx); err != nil {\n\t\ttx.Rollback()\n        return fmt.Errorf(\"error executing crudPreUpdate() in {{.Entity.Name}}.Update(): %s\", err)\n\t}\n    {{end}}\n\t_, err = stmt.Exec({{.StructFieldsUpdate}})\n\tif err != nil {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing transaction statement in {{.Entity.Name}}.Update(): %s\", err)\n\t}\n\t{{range .ManyManyFields}}\n\tstmtMmany, err = tx.Prepare(\"DELETE FROM {{.Relationship.Target.Table}} WHERE {{.Relationship.Target.ThisID}} = $1\")\n\n\tif err != nil {\n\t\treturn fmt.Errorf(\"error preparing transaction statement in ManyManyDelete(%d) {{.Relationship.Target.ThisID}}-{{.Relationship.Target.ThatID}} for table '{{.Relationship.Target.Table}}': %s\", id, err)\n\t}\n\n\t_, err = stmtMmany.Exec(id)\n\tif err != nil {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing transaction statement in ManyManyDelete(%d) {{.Relationship.Target.ThisID}}-{{.Relationship.Target.ThatID}} for table '{{.Relationship.Target.Table}}': %s\", id, err)\n\t}\n\n\tstmtMmany, err = tx.Prepare(\"INSERT INTO {{.Relationship.Target.Table}} ({{.Relationship.Target.ThisID}}, {{.Relationship.Target.ThatID}}) VALUES ($1, $2)\")\n\t\n\tif err != nil {\n\t\treturn fmt.Errorf(\"error preparing transaction statement in ManyManyInsert(%d) {{.Relationship.Target.ThisID}}-{{.Relationship.Target.ThatID}} for table '{{.Relationship.Target.Table}}': %s\", *entity.ID, err)\n\t}\n\n\tfor _, relatedID := range entity.{{.Property.Name}} {\n\t\t_, err = stmtMmany.Exec(entity.ID, relatedID)\n\t\tif err != nil {\n\t\t\ttx.Rollback()\n\t\t\treturn fmt.Errorf(\"error executing transaction statement in ManyManyInsert(%d) {{.Relationship.Target.ThisID}}-{{.Relationship.Target.ThatID}} for table '{{.Relationship.Target.Table}}': %s\", *entity.ID, err)\n\t\t}\n\t}\n\t{{end}}\n\t{{if .Entity.Crud.Hooks.PostUpdate}}\n\tif err := crudPostUpdate(entity, tx); err != nil {\n\t\ttx.Rollback()\n\t\treturn fmt.Errorf(\"error executing crudPostUpdate() in {{.Entity.Name}}.Update(): %s\", err)\n\t}\n\t{{end}}\n\tif autocommit {\n\t\terr = tx.Commit()\n\t\tif err != nil {\n\t\t\treturn fmt.Errorf(\"error committing transaction in {{.Entity.Name}}.Update(): %s\", err)\n\t\t}\n\t}\n\n\treturn nil\n}"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    "crud_hooks.go.tmpl",
		FileModTime: time.Unix(1524502956, 0),
		Content:     string("package user\n\nimport (\n\t\"database/sql\"\n)\n\n{{if .Hooks.PreRead}}\nfunc crudPreGet(id int64) error {\n\treturn nil\n}\n{{end}}\n{{if .Hooks.PostRead}}\nfunc crudPostGet(entity *{{.Name}}) error {\n\treturn nil\n}\n{{end}}\n\n{{if .Hooks.PreList}}\nfunc crudPreList(filters []models.ListFilter) ([]models.ListFilter, error) {\n\treturn filters, nil\n}\n{{end}}\n{{if .Hooks.PostList}}\nfunc crudPostList(list []*{{.Name}}) ([]*{{.Name}}, error) {\n\treturn list, nil\n}\n{{end}}\n\n{{if .Hooks.PreDelete}}\nfunc crudPreDelete(id int64, tx *sql.Tx) error {\n\treturn nil\n}\n{{end}}\n{{if .Hooks.PostDelete}}\nfunc crudPostDelete(id int64, tx *sql.Tx) error {\n\treturn nil\n}\n{{end}}\n\n\n{{if .Hooks.PreCreate }}\nfunc crudPreCreate(entity *{{.Name}}, tx *sql.Tx) error {\n\treturn nil\n}\n{{end}}\n{{if .Hooks.PreCreate }}\nfunc crudPostCreate(entity *{{.Name}}, tx *sql.Tx) error {\n\treturn nil\n}\n{{end}}\n\n{{if .Hooks.PreUpdate}}\nfunc crudPreUpdate(entity *{{.Name}}, tx *sql.Tx) error {\n\treturn nil\n}\n{{end}}\n{{if .Hooks.PostUpdate}}\nfunc crudPostUpdate(entity *{{.Name}}, tx *sql.Tx) error {\n\treturn nil\n}\n{{end}}"),
	}
	file5 := &embedded.EmbeddedFile{
		Filename:    "http.go.tmpl",
		FileModTime: time.Unix(1524431103, 0),
		Content:     string("package main\n\nimport (\n\t\"log\"\n\t\"net/http\"\n\t\"os\"\n\t\"os/signal\"\n\t\"syscall\"\n\n\t\"github.com/gorilla/mux\"\n)\n\n// serve starts an http server\nfunc serve(route func(prefix string, router *mux.Router) error) {\n\tvar err error\n\tsigs := make(chan os.Signal, 1)\n\tsignal.Notify(sigs, syscall.SIGTERM)\n\n\trouter := mux.NewRouter()\n\terr = route(\"{{.Prefix}}\", router)\n\n\tif err != nil {\n\t\tlog.Fatal(\"Failed to register routes: \", err)\n\t}\n\t\n\tgo func() {\n\t\terr = http.ListenAndServe(\":{{.Port}}\", router)\n\t\tif err != nil {\n\t\t\tlog.Fatal(\"Failed to start http server: \", err)\n\t\t}\n\t}()\n\n\tlog.Println(\"Listening on : {{.Port}}\")\n\t<-sigs\n\tlog.Println(\"Server stopped\")\n}"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1524501596, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "bootstrap.go.tmpl"
			file3, // "crud.go.tmpl"
			file4, // "crud_hooks.go.tmpl"
			file5, // "http.go.tmpl"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`templates`, &embedded.EmbeddedBox{
		Name: `templates`,
		Time: time.Unix(1524501596, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"bootstrap.go.tmpl":  file2,
			"crud.go.tmpl":       file3,
			"crud_hooks.go.tmpl": file4,
			"http.go.tmpl":       file5,
		},
	})
}
