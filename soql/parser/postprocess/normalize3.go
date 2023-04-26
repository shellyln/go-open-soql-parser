package postprocess

import (
	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
)

func (ctx *normalizeQueryContext) applyColIndexToFields(fields []SoqlFieldInfo, colIndexMap map[string]int) {
	for i := range fields {
		switch fields[i].Type {
		case SoqlFieldInfo_Field:
			if idx, ok := colIndexMap[fields[i].Key]; ok {
				fields[i].ColIndex = idx
			}
		case SoqlFieldInfo_Function:
			ctx.applyColIndexToFields(fields[i].Parameters, colIndexMap)
		}
	}
}

func (ctx *normalizeQueryContext) applyColIndexToConditions(conditions []SoqlCondition, colIndexMap map[string]int) {
	for i := range conditions {
		if conditions[i].Opcode == SoqlConditionOpcode_FieldInfo {
			switch conditions[i].Value.Type {
			case SoqlFieldInfo_Field:
				if idx, ok := colIndexMap[conditions[i].Value.Key]; ok {
					conditions[i].Value.ColIndex = idx
				}
			case SoqlFieldInfo_Function:
				ctx.applyColIndexToFields(conditions[i].Value.Parameters, colIndexMap)
			}
		}
	}
}

func (ctx *normalizeQueryContext) applyColIndexToOrders(orderBy []SoqlOrderByInfo, colIndexMap map[string]int) {
	for i := range orderBy {
		if idx, ok := colIndexMap[orderBy[i].Field.Key]; ok {
			orderBy[i].Field.ColIndex = idx
		}
	}
}

func (ctx *normalizeQueryContext) applyColIndex(q *SoqlQuery, colIndexMap map[string]int) {

	ctx.applyColIndexToFields(q.Fields, colIndexMap)

	ctx.applyColIndexToOrders(q.OrderBy, colIndexMap)

	ctx.applyColIndexToFields(q.GroupBy, colIndexMap)

	ctx.applyColIndexToConditions(q.Where, colIndexMap)

	ctx.applyColIndexToConditions(q.Having, colIndexMap)
}
