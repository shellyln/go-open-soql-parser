package postprocess

import (
	"errors"
	"sort"
	"strings"

	"github.com/shellyln/go-nameutil/nameutil"
	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
)

type normalizeQueryContext struct {
	queryId            int
	viewId             int
	columnId           int
	viewIdMap          map[string]int
	columnIdMap        map[string]int
	colIndexMap        map[string]int
	headObjDepthOffset int
	maxQueryDepth      int
	maxViewDepth       int
	queryGraph         map[int]SoqlQueryGraphLeaf
	viewGraph          map[int]SoqlViewGraphLeaf
	functions          map[string]struct{}
	parameters         map[string]struct{}
	dateTimeLiterals   map[string]struct{}
}

func (ctx *normalizeQueryContext) normalizeQuery(
	qPlace soqlQueryPlace, q, callParentQuery *SoqlQuery, queryDepth int, objNameMap map[string][]string) error {

	q.QueryId = ctx.queryId
	ctx.queryId++

	if queryDepth > ctx.maxQueryDepth {
		ctx.maxQueryDepth = queryDepth
	}

	var parentQueryId int
	if callParentQuery != nil {
		parentQueryId = callParentQuery.QueryId
	}
	ctx.queryGraph[q.QueryId] = SoqlQueryGraphLeaf{
		ParentQueryId: parentQueryId,
		Depth:         queryDepth,
		IsConditional: qPlace == soqlQueryPlace_ConditionalOperand,
		Query:         q,
	}

	var primaryObjectName []string // TODO: name is not good? it is primary or parent object

	if qPlace == soqlQueryPlace_Primary || qPlace == soqlQueryPlace_ConditionalOperand {
		primaryObjectName = q.From[0].Name
		if len(primaryObjectName) != 1 {
			return errors.New(
				"The name of the primary object is qualified by the parent object name: " +
					strings.Join(primaryObjectName, "."),
			)
		}

		objNameMap = make(map[string][]string) // dotted name (include alias) -> fully qualified name
		// BUG: ^^^^^ Case of soqlQueryPlace_ConditionalOperand may be derived (duplicate) the objNameMap.
	} else {
		primaryObjectName = q.Parent.From[0].Name

		m := make(map[string][]string)
		for k, v := range objNameMap {
			m[k] = v
		}
		objNameMap = m
	}

	objectAliasMap := make(map[string]*SoqlObjectInfo)

	for i := 0; i < len(q.From); i++ {
		object := q.From[i]

		if i != 0 || !(qPlace == soqlQueryPlace_Primary || qPlace == soqlQueryPlace_ConditionalOperand) {
			key := nameutil.MakeDottedKeyIgnoreCase(object.Name, 1)
			if _, ok := objNameMap[key]; !ok {
				// The object name is not qualified by another entity name or alias name.

				s := make([]string, 0, len(primaryObjectName)+len(object.Name))
				s = append(s, primaryObjectName...)
				s = append(s, object.Name...)
				object.Name = s
			}
		}

		if err := ctx.normalizeObjectName(&object, q, objNameMap); err != nil {
			return err
		}
		q.From[i] = object

		if object.AliasName != "" {
			aliasName := strings.ToLower(object.AliasName)

			if _, ok := objectAliasMap[aliasName]; !ok {
				objectAliasMap[aliasName] = &q.From[i]
			} else {
				return errors.New("Duplicate object alias name found: " + object.AliasName)
			}
		}
	}

	groupingFields := make(map[string]struct{})

	if q.GroupBy != nil {
		for i := 0; i < len(q.GroupBy); i++ {
			field := q.GroupBy[i]

			if err := ctx.normalizeFieldName(
				&field, soqlQueryPlace_ConditionalOperand, q, queryDepth,
				objNameMap, nil, normalizeFieldNameConf{
					isSelectClause:          false,
					isWhereClause:           false,
					isHavingClause:          false,
					isFunctionParameter:     false,
					allowUnregisteredObject: true,
				}); err != nil {

				return err
			}
			q.GroupBy[i] = field
			groupingFields[field.Key] = struct{}{}
		}
	}

	fieldAliasMap := make(map[string]*SoqlFieldInfo)

	for i := 0; i < len(q.Fields); i++ {
		field := q.Fields[i]

		if field.Type != SoqlFieldInfo_SubQuery {
			if err := ctx.normalizeFieldName(
				&field, soqlQueryPlace_Select, q, queryDepth,
				objNameMap, groupingFields, normalizeFieldNameConf{
					isSelectClause:          true,
					isWhereClause:           false,
					isHavingClause:          false,
					isFunctionParameter:     false,
					allowUnregisteredObject: true,
				}); err != nil {

				return err
			}
			q.Fields[i] = field
		}

		if field.AliasName != "" {
			aliasName := strings.ToLower(field.AliasName)

			if _, ok := fieldAliasMap[aliasName]; !ok {
				fieldAliasMap[aliasName] = &q.Fields[i]
			} else {
				return errors.New("Duplicate field alias name found: " + field.AliasName)
			}
		}
	}

	if q.GroupBy != nil {
		for i := 0; i < len(q.GroupBy); i++ {
			field := q.GroupBy[i]

			if len(field.Name) == len(q.From[0].Name)+1 {
				if p, ok := fieldAliasMap[field.Name[len(field.Name)-1]]; ok {
					delete(groupingFields, field.Key)
					q.GroupBy[i] = *p
					groupingFields[p.Key] = struct{}{}
				}
			}
		}
	}

	if q.Where != nil {
		// NOTE: Distribute the not operator to each inner operators.
		//       If the condition is reversed by not,
		//       records that have not been retrieved by the data adapter will be targeted for retrieval.

		q.Where = distributeNotOperators(q.Where)

		for i := 0; i < len(q.Where); i++ {
			condition := q.Where[i]
			if condition.Opcode == SoqlConditionOpcode_FieldInfo && condition.Value.Type != SoqlFieldInfo_SubQuery {
				if err := ctx.normalizeFieldName(
					&condition.Value, soqlQueryPlace_ConditionalOperand, q, queryDepth,
					objNameMap, nil, normalizeFieldNameConf{
						isSelectClause:          false,
						isWhereClause:           true,
						isHavingClause:          false,
						isFunctionParameter:     false,
						allowUnregisteredObject: true,
					}); err != nil {

					return err
				}
				q.Where[i] = condition
			}
		}
	}

	if q.Having != nil {
		if q.GroupBy == nil {
			return errors.New("Group by clause not found: " + strings.Join(primaryObjectName, "."))
		}

		q.Having = distributeNotOperators(q.Having)

		for i := 0; i < len(q.Having); i++ {
			condition := q.Having[i]
			if condition.Opcode == SoqlConditionOpcode_FieldInfo && condition.Value.Type != SoqlFieldInfo_SubQuery {
				if err := ctx.normalizeFieldName(
					&condition.Value, soqlQueryPlace_ConditionalOperand, q, queryDepth,
					objNameMap, groupingFields, normalizeFieldNameConf{
						isSelectClause:          false,
						isWhereClause:           false,
						isHavingClause:          true,
						isFunctionParameter:     false,
						allowUnregisteredObject: true,
					}); err != nil {

					return err
				}
				q.Having[i] = condition
			}
		}
	}

	for i := 0; i < len(q.Fields); i++ {
		preScanFunctionFields(&q.Fields[i], q)
	}

	if q.OrderBy != nil {
		for i := 0; i < len(q.OrderBy); i++ {
			field := q.OrderBy[i].Field

			aliasFound := false
			if len(field.Name) == 1 {
				if p, ok := fieldAliasMap[strings.ToLower(field.Name[0])]; ok {
					field = *p
					aliasFound = true
				}
			}

			if !aliasFound {
				if err := ctx.normalizeFieldName(
					&field, soqlQueryPlace_ConditionalOperand, q, queryDepth,
					objNameMap, nil, normalizeFieldNameConf{
						isSelectClause:          false,
						isWhereClause:           false,
						isHavingClause:          false,
						isFunctionParameter:     false,
						allowUnregisteredObject: false,
					}); err != nil {

					return err
				}
			}
			q.OrderBy[i].Field = field
		}
	}

	sort.SliceStable(q.From[1:], func(i, j int) bool {
		return len(q.From[i+1].Name) < len(q.From[j+1].Name)
	})

	if err := ctx.addUnselectedFields(q); err != nil {
		return err
	}

	for i := 0; i < len(q.From); i++ {
		if viewId, ok := ctx.viewIdMap[q.From[i].Key]; !ok {
			q.From[i].ViewId = ctx.viewId
			ctx.viewIdMap[q.From[i].Key] = ctx.viewId
			ctx.viewId++
		} else {
			q.From[i].ViewId = viewId
		}

		nameLen := len(q.From[i].Name)
		objDepth := nameLen + ctx.headObjDepthOffset
		if ctx.maxViewDepth < objDepth {
			ctx.maxViewDepth = objDepth
		}

		if nameLen > 1 {
			parentKey := nameutil.MakeDottedKeyIgnoreCase(q.From[i].Name, nameLen-1)
			if parentViewId, ok := ctx.viewIdMap[parentKey]; ok {
				q.From[i].ParentViewId = parentViewId
			}
		}

		ctx.viewGraph[q.From[i].ViewId] = SoqlViewGraphLeaf{
			ParentViewId: q.From[i].ParentViewId,
			QueryId:      q.QueryId,
			Depth:        objDepth,
			QueryDepth:   queryDepth,
			Object:       &q.From[i],
			Query:        q,
		}
	}

	// TODO: * check object graph when aggregation(group by)
	//           * subquery on select clause is not allowed.
	//           * ...

	ctx.assignColumnIds(q)

	ctx.assignImplicitAliasNames(q, fieldAliasMap)

	{
		usedColumnIds := make(map[int]struct{})
		for i := 0; i < len(q.GroupBy); i++ {
			if _, ok := usedColumnIds[q.GroupBy[i].ColumnId]; ok {
				return errors.New(
					"Duplicate field found in Group by clause: " +
						strings.Join(q.GroupBy[i].Name, "."))
			}
			usedColumnIds[q.GroupBy[i].ColumnId] = struct{}{}
		}
	}
	{
		usedColumnIds := make(map[int]struct{})
		for i := 0; i < len(q.OrderBy); i++ {
			if _, ok := usedColumnIds[q.OrderBy[i].Field.ColumnId]; ok {
				return errors.New(
					"Duplicate field found in Order by clause: " +
						strings.Join(q.OrderBy[i].Field.Name, "."))
			}
			usedColumnIds[q.OrderBy[i].Field.ColumnId] = struct{}{}
		}
	}

	if err := ctx.buildPerObjectInfo(q); err != nil {
		return err
	}

	for i := 0; i < len(q.Fields); i++ {
		field := q.Fields[i]

		if field.Type == SoqlFieldInfo_SubQuery {
			if err := ctx.normalizeFieldName(
				&field, soqlQueryPlace_Select, q, queryDepth,
				objNameMap, groupingFields, normalizeFieldNameConf{
					isSelectClause:          true,
					isWhereClause:           false,
					isHavingClause:          false,
					isFunctionParameter:     false,
					allowUnregisteredObject: true,
				}); err != nil {

				return err
			}
			q.Fields[i] = field
		}
	}

	if q.OffsetAndLimit.OffsetParamName != "" {
		ctx.parameters[q.OffsetAndLimit.OffsetParamName] = struct{}{}
	}
	if q.OffsetAndLimit.LimitParamName != "" {
		ctx.parameters[q.OffsetAndLimit.LimitParamName] = struct{}{}
	}

	savedViewIdMap := ctx.viewIdMap
	savedColumnIdMap := ctx.columnIdMap
	savedColIndexMap := ctx.colIndexMap
	savedHeadObjDepthOffset := ctx.headObjDepthOffset
	ctx.viewIdMap = make(map[string]int)
	ctx.columnIdMap = make(map[string]int)
	ctx.colIndexMap = make(map[string]int)
	ctx.headObjDepthOffset = len(q.From[0].Name)

	if q.Where != nil {
		for i := 0; i < len(q.Where); i++ {
			condition := q.Where[i]
			if condition.Opcode == SoqlConditionOpcode_FieldInfo && condition.Value.Type == SoqlFieldInfo_SubQuery {
				if err := ctx.normalizeFieldName(
					&condition.Value, soqlQueryPlace_ConditionalOperand, q, queryDepth,
					objNameMap, nil, normalizeFieldNameConf{
						isSelectClause:          false,
						isWhereClause:           true,
						isHavingClause:          false,
						isFunctionParameter:     false,
						allowUnregisteredObject: true,
					}); err != nil {

					return err
				}
				q.Where[i] = condition
			}
		}
	}

	if q.Having != nil {
		for i := 0; i < len(q.Having); i++ {
			condition := q.Having[i]
			if condition.Opcode == SoqlConditionOpcode_FieldInfo && condition.Value.Type == SoqlFieldInfo_SubQuery {
				if err := ctx.normalizeFieldName(
					&condition.Value, soqlQueryPlace_ConditionalOperand, q, queryDepth,
					objNameMap, groupingFields, normalizeFieldNameConf{
						isSelectClause:          false,
						isWhereClause:           false,
						isHavingClause:          true,
						isFunctionParameter:     false,
						allowUnregisteredObject: true,
					}); err != nil {

					return err
				}
				q.Having[i] = condition
			}
		}
	}

	ctx.viewIdMap = savedViewIdMap
	ctx.columnIdMap = savedColumnIdMap
	ctx.colIndexMap = savedColIndexMap
	ctx.headObjDepthOffset = savedHeadObjDepthOffset

	return nil
}

