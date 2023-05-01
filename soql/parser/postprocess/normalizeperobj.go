package postprocess

import (
	"errors"
	"sort"
	"strconv"

	"github.com/shellyln/go-nameutil/nameutil"
	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
)

func makePerObjectFields(perObjQuery *SoqlQuery, srcFields []SoqlFieldInfo) []SoqlFieldInfo {
	targetFields := make([]SoqlFieldInfo, 0, len(srcFields))

	for i := 0; i < len(srcFields); i++ {
		switch srcFields[i].Type {
		case SoqlFieldInfo_Field, SoqlFieldInfo_FieldSet:
			{
				name := srcFields[i].Name
				if nameutil.NameEqualsIgnoreCase(perObjQuery.From[0].Name, name[:len(name)-1]) {
					targetFields = append(targetFields, srcFields[i])
				}
			}
		case SoqlFieldInfo_Function, SoqlFieldInfo_SubQuery:
			// NOTE: do nothing
			//       Because the function is executed locally, not in the adapter.
			//       The subqueries are joined in a post-process.
		}
	}

	// TODO: find the primary key field and append if it is not exist

	if len(targetFields) > 0 {
		return targetFields
	} else {
		return nil
	}
}

func makePerObjectOrders(perObjQuery *SoqlQuery, srcOrders []SoqlOrderByInfo) []SoqlOrderByInfo {
	targetOrders := make([]SoqlOrderByInfo, 0, len(srcOrders))

	for i := 0; i < len(srcOrders); i++ {
		switch srcOrders[i].Field.Type {
		case SoqlFieldInfo_Field:
			{
				name := srcOrders[i].Field.Name
				if nameutil.NameEqualsIgnoreCase(perObjQuery.From[0].Name, name[:len(name)-1]) {
					targetOrders = append(targetOrders, srcOrders[i])
				}
			}
		}
	}

	if len(targetOrders) > 0 {
		return targetOrders
	} else {
		return nil
	}
}

