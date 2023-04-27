package postprocess

import (
	"errors"
	"strings"

	"github.com/shellyln/go-nameutil/nameutil"
	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
)

func (ctx *normalizeQueryContext) applyColIndexToFields(q *SoqlQuery, fields []SoqlFieldInfo) error {
	for i := range fields {
		switch fields[i].Type {
		case SoqlFieldInfo_Field:
			if idx, ok := ctx.colIndexMap[fields[i].Key]; ok {
				fields[i].ColIndex = idx
			} else {
				fields[i].ColIndex = -1
			}

			if len(fields[i].Name) <= len(q.From[0].Name) {
				q.IsCorelated = true

				if fields[i].ColIndex == -1 {
					isSet := false

					if q.Parent != nil {
						ns := nameutil.GetNamespaceFromName(fields[i].Name)
						nsKey := nameutil.MakeDottedKeyIgnoreCase(ns, len(ns))

						for j := range q.Parent.From {
							if q.Parent.From[j].Key == nsKey {
								idx := len(q.Parent.From[j].PerObjectQuery.Fields)

								ctx.colIndexMap[fields[i].Key] = idx
								fields[i].ColIndex = idx

								q.Parent.Fields = append(q.Parent.Fields, fields[i])
								q.Parent.From[j].PerObjectQuery.Fields = append(q.Parent.From[j].PerObjectQuery.Fields, fields[i])

								isSet = true
								break
							}
						}
					}
					if !isSet {
						return errors.New(
							"An incorrect ancestor field of object referred to in the correlated subquery: " +
								strings.Join(fields[i].Name, "."))
					}
				}
			}
		case SoqlFieldInfo_Function:
			ctx.applyColIndexToFields(q, fields[i].Parameters)
		}
	}
	return nil
}

func (ctx *normalizeQueryContext) applyColIndexToConditions(q *SoqlQuery, conditions []SoqlCondition) {
	for i := range conditions {
		if conditions[i].Opcode == SoqlConditionOpcode_FieldInfo {
			switch conditions[i].Value.Type {
			case SoqlFieldInfo_Field:
				if idx, ok := ctx.colIndexMap[conditions[i].Value.Key]; ok {
					conditions[i].Value.ColIndex = idx
				} else {
					conditions[i].Value.ColIndex = -1
				}
			case SoqlFieldInfo_Function:
				ctx.applyColIndexToFields(q, conditions[i].Value.Parameters)
			}
		}
	}
}

func (ctx *normalizeQueryContext) applyColIndexToOrders(orderBy []SoqlOrderByInfo) {
	for i := range orderBy {
		if idx, ok := ctx.colIndexMap[orderBy[i].Field.Key]; ok {
			orderBy[i].Field.ColIndex = idx
		} else {
			orderBy[i].Field.ColIndex = -1
		}
	}
}

func (ctx *normalizeQueryContext) applyColIndex(q *SoqlQuery) {

	ctx.applyColIndexToFields(q, q.Fields)

	ctx.applyColIndexToOrders(q.OrderBy)

	ctx.applyColIndexToFields(q, q.GroupBy)

	ctx.applyColIndexToConditions(q, q.Where)

	ctx.applyColIndexToConditions(q, q.Having)
}