func Normalize(q *SoqlQuery) error {
	ctx := normalizeQueryContext{
		queryId:            1,
		viewId:             1,
		columnId:           1,
		viewIdMap:          make(map[string]int),
		columnIdMap:        make(map[string]int),
		colIndexMap:        map[string]int{},
		headObjDepthOffset: 0,
		maxQueryDepth:      0,
		maxViewDepth:       0,
		queryGraph:         make(map[int]SoqlQueryGraphLeaf),
		viewGraph:          make(map[int]SoqlViewGraphLeaf),
		functions:          make(map[string]struct{}),
		parameters:         make(map[string]struct{}),
		dateTimeLiterals:   make(map[string]struct{}),
	}

	if err := ctx.normalizeQuery(soqlQueryPlace_Primary, q, nil, 1, nil); err != nil {
		return err
	}

	q.Meta.NextQueryId = ctx.queryId
	q.Meta.NextViewId = ctx.viewId
	q.Meta.NextColumnId = ctx.columnId
	q.Meta.MaxQueryDepth = ctx.maxQueryDepth
	q.Meta.MaxViewDepth = ctx.maxViewDepth
	q.Meta.QueryGraph = ctx.queryGraph
	q.Meta.ViewGraph = ctx.viewGraph
	q.Meta.Functions = ctx.functions
	q.Meta.Parameters = ctx.parameters
	q.Meta.DateTimeLiterals = ctx.dateTimeLiterals

	return nil
}