func makePerObjectConditions(perObjQuery *SoqlQuery, srcCondtions []SoqlCondition) (bool, []SoqlCondition, bool) {
	conditions := make([]SoqlCondition, 0, len(srcCondtions))
	conditions = append(conditions, srcCondtions...)

	stack := make([]soqlConditionStack, len(conditions), len(conditions))
	sp := 0

	objectName := perObjQuery.From[0].Name

	hasConditionsOriginally := false
	// If there are no fields originally, then inner join
	innerJoined := false

CHECK_FIELDS:
	for i := 0; i < len(conditions); i++ {
		cond := conditions[i]

		switch cond.Opcode {
		case SoqlConditionOpcode_FieldInfo:
			switch cond.Value.Type {
			case SoqlFieldInfo_Field:
				if nameutil.NameEqualsIgnoreCase(objectName, cond.Value.Name[:len(cond.Value.Name)-1]) {
					hasConditionsOriginally = true
					break CHECK_FIELDS
				}
			}
		}
	}

	for i := 0; i < len(conditions); i++ {
		cond := conditions[i]

		switch cond.Opcode {
		case SoqlConditionOpcode_Noop:
			// do nothing
		case SoqlConditionOpcode_Unknown:
			{
				stack[sp].cond = cond
				stack[sp].srcPos = i
				stack[sp].minPos = i
				sp++
			}
		case SoqlConditionOpcode_FieldInfo:
			{
				stack[sp].cond = cond
				stack[sp].srcPos = i
				stack[sp].minPos = i

				switch cond.Value.Type {
				case SoqlFieldInfo_Field:
					if !nameutil.NameEqualsIgnoreCase(objectName, cond.Value.Name[:len(cond.Value.Name)-1]) {
						conditions[i].Opcode = SoqlConditionOpcode_Unknown
						stack[sp].cond.Opcode = SoqlConditionOpcode_Unknown
					}
				case SoqlFieldInfo_Function, SoqlFieldInfo_SubQuery:
					// NOTE: Function always replace with Unknown.
					//       Because the function is executed locally, not in the adapter.
					conditions[i].Opcode = SoqlConditionOpcode_Unknown
					stack[sp].cond.Opcode = SoqlConditionOpcode_Unknown
				}
				sp++
			}
		case SoqlConditionOpcode_Not:
			{
				op1 := stack[sp-1]
				sp -= 1

				stack[sp].cond = cond
				stack[sp].srcPos = i
				stack[sp].minPos = op1.minPos

				if op1.cond.Opcode == SoqlConditionOpcode_Unknown {
					// Retain one Unknown on the stack.
					for j := op1.minPos; j <= op1.srcPos; j++ {
						conditions[j].Opcode = SoqlConditionOpcode_Noop
					}
					conditions[i].Opcode = SoqlConditionOpcode_Unknown
					stack[sp].cond.Opcode = SoqlConditionOpcode_Unknown
				}
				sp++
			}
		case SoqlConditionOpcode_And:
			{
				op1 := stack[sp-2]
				op2 := stack[sp-1]
				sp -= 2

				stack[sp].cond = cond
				stack[sp].srcPos = i

				maxPos := -1
				if op1.srcPos > op2.srcPos {
					maxPos = op1.srcPos
				} else {
					maxPos = op2.srcPos
				}

				if op1.minPos < op2.minPos {
					stack[sp].minPos = op1.minPos
				} else {
					stack[sp].minPos = op2.minPos
				}

				op1IsUnknown := op1.cond.Opcode == SoqlConditionOpcode_Unknown
				op2IsUnknown := op2.cond.Opcode == SoqlConditionOpcode_Unknown

				if op1IsUnknown && op2IsUnknown {
					// Retain one Unknown on the stack.
					for j := stack[sp].minPos; j <= maxPos; j++ {
						conditions[j].Opcode = SoqlConditionOpcode_Noop
					}
					conditions[i].Opcode = SoqlConditionOpcode_Unknown
					stack[sp].cond.Opcode = SoqlConditionOpcode_Unknown

				} else if op1IsUnknown || op2IsUnknown {
					// Leave one side of the path. To do this, replace the top with Noop.
					if op1IsUnknown {
						for j := op1.minPos; j <= op1.srcPos; j++ {
							conditions[j].Opcode = SoqlConditionOpcode_Noop
						}
					} else if op2IsUnknown {
						for j := op2.minPos; j <= op2.srcPos; j++ {
							conditions[j].Opcode = SoqlConditionOpcode_Noop
						}
					}
					conditions[i].Opcode = SoqlConditionOpcode_Noop
				}

				sp++
			}
		default:
			{
				op1 := stack[sp-2]
				op2 := stack[sp-1]
				sp -= 2

				stack[sp].cond = cond
				stack[sp].srcPos = i

				maxPos := -1
				if op1.srcPos > op2.srcPos {
					maxPos = op1.srcPos
				} else {
					maxPos = op2.srcPos
				}

				if op1.minPos < op2.minPos {
					stack[sp].minPos = op1.minPos
				} else {
					stack[sp].minPos = op2.minPos
				}

				if op1.cond.Opcode == SoqlConditionOpcode_Unknown || op2.cond.Opcode == SoqlConditionOpcode_Unknown {
					// Retain one Unknown on the stack.
					for j := stack[sp].minPos; j <= maxPos; j++ {
						conditions[j].Opcode = SoqlConditionOpcode_Noop
					}
					conditions[i].Opcode = SoqlConditionOpcode_Unknown
					stack[sp].cond.Opcode = SoqlConditionOpcode_Unknown
				}

				sp++
			}
		}
	}

	if sp > 0 && stack[sp-1].cond.Opcode == SoqlConditionOpcode_Unknown {
		for i := 0; i < len(conditions); i++ {
			conditions[i].Opcode = SoqlConditionOpcode_Noop
		}
	}

	// If the conditions survive, then inner join
	for i := 0; i < len(conditions); i++ {
		if conditions[i].Opcode != SoqlConditionOpcode_Noop {
			innerJoined = true
			break
		}
	}

	return hasConditionsOriginally, conditions, innerJoined
}

