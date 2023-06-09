package types

import (
	"encoding/json"
	"time"
)

// -1 is error value
type SoqlFieldInfoType int

const (
	SoqlFieldInfo_Field                  SoqlFieldInfoType = iota + 1 // field name
	SoqlFieldInfo_FieldSet                                            // fieldset name
	SoqlFieldInfo_Function                                            // function name and parameters
	SoqlFieldInfo_SubQuery                                            // SoqlQuery
	SoqlFieldInfo_Literal_Null                                        // nil
	SoqlFieldInfo_Literal_Int                                         // int64
	SoqlFieldInfo_Literal_Float                                       // float64
	SoqlFieldInfo_Literal_Bool                                        // bool
	SoqlFieldInfo_Literal_String                                      // string
	SoqlFieldInfo_Literal_Blob                                        // []byte
	SoqlFieldInfo_Literal_Date                                        // timer.Time
	SoqlFieldInfo_Literal_DateTime                                    // timer.Time
	SoqlFieldInfo_Literal_Time                                        // timer.Time
	SoqlFieldInfo_Literal_DateTimeRange                               // SoqlTimeRange
	SoqlFieldInfo_Literal_List                                        // []SoqlListItem
	SoqlFieldInfo_ParameterizedValue                                  // string
	SoqlFieldInfo_DateTimeLiteralName                                 // SoqlDateTimeLiteralName
	SoqlFieldInfo_EndOfConstDefinitions_                              // For UnmarshalJSON (internal use)
)

func (t SoqlFieldInfoType) String() string {
	switch t {
	case SoqlFieldInfo_Field:
		return "Field"
	case SoqlFieldInfo_FieldSet:
		return "FieldSet"
	case SoqlFieldInfo_Function:
		return "Function"
	case SoqlFieldInfo_SubQuery:
		return "SubQuery"
	case SoqlFieldInfo_Literal_Null:
		return "Null"
	case SoqlFieldInfo_Literal_Int:
		return "Int"
	case SoqlFieldInfo_Literal_Float:
		return "Float"
	case SoqlFieldInfo_Literal_Bool:
		return "Bool"
	case SoqlFieldInfo_Literal_String:
		return "String"
	case SoqlFieldInfo_Literal_Blob:
		return "Blob"
	case SoqlFieldInfo_Literal_Date:
		return "Date"
	case SoqlFieldInfo_Literal_DateTime:
		return "DateTime"
	case SoqlFieldInfo_Literal_Time:
		return "Time"
	case SoqlFieldInfo_Literal_DateTimeRange:
		return "DateTimeRange"
	case SoqlFieldInfo_Literal_List:
		return "List"
	case SoqlFieldInfo_ParameterizedValue:
		return "ParameterizedValue"
	case SoqlFieldInfo_DateTimeLiteralName:
		return "DateTimeLiteralName"
	default:
		return "Undefined"
	}
}

type SoqlQueryHint struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// NOTE: When adding items, also add them to soqlFieldInfo_unmarshal and UnmarshalJSON.
type SoqlFieldInfo struct {
	Type        SoqlFieldInfoType `json:"type,omitempty"`
	ClassName   string            `json:"-"`                     // (internal use) Used by parser
	Name        []string          `json:"name,omitempty"`        // for Field, FieldSet, Function, ParameterizedValue, DateTimeLiteralName
	Value       interface{}       `json:"value,omitempty"`       // for Literal_*
	AliasName   string            `json:"aliasName,omitempty"`   // for all (optional)
	Parameters  []SoqlFieldInfo   `json:"parameters,omitempty"`  // for Function; It will be null within the filter, sort, select conditions in the execution plan.
	SubQuery    *SoqlQuery        `json:"subQuery,omitempty"`    // for SubQuery; It will be null within the filter, sort, select conditions in the execution plan.
	NotSelected bool              `json:"notSelected,omitempty"` // It appears only in parameters and conditional expressions.
	Aggregated  bool              `json:"aggregated,omitempty"`  // It is an aggregation function result field or not
	Hints       []SoqlQueryHint   `json:"hints,omitempty"`       // TODO: hints
	ColumnId    int               `json:"columnId,omitempty"`    // Column unique id; 1-based; If 0, it is not set.; Unique column Id across all main and sub queries
	ColIndex    int               `json:"colIndex"`              // Column index in the object
	ViewId      int               `json:"viewId,omitempty"`      // View (table/object) unique id; 1-based; If 0, it is not set.
	Key         string            `json:"key,omitempty"`         // (internal use) Base64-encoded, dot-delimited Name field value
}

