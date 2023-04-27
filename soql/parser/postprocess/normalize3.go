package postprocess

import (
	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
)

func (ctx *normalizeQueryContext) applyColIndexToFields(fields []SoqlFieldInfo) {
	for i := range fields {
		switch fields[i].Type {
		case SoqlFieldInfo_Field:
			if idx, ok := ctx.colIndexMap[fields[i].Key]; ok {
				fields[i].ColIndex = idx
			}
		case SoqlFieldInfo_Function:
			ctx.applyColIndexToFields(fields[i].Parameters)
		}
	}
}

func (ctx *normalizeQueryContext) applyColIndexToConditions(conditions []SoqlCondition) {
	for i := range conditions {
		if conditions[i].Opcode == SoqlConditionOpcode_FieldInfo {
			switch conditions[i].Value.Type {
			case SoqlFieldInfo_Field:
				if idx, ok := ctx.colIndexMap[conditions[i].Value.Key]; ok {
					conditions[i].Value.ColIndex = idx
				}
			case SoqlFieldInfo_Function:
				ctx.applyColIndexToFields(conditions[i].Value.Parameters)
			}
		}
	}
}

func (ctx *normalizeQueryContext) applyColIndexToOrders(orderBy []SoqlOrderByInfo) {
	for i := range orderBy {
		if idx, ok := ctx.colIndexMap[orderBy[i].Field.Key]; ok {
			orderBy[i].Field.ColIndex = idx
		}
	}
}

func (ctx *normalizeQueryContext) applyColIndex(q *SoqlQuery) {

	ctx.applyColIndexToFields(q.Fields)

	ctx.applyColIndexToOrders(q.OrderBy)

	ctx.applyColIndexToFields(q.GroupBy)

	ctx.applyColIndexToConditions(q.Where)

	ctx.applyColIndexToConditions(q.Having)
}
