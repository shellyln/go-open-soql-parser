package core

import (
	"errors"

	"github.com/shellyln/go-open-soql-parser/soql/parser/core/class"
	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
	. "github.com/shellyln/takenoco/base"
	. "github.com/shellyln/takenoco/string"
)

func fieldInfoTypeFromClassName(className string) SoqlFieldInfoType {
	var ty SoqlFieldInfoType
	// TODO: time value
	switch className {
	case class.Null:
		ty = SoqlFieldInfo_Literal_Null
	case class.Int:
		ty = SoqlFieldInfo_Literal_Int
	case class.Float:
		ty = SoqlFieldInfo_Literal_Float
	case class.Bool:
		ty = SoqlFieldInfo_Literal_Bool
	case class.String:
		ty = SoqlFieldInfo_Literal_String
	case class.Blob:
		ty = SoqlFieldInfo_Literal_Blob
	case class.Date:
		ty = SoqlFieldInfo_Literal_Date
	case class.DateTime:
		ty = SoqlFieldInfo_Literal_DateTime
	case class.Time:
		ty = SoqlFieldInfo_Literal_Time
	case class.List:
		ty = SoqlFieldInfo_Literal_List
	case class.ParameterizedValue:
		ty = SoqlFieldInfo_ParameterizedValue
	case class.DateTimeLiteralName:
		ty = SoqlFieldInfo_DateTimeLiteralName
	default:
		ty = -1
	}
	return ty
}