type soqlFieldInfo_unmarshal struct {
	Type        SoqlFieldInfoType `json:"type,omitempty"`
	Name        []string          `json:"name,omitempty"`
	Value       json.RawMessage   `json:"value"`
	AliasName   string            `json:"aliasName,omitempty"`
	Parameters  []SoqlFieldInfo   `json:"parameters,omitempty"`
	SubQuery    *SoqlQuery        `json:"subQuery,omitempty"`
	NotSelected bool              `json:"notSelected,omitempty"`
	Aggregated  bool              `json:"Aggregated,omitempty"`
	Hints       []SoqlQueryHint   `json:"hints,omitempty"`
	ColumnId    int               `json:"columnId,omitempty"`
	ColIndex    int               `json:"colIndex"`
	ViewId      int               `json:"viewId,omitempty"`
	Key         string            `json:"key,omitempty"`
}

type SoqlListItem struct {
	Type  SoqlFieldInfoType `json:"type,omitempty"`
	Value interface{}       `json:"value,omitempty"` // ParameterizedValue, DateTimeLiteralName: string ; Literal_*: any
}

type SoqlDateTimeLiteralName struct {
	Name string `json:"name,omitempty"`
	N    int    `json:"n,omitempty"`
}

type SoqlTimeRange struct {
	Start time.Time
	End   time.Time
}

// Object (table)
type SoqlObjectInfo struct {
	Name           []string        `json:"name,omitempty"`          // Object (table) or relationship name with namespace (object graph path)
	AliasName      string          `json:"aliasName,omitempty"`     // Alias name
	HasConditions  bool            `json:"hasConditions,omitempty"` // Query has conditions originally. If false and this object is on the right side, prevent performing an inner join.
	InnerJoin      bool            `json:"innerJoin,omitempty"`     // When this object is on the left side, an inner join is performed.
	Hints          []SoqlQueryHint `json:"hints,omitempty"`         // TODO: hints
	PerObjectQuery *SoqlQuery      `json:"perObjectQuery"`          // A query that extracts only the filter and sort conditions and fields related to this object. A simple query, not including function calls, etc.
	ViewId         int             `json:"viewId,omitempty"`        // View (table/object) unique id; 1-based; If 0, it is not set.
	ParentViewId   int             `json:"parentViewId,omitempty"`  // View id of parent (left side on joining) relationship object.
	Key            string          `json:"key,omitempty"`           // (internal use) Base64-encoded, dot-delimited Name field value
}

type SoqlConditionOpcode int

const (
	SoqlConditionOpcode_Noop                   SoqlConditionOpcode = iota // NOOP
	SoqlConditionOpcode_Unknown                                           // Unknown value (Three-valued logic)
	SoqlConditionOpcode_FieldInfo                                         // SoqlFieldInfo
	SoqlConditionOpcode_Not                                               // unary operator not
	SoqlConditionOpcode_And                                               // binary operator and
	SoqlConditionOpcode_Or                                                // binary operator or
	SoqlConditionOpcode_Eq                                                // binary operator =
	SoqlConditionOpcode_NotEq                                             // binary operator !=
	SoqlConditionOpcode_Lt                                                // binary operator <
	SoqlConditionOpcode_Le                                                // binary operator <=
	SoqlConditionOpcode_Gt                                                // binary operator >
	SoqlConditionOpcode_Ge                                                // binary operator >=
	SoqlConditionOpcode_Like                                              // binary operator like
	SoqlConditionOpcode_NotLike                                           // binary operator not like
	SoqlConditionOpcode_In                                                // binary operator in
	SoqlConditionOpcode_NotIn                                             // binary operator not in
	SoqlConditionOpcode_Includes                                          // binary operator includes
	SoqlConditionOpcode_Excludes                                          // binary operator excludes
	soqlConditionOpcode_EndOfConstDefinitions_                            // For UnmarshalJSON (internal use)
)

