package postprocess

import (
	"strconv"

	"github.com/shellyln/go-nameutil/nameutil"
	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
)

func (ctx *normalizeQueryContext) addUnselectedSelectFields(
	q *SoqlQuery, fields []SoqlFieldInfo, fieldMap map[string][]string, level int) {

	fieldsLen := len(fields)

	if level == 0 {
		for i := 0; i < fieldsLen; i++ {
			// TODO: key is probably calculated.
			key := nameutil.MakeDottedKeyIgnoreCase(fields[i].Name, len(fields[i].Name))
			switch fields[i].Type {
			case SoqlFieldInfo_Field:
				if _, ok := fieldMap[key]; !ok {
					fieldMap[key] = fields[i].Name
				}
			}
		}
	}

	for i := 0; i < fieldsLen; i++ {
		// TODO: key is probably calculated.
		key := nameutil.MakeDottedKeyIgnoreCase(fields[i].Name, len(fields[i].Name))
		switch fields[i].Type {
		case SoqlFieldInfo_Field:
			if _, ok := fieldMap[key]; !ok {
				f := fields[i]
				fieldMap[key] = f.Name
				f.NotSelected = true
				q.Fields = append(q.Fields, f)
			}
		case SoqlFieldInfo_Function:
			ctx.addUnselectedSelectFields(q, fields[i].Parameters, fieldMap, level+1)
		}
	}
}

func (ctx *normalizeQueryContext) addUnselectedConditionalFields(
	q *SoqlQuery, conditions []SoqlCondition, fieldMap map[string][]string) error {

	conditionsLen := len(conditions)

	for i := 0; i < conditionsLen; i++ {
		switch conditions[i].Opcode {
		case SoqlConditionOpcode_FieldInfo:
			{
				// TODO: key is probably calculated.
				key := nameutil.MakeDottedKeyIgnoreCase(conditions[i].Value.Name, len(conditions[i].Value.Name))
				switch conditions[i].Value.Type {
				case SoqlFieldInfo_Field:
					if _, ok := fieldMap[key]; !ok {
						// TODO: check it is co-related (field in the ancestor object); siblings are error
						f := conditions[i].Value
						fieldMap[key] = f.Name
						f.NotSelected = true
						q.Fields = append(q.Fields, f)
					}
				case SoqlFieldInfo_Function:
					ctx.addUnselectedSelectFields(q, conditions[i].Value.Parameters, fieldMap, 1)
				}
			}
		}
	}
	return nil
}

func (ctx *normalizeQueryContext) addUnselectedFields(q *SoqlQuery) error {
	fieldMap := make(map[string][]string)

	ctx.addUnselectedSelectFields(q, q.Fields, fieldMap, 0)

	if err := ctx.addUnselectedConditionalFields(q, q.Where, fieldMap); err != nil {
		return err
	}

	ctx.addUnselectedSelectFields(q, q.GroupBy, fieldMap, 1)

	if err := ctx.addUnselectedConditionalFields(q, q.Having, fieldMap); err != nil {
		return err
	}

	for i := 0; i < len(q.OrderBy); i++ {
		key := nameutil.MakeDottedKeyIgnoreCase(q.OrderBy[i].Field.Name, len(q.OrderBy[i].Field.Name))
		if _, ok := fieldMap[key]; !ok {
			f := q.OrderBy[i].Field
			fieldMap[key] = f.Name
			f.NotSelected = true
			q.Fields = append(q.Fields, f)
		}
	}
	return nil
}

func distributeNotOperators(conditions []SoqlCondition) []SoqlCondition {
	for i := 0; i < len(conditions); i++ {
		switch conditions[i].Opcode {
		case SoqlConditionOpcode_Not:
			for j := 0; j < i; j++ {
				switch conditions[j].Opcode {
				case SoqlConditionOpcode_And:
					conditions[j].Opcode = SoqlConditionOpcode_Or
				case SoqlConditionOpcode_Or:
					conditions[j].Opcode = SoqlConditionOpcode_And
				case SoqlConditionOpcode_Eq:
					conditions[j].Opcode = SoqlConditionOpcode_NotEq
				case SoqlConditionOpcode_NotEq:
					conditions[j].Opcode = SoqlConditionOpcode_Eq
				case SoqlConditionOpcode_Lt:
					conditions[j].Opcode = SoqlConditionOpcode_Ge
				case SoqlConditionOpcode_Le:
					conditions[j].Opcode = SoqlConditionOpcode_Gt
				case SoqlConditionOpcode_Gt:
					conditions[j].Opcode = SoqlConditionOpcode_Le
				case SoqlConditionOpcode_Ge:
					conditions[j].Opcode = SoqlConditionOpcode_Lt
				case SoqlConditionOpcode_Like:
					conditions[j].Opcode = SoqlConditionOpcode_NotLike
				case SoqlConditionOpcode_NotLike:
					conditions[j].Opcode = SoqlConditionOpcode_Like
				case SoqlConditionOpcode_In:
					conditions[j].Opcode = SoqlConditionOpcode_NotIn
				case SoqlConditionOpcode_NotIn:
					conditions[j].Opcode = SoqlConditionOpcode_In
				case SoqlConditionOpcode_Includes:
					conditions[j].Opcode = SoqlConditionOpcode_Excludes
				case SoqlConditionOpcode_Excludes:
					conditions[j].Opcode = SoqlConditionOpcode_Includes
				}
			}
			conditions[i].Opcode = SoqlConditionOpcode_Noop
		}
	}
	return compactConditions(conditions)
}