func selectFieldFunctionCall() ParserFn {
	return Trans(
		FlatGroup(
			symbolName(),
			sp0(),
			erase(CharClass("(")),
			First(
				FlatGroup(
					sp0(),
					ZeroOrOnce(
						First(
							Indirect(selectFieldFunctionCall),
							complexSymbolName(),
							literalValue(),
						),
						ZeroOrMoreTimes(
							sp0(),
							erase(CharClass(",")),
							sp0(),
							First(
								Indirect(selectFieldFunctionCall),
								complexSymbolName(),
								literalValue(),
							),
						),
						sp0(),
					),
					erase(CharClass(")")),
					sp0(),
				),
				Error("')' is expected"),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			astsLen := len(asts)
			z := make([]SoqlFieldInfo, astsLen-1, astsLen-1)
			for i := 1; i < astsLen; i++ {
				switch asts[i].ClassName {
				case class.SelectFieldFunctionCall:
					z[i-1] = asts[i].Value.(SoqlFieldInfo)
				case class.ComplexSymbol:
					z[i-1] = SoqlFieldInfo{
						Type: SoqlFieldInfo_Field,
						Name: asts[i].Value.([]string),
					}
				default:
					{
						ty := fieldInfoTypeFromClassName(asts[i].ClassName)
						if ty < 0 {
							return nil, errors.New("Unexpected token at field function expression: " + asts[0].ClassName)
						}

						switch ty {
						case SoqlFieldInfo_ParameterizedValue, SoqlFieldInfo_DateTimeLiteralName:
							z[i-1] = SoqlFieldInfo{
								Type:      ty,
								ClassName: asts[i].ClassName,
								Name:      []string{asts[i].Value.(string)},
							}
						default:
							z[i-1] = SoqlFieldInfo{
								Type:      ty,
								ClassName: asts[i].ClassName,
								Value:     asts[i].Value,
							}
						}
					}
				}
			}
			return AstSlice{{
				ClassName: class.SelectFieldFunctionCall,
				Type:      AstType_Any,
				Value: SoqlFieldInfo{
					Type:       SoqlFieldInfo_Function,
					Name:       []string{asts[0].Value.(string)},
					Parameters: z,
				},
			}}, nil
		},
	)
}

func subQuery() ParserFn {
	return Trans(
		FlatGroup(
			erase(CharClass("(")),
			sp0(),
			Indirect(subQuerySelectStatement),
			sp0(),
			erase(CharClass(")")),
			sp0(),
		),
		ChangeClassName(class.SubQuery),
	)
}

func transComplexSelectFieldName(ctx ParserContext, asts AstSlice) (AstSlice, error) {
	field := SoqlFieldInfo{
		AliasName: asts[1].Value.(string),
	}

	switch asts[0].ClassName {
	case class.SelectFieldFunctionCall:
		{
			field = asts[0].Value.(SoqlFieldInfo)
			field.AliasName = asts[1].Value.(string)
		}
	case class.SubQuery:
		{
			field.Type = SoqlFieldInfo_SubQuery
			subQuery := asts[0].Value.(SoqlQuery)
			field.SubQuery = &subQuery
		}
	case class.ComplexSymbol:
		{
			field.Type = SoqlFieldInfo_Field
			field.Name = asts[0].Value.([]string)
		}
	case class.List:
		{
			field.Type = SoqlFieldInfo_Literal_List
			field.ClassName = asts[0].ClassName
			field.Value = asts[0].Value
		}
	default:
		{
			ty := fieldInfoTypeFromClassName(asts[0].ClassName)
			if ty < 0 {
				return nil, errors.New("Unexpected token at field expression: " + asts[0].ClassName)
			}

			field.Type = ty
			field.ClassName = asts[0].ClassName

			switch ty {
			case SoqlFieldInfo_ParameterizedValue:
				field.Name = []string{asts[0].Value.(string)}
			case SoqlFieldInfo_DateTimeLiteralName:
				{
					v := asts[0].Value.(SoqlDateTimeLiteralName)
					field.Name = []string{v.Name}
					field.Value = v
				}
			default:
				field.Value = asts[0].Value
			}
		}
	}

	return AstSlice{{
		ClassName: "soql:FieldInfo",
		Type:      AstType_Any,
		Value:     field,
	}}, nil
}

func complexSelectFieldName() ParserFn {
	return Trans(
		FlatGroup(
			notAheadReservedKeywords(),
			First(
				selectFieldFunctionCall(), // SoqlFieldInfo
				complexSymbolName(),       // []string
				subQuery(),                // SoqlQuery
			),
			First(
				// Alias name
				FlatGroup(
					notAheadReservedKeywords(),
					symbolName(),
				),
				Zero(Ast{Value: ""}),
			),
			// TODO: hinting string
		),
		transComplexSelectFieldName,
	)
}

func selectFieldList() ParserFn {
	return Trans(
		FlatGroup(
			First(
				FlatGroup(
					complexSelectFieldName(),
					sp0(),
				),
				Error("Unexpected token aheads near by the select clause (field list)"),
			),
			ZeroOrMoreTimes(
				erase(CharClass(",")),
				First(
					FlatGroup(
						sp0(),
						complexSelectFieldName(),
						sp0(),
					),
					Error("Unexpected token aheads near by the select clause (field list)"),
				),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			astsLen := len(asts)
			z := make([]SoqlFieldInfo, astsLen, astsLen)
			for i := 0; i < astsLen; i++ {
				z[i] = asts[i].Value.(SoqlFieldInfo)
			}
			return AstSlice{{
				ClassName: "soql:SelectFieldList",
				Type:      AstType_Any,
				Value:     z,
			}}, nil
		},
	)
}

func fromClause() ParserFn {
	return Trans(
		FlatGroup(
			erase(SeqI("from")),
			sp1(),
			First(
				FlatGroup(
					notAheadReservedKeywords(),
					complexSymbolName(),
					First(
						// Alias name
						FlatGroup(
							notAheadReservedKeywords(),
							symbolName(),
							sp0(),
						),
						Zero(Ast{Value: ""}),
					),
					// TODO: `USING SCOPE` clause
					// TODO: hinting string
				),
				Error("Unexpected token aheads near by the 'from' clause"),
			),
			ZeroOrMoreTimes(
				erase(CharClass(",")),
				First(
					FlatGroup(
						sp0(),
						notAheadReservedKeywords(),
						complexSymbolName(),
						First(
							// Alias name
							FlatGroup(
								notAheadReservedKeywords(),
								symbolName(),
								sp0(),
							),
							Zero(Ast{Value: ""}),
						),
					),
					Error("Unexpected token aheads near by the 'from' clause"),
				),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			astsLen := len(asts) / 2
			z := make([]SoqlObjectInfo, astsLen, astsLen)
			for i := 0; i < astsLen; i++ {
				z[i] = SoqlObjectInfo{
					Name:      asts[i*2].Value.([]string),
					AliasName: asts[i*2+1].Value.(string),
				}
			}
			return AstSlice{{
				ClassName: "soql:From",
				Type:      AstType_Any,
				Value:     z,
			}}, nil
		},
	)
}