func (t SoqlConditionOpcode) String() string {
	switch t {
	case SoqlConditionOpcode_Noop:
		return "Noop"
	case SoqlConditionOpcode_Unknown:
		return "Unknown"
	case SoqlConditionOpcode_FieldInfo:
		return "FieldInfo"
	case SoqlConditionOpcode_Not:
		return "Not"
	case SoqlConditionOpcode_And:
		return "And"
	case SoqlConditionOpcode_Or:
		return "Or"
	case SoqlConditionOpcode_Eq:
		return "Eq"
	case SoqlConditionOpcode_NotEq:
		return "NotEq"
	case SoqlConditionOpcode_Lt:
		return "Lt"
	case SoqlConditionOpcode_Le:
		return "Le"
	case SoqlConditionOpcode_Gt:
		return "Gt"
	case SoqlConditionOpcode_Ge:
		return "Ge"
	case SoqlConditionOpcode_Like:
		return "Like"
	case SoqlConditionOpcode_NotLike:
		return "NotLike"
	case SoqlConditionOpcode_In:
		return "In"
	case SoqlConditionOpcode_NotIn:
		return "NotIn"
	case SoqlConditionOpcode_Includes:
		return "Includes"
	case SoqlConditionOpcode_Excludes:
		return "Excludes"
	default:
		return "Undefined"
	}
}

type SoqlCondition struct {
	Opcode SoqlConditionOpcode `json:"opcode,omitempty"`
	Value  SoqlFieldInfo       `json:"value,omitempty"`
}

type soqlCondition_marshalAll struct {
	Opcode SoqlConditionOpcode `json:"opcode,omitempty"`
	Value  SoqlFieldInfo       `json:"value,omitempty"`
}

type soqlCondition_marshalOp struct {
	Opcode SoqlConditionOpcode `json:"opcode,omitempty"`
}

type SoqlOrderByInfo struct {
	Field     SoqlFieldInfo `json:"field,omitempty"`
	Desc      bool          `json:"desc,omitempty"`
	NullsLast bool          `json:"nullsLast,omitempty"`
}

type SoqlOffsetAndLimitClause struct {
	Offset          int64  `json:"offset,omitempty"`          // offset
	Limit           int64  `json:"limit,omitempty"`           // limit; 0 represents not limited.
	OffsetParamName string `json:"offsetParamName,omitempty"` // offset for parameterized query; If set, it has precedence over the value
	LimitParamName  string `json:"limitParamName,omitempty"`  // limit for parameterized query; If set, it has precedence over the value
}

type SoqlForClause struct {
	View           bool `json:"view,omitempty"`           // for view
	Reference      bool `json:"reference,omitempty"`      // for reference
	Update         bool `json:"update,omitempty"`         // for update
	UpdateTracking bool `json:"updateTracking,omitempty"` // for update tracking (set with Update)
	UpdateViewstat bool `json:"updateViewstat,omitempty"` // for update viewstat (set with Update)
}

