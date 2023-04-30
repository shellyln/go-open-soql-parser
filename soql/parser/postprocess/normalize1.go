package postprocess

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/shellyln/go-nameutil/nameutil"
	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
)

type normalizeFieldNameConf struct {
	isSelectClause      bool
	isWhereClause       bool
	isHavingClause      bool
	isFunctionParameter bool
	// TODO: shouldBeScalar bool // functhin should be scalar
	allowUnregisteredObject bool
}

func preScanFunctionFields(field *SoqlFieldInfo, q *SoqlQuery) {
	switch field.Type {
	case SoqlFieldInfo_Function:
		{
			funcName := ""
			if len(field.Name) == 1 {
				funcName = strings.ToLower(field.Name[0])
			}

			switch funcName {
			case "count":
				q.IsAggregation = true
			}

			for i := 0; i < len(field.Parameters); i++ {
				preScanFunctionFields(&field.Parameters[i], q)
			}
		}
	}
}

func (ctx *normalizeQueryContext) normalizeFieldName(
	field *SoqlFieldInfo, qPlace soqlQueryPlace, q *SoqlQuery, queryDepth int,
	objNameMap map[string][]string,
	groupingFields map[string]struct{},
	conf normalizeFieldNameConf) error {

	qualifyName := func() {
		if len(field.Name) == 1 {
			// The field name is not qualified by an entity name or alias name.

			s := make([]string, len(q.From[0].Name), len(q.From[0].Name)+1)
			copy(s, q.From[0].Name)
			s = append(s, field.Name...)
			field.Name = s
		}
	}

	switch field.Type {
	case SoqlFieldInfo_Field:
		{
			qualifyName()

			key := nameutil.MakeDottedKeyIgnoreCase(field.Name, len(field.Name)-1)
			fullyQualifiedName, ok := objNameMap[key]

			if !ok {
				// Implicit declaration of relationship name.

				currentName := make([]string, len(field.Name))
				copy(currentName, field.Name)

				key = nameutil.MakeDottedKeyIgnoreCase(currentName, 1)
				if _, ok = objNameMap[key]; !ok {
					// currentName does not start with a known entity name or alias name.

					s := make([]string, 0, len(q.From[0].Name)+len(currentName))
					s = append(s, q.From[0].Name...)
					s = append(s, currentName...)
					currentName = s
				}

				for i := 0; i < len(currentName)-1; i++ {
					s := currentName[:i+1]
					key := nameutil.MakeDottedKeyIgnoreCase(s, i+1)
					if fullyQualifiedName, ok = objNameMap[key]; !ok {
						// Register an unregistered object name.

						if !conf.allowUnregisteredObject {
							return errors.New(
								"Unregistered object names are not allowed: " +
									strings.Join(field.Name, ".") + " at " + strings.Join(q.From[0].Name, "."),
							)
						}

						o := SoqlObjectInfo{
							// TODO: Type (ParentRelationship if conf.isSelectClause otherwise ConditionalOperand)
							Name: s,
							Key:  nameutil.MakeDottedKeyIgnoreCase(s, len(s)),
						}
						objNameMap[key] = s
						q.From = append(q.From, o)
					} else {
						// Replace the alias name with the entity name to normalize it.

						s := make([]string, 0, len(fullyQualifiedName)+len(currentName)-(i+1))
						s = append(s, fullyQualifiedName...)
						s = append(s, currentName[i+1:]...)
						currentName = s
					}
				}

				field.Name = currentName
				key := nameutil.MakeDottedKeyIgnoreCase(field.Name, len(field.Name)-1)
				fullyQualifiedName = objNameMap[key]
			}

			fqnLen := len(fullyQualifiedName)
			s := make([]string, 0, fqnLen+1)
			s = append(s, fullyQualifiedName...)
			s = append(s, field.Name[len(field.Name)-1])
			field.Name = s

			if conf.isSelectClause {
				if !conf.isFunctionParameter && (len(field.Name) <= len(q.From[0].Name)) {
					// TODO: Is `!conf.isFunctionParameter` right condition? // <- BUG: It's a bug!
					return errors.New(
						"The ancestor object item is not allowed to be selected: " +
							strings.Join(field.Name, ".") + " at " + strings.Join(q.From[0].Name, "."),
					)
				}

				for i := 0; i < len(q.From[0].Name); i++ {
					if q.From[0].Name[i] != field.Name[i] {
						return errors.New(
							"The siblings object item is not allowed to be selected: " +
								strings.Join(field.Name, ".") + " at " + strings.Join(q.From[0].Name, "."),
						)
					}
				}
			}

			if field.Name != nil {
				field.Key = nameutil.MakeDottedKeyIgnoreCase(field.Name, len(field.Name))
			}

			if !conf.isFunctionParameter && (conf.isHavingClause || (conf.isSelectClause && q.IsAggregation)) {
				if field.Key != "" {
					if _, ok := groupingFields[field.Key]; !ok {
						if field.AliasName != "" {
							nm := make([]string, 0, len(q.From[0].Name)+1)
							nm = append(nm, q.From[0].Name...)
							nm = append(nm, field.AliasName)
							_, ok = groupingFields[nameutil.MakeDottedKeyIgnoreCase(nm, len(nm))]
						}
						if !ok {
							return errors.New("The item must be included in a Group By clause: " + strings.Join(field.Name, "."))
						}
					}
				}
			}
		}
	case SoqlFieldInfo_Function:
		{
			funcName := ""
			if len(field.Name) == 1 {
				funcName = strings.ToLower(field.Name[0])
			}
			ctx.functions[strings.ToLower(funcName)] = struct{}{}

			if conf.isFunctionParameter {
				// Check function names not allowed in nested
				switch funcName {
				case "fields":
					return errors.New("The function name is not allowed in nested function: " + field.Name[0])
				}
			}
			if !conf.isSelectClause {
				// Check function names not allowed in conditions
				switch funcName {
				case "count", "count_distinct":
					if conf.isHavingClause {
						break
					}
					fallthrough
				case "fields":
					return errors.New("The function name is not allowed in conditional clause: " + field.Name[0])
				}
			}
			if q.IsAggregation {
				switch funcName {
				case "fields":
					return errors.New("The function name is not allowed in aggregation result: " + field.Name[0])
				}
			}

			switch funcName {
			case "fields":
				if len(field.Parameters) != 1 {
					return errors.New("Field set 'Fields()' requires a single parameter")
				}
				if field.Parameters[0].Type != SoqlFieldInfo_Field {
					return errors.New("Field set 'Fields()' parameter must be a name")
				}

				field.Type = SoqlFieldInfo_FieldSet
				field.Name = field.Parameters[0].Name
				field.AliasName = ""
				field.Parameters = nil

				qualifyName()

				key := nameutil.MakeDottedKeyIgnoreCase(field.Name, len(field.Name)-1)
				fullyQualifiedName, ok := objNameMap[key]

				if !ok {
					return errors.New(
						"Field set 'Fields()' parameter refers unknown object: " +
							strings.Join(field.Name, "."))
				}

				fqnLen := len(fullyQualifiedName)
				s := make([]string, 0, fqnLen+1)
				s = append(s, fullyQualifiedName...)
				s = append(s, field.Name[len(field.Name)-1])
				field.Name = s

			case "count":
				switch len(field.Parameters) {
				case 0:
					field.Aggregated = true
					q.IsAggregation = true
				case 1:
					if field.Parameters[0].Type != SoqlFieldInfo_Field {
						return errors.New("Function 'Count()' parameter must be a name")
					}
					field.Aggregated = true
					q.IsAggregation = true
				default:
					return errors.New("Function 'Count()' requires 0 or 1 parameter")
				}

			case "count_distinct":
				switch len(field.Parameters) {
				case 1:
					if field.Parameters[0].Type != SoqlFieldInfo_Field {
						return errors.New("Function 'Count()' parameter must be a name")
					}
					field.Aggregated = true
					q.IsAggregation = true
				default:
					return errors.New("Function 'Count_distinct()' requires 1 parameter")
				}
			}

			for i := 0; i < len(field.Parameters); i++ {
				if err := ctx.normalizeFieldName(
					&field.Parameters[i], qPlace, q, queryDepth,
					objNameMap, groupingFields, normalizeFieldNameConf{
						isSelectClause:          conf.isSelectClause,
						isWhereClause:           conf.isWhereClause,
						isHavingClause:          conf.isHavingClause,
						isFunctionParameter:     true,
						allowUnregisteredObject: conf.allowUnregisteredObject,
					}); err != nil {

					return err
				}
			}

			if conf.isHavingClause || (conf.isSelectClause && q.IsAggregation) {
				// If the parameter contains items and not all parameters are functions, it is an aggregate function.

				aggregated := false
			FN_HAVING_OUTER:
				for i := 0; i < len(field.Parameters); i++ {
					switch field.Parameters[i].Type {
					case SoqlFieldInfo_Field:
						aggregated = true
					case SoqlFieldInfo_Function:
						aggregated = false
						break FN_HAVING_OUTER
					}
				}
				field.Aggregated = aggregated

				for i := 0; i < len(field.Parameters); i++ {
					// TODO: FIXIT: Extract this condition outside the loop.
					if aggregated {
						continue
					}
					param := &field.Parameters[i]
					if param.Type == SoqlFieldInfo_Field && param.Key != "" {
						if _, ok := groupingFields[param.Key]; !ok {
							if param.AliasName != "" {
								nm := make([]string, 0, len(q.From[0].Name)+1)
								nm = append(nm, q.From[0].Name...)
								nm = append(nm, param.AliasName)
								_, ok = groupingFields[nameutil.MakeDottedKeyIgnoreCase(nm, len(nm))]
							}
							if !ok {
								return errors.New("The item must be included in a Group By clause: " + strings.Join(param.Name, "."))
							}
						}
					}
				}
			}
		}
	case SoqlFieldInfo_SubQuery:
		if conf.isSelectClause {
			field.SubQuery.Parent = q
		}
		if err := ctx.normalizeQuery(qPlace, field.SubQuery, q, queryDepth+1, objNameMap); err != nil {
			return err
		}
	case SoqlFieldInfo_FieldSet:
		if !conf.isSelectClause || conf.isFunctionParameter {
			return errors.New(
				"The fields() is not allowed in conditional clause or function parameter: " +
					strings.Join(field.Name, "."))
		}
	case SoqlFieldInfo_ParameterizedValue:
		ctx.parameters[strings.ToLower(field.Name[0])] = struct{}{}
	case SoqlFieldInfo_DateTimeLiteralName:
		ctx.dateTimeLiterals[strings.ToLower(field.Name[0])] = struct{}{}
	}

	return nil
}

