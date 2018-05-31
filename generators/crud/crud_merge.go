package crud

import (
	"fmt"
	"strings"

	"github.com/fluxynet/gocipe/util"
)

// generateMerge produces code for database merge of entity (INSERT/ON CONFLICT UPDATE)
func generateMerge(entities map[string]util.Entity, entity util.Entity) (string, error) {
	var (
		before, after, sqlfieldsInsert, sqlfieldsUpdate, sqlPlaceholders, structFields []string
		count                                                                          = 1
	)

	sqlfieldsInsert = append(sqlfieldsInsert, "id")
	structFields = append(structFields, "entity.ID")
	sqlPlaceholders = append(sqlPlaceholders, fmt.Sprintf("$%d", count))
	count++

	for _, field := range entity.Fields {
		if field.Property.Name == "CreatedAt" {
			before = append(before, "entity.CreatedAt = timestamp.TimestampNow()")
		} else if field.Property.Name == "UpdatedAt" {
			before = append(before, "entity.UpdatedAt = timestamp.TimestampNow()")
		}

		sqlPlaceholders = append(sqlPlaceholders, fmt.Sprintf("$%d", count))
		sqlfieldsUpdate = append(sqlfieldsUpdate, fmt.Sprintf("%s = $%d", field.Schema.Field, count))
		sqlfieldsInsert = append(sqlfieldsInsert, fmt.Sprintf("%s", field.Schema.Field))

		if field.Property.Type == "time" {
			prop := strings.ToLower(field.Property.Name)
			before = append(before, fmt.Sprintf("%s, _ := ptypes.Timestamp(entity.%s)", prop, field.Property.Name))
			structFields = append(structFields, fmt.Sprintf("%s", prop))
		} else {
			structFields = append(structFields, fmt.Sprintf("entity.%s", field.Property.Name))
		}

		count++
	}

	return util.ExecuteTemplate("crud/partials/merge.go.tmpl", struct {
		EntityName      string
		PrimaryKey      string
		Before          []string
		After           []string
		Table           string
		SQLFieldsInsert string
		SQLPlaceholders string
		SQLFieldsUpdate string
		StructFields    string
		HasPreHook      bool
		HasPostHook     bool
	}{
		EntityName:      entity.Name,
		PrimaryKey:      entity.PrimaryKey,
		Before:          before,
		After:           after,
		Table:           entity.Table,
		SQLFieldsInsert: strings.Join(sqlfieldsInsert, ", "),
		SQLPlaceholders: strings.Join(sqlPlaceholders, ", "),
		SQLFieldsUpdate: strings.Join(sqlfieldsUpdate, ", "),
		StructFields:    strings.Join(structFields, ", "),
		HasPreHook:      entity.Crud.Hooks.PreSave,
		HasPostHook:     entity.Crud.Hooks.PostSave,
	})
}