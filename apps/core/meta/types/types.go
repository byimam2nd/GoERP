package types

type FieldType string

const (
	Data     FieldType = "Data"
	Int      FieldType = "Int"
	Float    FieldType = "Float"
	Currency FieldType = "Currency"
	Date     FieldType = "Date"
	DateTime FieldType = "DateTime"
	Check    FieldType = "Check"
	Link     FieldType = "Link"
	Select   FieldType = "Select"
	Text     FieldType = "Text"
	Table    FieldType = "Table"
	SectionBreak FieldType = "Section Break"
	ColumnBreak  FieldType = "Column Break"
)

type DocField struct {
	Name      string    `json:"name"`
	Label     string    `json:"label"`
	FieldType FieldType `json:"fieldtype"`
	Options   string    `json:"options,omitempty"` // For Link (DocType name) or Select (options list)
	PermLevel int       `json:"perm_level,omitempty"`
	Required  bool      `json:"required,omitempty"`
	Unique    bool      `json:"unique,omitempty"`
	InListView bool     `json:"in_list_view,omitempty"`
	ReadOnly  bool      `json:"read_only,omitempty"`
	Default   string    `json:"default,omitempty"`
	FetchFrom string    `json:"fetch_from,omitempty"` // Format: "link_fieldname.source_fieldname"
	Formula   string    `json:"formula,omitempty"`    // Example: "qty * rate"
	SearchIndex bool    `json:"search_index,omitempty"`
}

type DocPerm struct {
	Role      string `json:"role"`
	PermLevel int    `json:"perm_level,omitempty"`
	Read      bool   `json:"read"`
	Write     bool   `json:"write"`
	Create    bool   `json:"create"`
	Delete    bool   `json:"delete"`
	Submit    bool   `json:"submit"`
	Cancel    bool   `json:"cancel"`
}

type DocType struct {
	Name        string     `json:"name"`
	Module      string     `json:"module"`
	IsSubmittable bool     `json:"is_submittable,omitempty"`
	IsTree      bool       `json:"is_tree,omitempty"`
	ClientScript string    `json:"client_script,omitempty"` // Custom JS logic
	AllowMapTo   []string  `json:"allow_map_to,omitempty"`   // Targets for mapping (e.g. ["SalesInvoice"])
	Fields      []DocField `json:"fields"`
	Permissions []DocPerm  `json:"permissions"`
}
