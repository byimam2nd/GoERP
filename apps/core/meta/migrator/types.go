package migrator

import "github.com/goerp/goerp/apps/core/meta/types"

type StepType string

const (
	AddColumn    StepType = "ADD_COLUMN"
	AlterColumn  StepType = "ALTER_COLUMN"
	DropColumn   StepType = "DROP_COLUMN"
	CreateIndex  StepType = "CREATE_INDEX"
)

type MigrationStep struct {
	Type      StepType        `json:"type"`
	TableName string          `json:"table_name"`
	Field     types.DocField  `json:"field"`
	OldField  *types.DocField `json:"old_field,omitempty"`
}

type MigrationPlan struct {
	DocType string          `json:"doctype"`
	Version int             `json:"version"`
	Steps   []MigrationStep `json:"steps"`
}