func (ctx *normalizeQueryContext) normalizeObjectName(
	object *SoqlObjectInfo, q *SoqlQuery, objNameMap map[string][]string) error {

	currentName := make([]string, len(object.Name))
	copy(currentName, object.Name)

	{
		key := nameutil.MakeDottedKeyIgnoreCase(currentName, 1)
		if fullyQualifiedName, ok := objNameMap[key]; ok {
			// Replace the alias name with the entity name to normalize it.

			fqnLen := len(fullyQualifiedName)
			s := make([]string, 0, fqnLen+len(currentName))
			s = append(s, fullyQualifiedName...)
			s = append(s, currentName[1:]...)
			currentName = s
		}
	}

	for i := 0; i < len(currentName)-1; i++ {
		s := currentName[:i+1]
		key := nameutil.MakeDottedKeyIgnoreCase(s, i+1)
		if fullyQualifiedName, ok := objNameMap[key]; !ok {
			o := SoqlObjectInfo{
				// TODO: Type // ParentRelationship
				Name: s,
				Key:  key,
			}
			objNameMap[key] = s
			q.From = append(q.From, o)
		} else {
			// Replace the alias name with the entity name to normalize it.

			s := make([]string, 0, len(fullyQualifiedName)+len(currentName)-(i+1))
			s = append(s, fullyQualifiedName...)
			s = append(s, currentName[i+1:]...)
			currentName = s
		}
	}

	{
		key := nameutil.MakeDottedKeyIgnoreCase(currentName, len(currentName))
		objNameMap[key] = currentName
		if len(object.AliasName) > 0 {
			// TODO: Check duplicated name
			objNameMap[base64.StdEncoding.EncodeToString([]byte(strings.ToLower(object.AliasName)))] = currentName

			s := make([]string, 0, len(currentName))
			s = append(s, currentName...)
			s[len(currentName)-1] = object.AliasName

			aliasKey := nameutil.MakeDottedKeyIgnoreCase(s, len(s))
			objNameMap[aliasKey] = currentName
		}

		object.Name = currentName
		object.Key = key
	}

	return nil
}