func (ctx *normalizeQueryContext) assignColumnIdToField(
	field *SoqlFieldInfo, isFunctionParameter bool, aggregated bool) {

	if field.ColumnId == 0 {
		var key string
		switch field.Type {
		case SoqlFieldInfo_Field:
			// TODO: key is probably calculated.
			key = field.Key

			if id, ok := ctx.columnIdMap[key]; !ok {
				field.ColumnId = ctx.columnId
				ctx.columnIdMap[key] = ctx.columnId
				ctx.columnId++
			} else {
				field.ColumnId = id
			}
		case SoqlFieldInfo_Function:
			if (!aggregated && !isFunctionParameter) || (aggregated && field.Aggregated) {
				field.ColumnId = ctx.columnId
				ctx.columnId++
			}
		case SoqlFieldInfo_FieldSet, SoqlFieldInfo_SubQuery:
			field.ColumnId = ctx.columnId
			ctx.columnId++
		}
	}

	switch field.Type {
	case SoqlFieldInfo_Function:
		for i := 0; i < len(field.Parameters); i++ {
			ctx.assignColumnIdToField(&field.Parameters[i], true, aggregated)
		}
	}
}

func (ctx *normalizeQueryContext) assignColumnIds(q *SoqlQuery) {

	// Apply ColumnId to Select clause fields
	for i := 0; i < len(q.Fields); i++ {
		field := &q.Fields[i]
		ctx.assignColumnIdToField(field, false, false)
	}

	for i := 0; i < len(q.OrderBy); i++ {
		field := &q.OrderBy[i].Field
		ctx.assignColumnIdToField(field, false, false)
	}

	for i := 0; i < len(q.GroupBy); i++ {
		field := &q.GroupBy[i]
		ctx.assignColumnIdToField(field, false, false)
	}

	// Apply ColumnId to Where clause fields
	for i := 0; i < len(q.Where); i++ {
		condition := &q.Where[i]
		ctx.assignColumnIdToField(&condition.Value, false, false)
	}

	// Apply ColumnId to Having clause fields and aggregation functions
	for i := 0; i < len(q.Having); i++ {
		condition := &q.Having[i]
		ctx.assignColumnIdToField(&condition.Value, false, true)
	}
	for i := 0; i < len(q.Having); i++ {
		condition := &q.Having[i]
		ctx.assignColumnIdToField(&condition.Value, false, false)
	}

	// Apply ColumnId to Select clause aggregation functions
	for i := 0; i < len(q.Fields); i++ {
		field := &q.Fields[i]
		switch field.Type {
		case SoqlFieldInfo_Function:
			ctx.assignColumnIdToField(field, false, true)
		}
	}
}

func (ctx *normalizeQueryContext) assignImplicitAliasNames(q *SoqlQuery, fieldAliasMap map[string]*SoqlFieldInfo) {
	noAliasNameColumnIds := make(map[int]struct{})
	exprCount := 0

	for i := 0; i < len(q.Fields); i++ {
		field := &q.Fields[i]
		if field.AliasName == "" {
			switch field.Type {
			case SoqlFieldInfo_Field:
				if _, ok := noAliasNameColumnIds[field.ColumnId]; !ok {
					noAliasNameColumnIds[field.ColumnId] = struct{}{}
					break
				}
				fallthrough
			case SoqlFieldInfo_Function:
				for {
					exprName := "expr" + strconv.Itoa(exprCount) // exprName is lower case
					exprCount++

					if _, ok := fieldAliasMap[exprName]; !ok {
						field.AliasName = exprName
						break
					}
				}
			}
		}
	}
}