type SoqlViewGraphLeaf struct {
	Name         string          `json:"name"`                // Name
	ParentViewId int             `json:"parentViewId"`        // View id of parent object on object graph
	QueryId      int             `json:"queryId"`             // Query unique id
	Depth        int             `json:"depth"`               // Depth on object graph
	QueryDepth   int             `json:"queryDepth"`          // Query depth
	Many         bool            `json:"many,omitempty"`      // True if it is one-to-many relationship (subquery)
	InnerJoin    bool            `json:"innerJoin,omitempty"` // Inner join to parent view id
	NonResult    bool            `json:"nonResult,omitempty"` // True if it is subquery on conditions (where | having clause)
	Object       *SoqlObjectInfo `json:"-"`                   // Object
	Query        *SoqlQuery      `json:"-"`                   // Query
}

type SoqlQueryGraphLeaf struct {
	ParentQueryId int        `json:"parentQueryId"` // Query id of parent query
	Depth         int        `json:"depth"`         // Depth on query graph
	IsConditional bool       `json:"isConditional"` // Query is part of filter (where/having) condition or not
	Query         *SoqlQuery `json:"-"`             // Query
}

type SoqlQueryMeta struct {
	Version          string                     `json:"version,omitempty"`          // format version
	Date             time.Time                  `json:"date,omitempty"`             // compiled datetime
	ElapsedTime      time.Duration              `json:"elapsedTime,omitempty"`      // time taken to compile
	Source           string                     `json:"source,omitempty"`           // source
	MaxQueryDepth    int                        `json:"maxQueryDepth,omitempty"`    // max depth of query graph
	MaxViewDepth     int                        `json:"maxViewDepth,omitempty"`     // max depth of object graph
	NextQueryId      int                        `json:"nextQueryId,omitempty"`      // next query id (a number of queries)
	NextViewId       int                        `json:"nextViewId,omitempty"`       // next view id (a number of views)
	NextColumnId     int                        `json:"nextColumnId,omitempty"`     // next column id (a number of columns)
	QueryGraph       map[int]SoqlQueryGraphLeaf `json:"queryGraph,omitempty"`       // query graph (child -> parent)
	ViewGraph        map[int]SoqlViewGraphLeaf  `json:"viewGraph,omitempty"`        // object graph (child -> parent)
	Functions        map[string]struct{}        `json:"functions,omitempty"`        // functions
	Parameters       map[string]struct{}        `json:"parameters,omitempty"`       // parameters
	DateTimeLiterals map[string]struct{}        `json:"dateTimeLiterals,omitempty"` // datetime literals
}

type SoqlQuery struct {
	Fields           []SoqlFieldInfo          `json:"fields,omitempty"`           // Select clause fields; possibly null
	From             []SoqlObjectInfo         `json:"from,omitempty"`             // From clause objects; has at least one element
	Where            []SoqlCondition          `json:"where,omitempty"`            // Where clause conditions; possibly null; Not used in the execution planning phase.
	GroupBy          []SoqlFieldInfo          `json:"groupBy,omitempty"`          // Group by clause fields; possibly null; Not used for "PerObjectQuery"
	Having           []SoqlCondition          `json:"having,omitempty"`           // Having clause conditions; possibly null; Not used for "PerObjectQuery"
	OrderBy          []SoqlOrderByInfo        `json:"orderBy,omitempty"`          // Order by clause fields; possibly null
	OffsetAndLimit   SoqlOffsetAndLimitClause `json:"offsetAndLimit,omitempty"`   // Offset and limit clause
	For              SoqlForClause            `json:"for,omitempty"`              // For clause
	Parent           *SoqlQuery               `json:"-"`                          // Pointer to parent query; Not used for "PerObjectQuery"
	IsAggregation    bool                     `json:"isAggregation,omitempty"`    // It is an aggregation result or not; Not used for "PerObjectQuery"
	IsCorelated      bool                     `json:"isCorelated,omitempty"`      // Co-related query if true
	PostProcessWhere []SoqlCondition          `json:"postProcessWhere,omitempty"` // Post-processing conditions (Conditions to apply after being filtered in the query for each object)
	QueryId          int                      `json:"queryId,omitempty"`          // Query unique id
	Meta             *SoqlQueryMeta           `json:"meta,omitempty"`             // Meta information
}