func makePostProcessWhereClause(q *SoqlQuery) error {
	postProcessWhere := make([]SoqlCondition, 0, len(q.Where))
	postProcessWhere = append(postProcessWhere, q.Where...)

	for i := 0; i < len(q.From); i++ {
		for j := 0; j < len(q.From[i].PerObjectQuery.Where); j++ {
			switch q.From[i].PerObjectQuery.Where[j].Opcode {
			case SoqlConditionOpcode_Noop, SoqlConditionOpcode_Unknown:
				// do nothing
			default:
				// The object is inner-joined.
				postProcessWhere[j].Opcode = SoqlConditionOpcode_Noop
			}
		}
	}

	for sp, i := 0, 0; i < len(postProcessWhere); i++ {
		switch postProcessWhere[i].Opcode {
		case SoqlConditionOpcode_Noop:
			// do nothing
		case SoqlConditionOpcode_Unknown, SoqlConditionOpcode_FieldInfo:
			sp++
		case SoqlConditionOpcode_Not:
			if sp == 0 {
				postProcessWhere[i].Opcode = SoqlConditionOpcode_Noop
			}
		case SoqlConditionOpcode_And:
			if sp < 2 {
				// The logical operation between two inner-joined objects is always true.
				// because it is filtered by each object in advance and
				// combinations that are false are eliminated by table joining.
				postProcessWhere[i].Opcode = SoqlConditionOpcode_Noop
			} else {
				sp--
			}
		default:
			if sp < 2 {
				return errors.New(
					"[FATAL]Internal error: " +
						"The operand for the binary operator is missing " +
						"in the Where clause of the post-process filter: " +
						"n=" + strconv.Itoa(sp),
				)
			} else {
				sp--
			}
		}
	}

	q.PostProcessWhere = postProcessWhere
	return nil
}

func compactConditions(conditions []SoqlCondition) []SoqlCondition {
	sort.SliceStable(conditions, func(i, j int) bool {
		// descending: true if i is larger
		if conditions[j].Opcode == SoqlConditionOpcode_Noop {
			return conditions[i].Opcode != SoqlConditionOpcode_Noop
		}
		return false
	})

	for i := 0; i < len(conditions); i++ {
		if conditions[i].Opcode == SoqlConditionOpcode_Noop {
			conditions = conditions[:i]
			break
		}
	}

	if len(conditions) == 0 {
		return nil
	} else {
		return conditions
	}
}

func (ctx *normalizeQueryContext) buildPerObjectInfo(q *SoqlQuery) error {
	for i := 0; i < len(q.From); i++ {
		perObjQuery := &SoqlQuery{
			From: []SoqlObjectInfo{q.From[i]},
		}
		q.From[i].PerObjectQuery = perObjQuery

		perObjQuery.Fields = makePerObjectFields(perObjQuery, q.Fields)

		for j := range perObjQuery.Fields {
			perObjQuery.Fields[j].ColIndex = j
			ctx.colIndexMap[perObjQuery.Fields[j].Key] = j
		}
	}

	if err := ctx.applyColIndex(q); err != nil {
		return err
	}

	for i := 0; i < len(q.From); i++ {
		perObjQuery := q.From[i].PerObjectQuery

		if len(q.Where) != 0 {
			q.From[i].HasConditions, perObjQuery.Where, q.From[i].InnerJoin =
				makePerObjectConditions(perObjQuery, q.Where)
		}

		// NOTE: GroupBy and Having are not processed.
		//       Aggregation is not done for multiple objects.

		if q.OrderBy != nil {
			perObjQuery.OrderBy = makePerObjectOrders(perObjQuery, q.OrderBy)
		}

		if i == 0 {
			perObjQuery.OffsetAndLimit = q.OffsetAndLimit
		}

		perObjQuery.For = q.For
	}
	q.From[0].InnerJoin = false

	if err := makePostProcessWhereClause(q); err != nil {
		return err
	}

	for i := 0; i < len(q.From); i++ {
		if q.From[i].PerObjectQuery.Where != nil {
			q.From[i].PerObjectQuery.Where = compactConditions(q.From[i].PerObjectQuery.Where)
		}
	}
	q.PostProcessWhere = compactConditions(q.PostProcessWhere)

	return nil
}
