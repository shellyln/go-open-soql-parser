package core

import (
	"strings"

	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
	. "github.com/shellyln/takenoco/base"
	. "github.com/shellyln/takenoco/string"
)

func orderByDirection() ParserFn {
	return Trans(
		First(
			FlatGroup(
				First(
					SeqI("asc"),
					SeqI("desc"),
				),
				wordBoundary(),
				sp0(),
			),
			Zero(Ast{Value: "asc"}),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			v := strings.ToLower(asts[0].Value.(string))
			desc := false
			if v == "desc" {
				desc = true
			}
			return AstSlice{{
				Type:  AstType_Bool,
				Value: desc,
			}}, nil
		},
	)
}

func orderByNulls() ParserFn {
	return Trans(
		First(
			FlatGroup(
				erase(SeqI("nulls")),
				sp1(),
				First(
					SeqI("first"),
					SeqI("last"),
				),
				wordBoundary(),
				sp0(),
			),
			Zero(Ast{Value: "first"}),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			v := strings.ToLower(asts[0].Value.(string))
			nullsLast := false
			if v == "last" {
				nullsLast = true
			}
			return AstSlice{{
				Type:  AstType_Bool,
				Value: nullsLast,
			}}, nil
		},
	)
}

func orderByClause() ParserFn {
	return Trans(
		FlatGroup(
			erase(SeqI("order")),
			wordBoundary(),
			First(
				FlatGroup(
					First(
						FlatGroup(
							sp1(),
							erase(SeqI("by")),
							sp1(),
							complexSymbolName(),
							orderByDirection(),
							orderByNulls(),
						),
						Error("Unexpected token aheads near by the 'order by' clause"),
					),
					ZeroOrMoreTimes(
						erase(CharClass(",")),
						First(
							FlatGroup(
								sp0(),
								complexSymbolName(),
								orderByDirection(),
								orderByNulls(),
							),
							Error("Unexpected token aheads near by the 'order by' clause"),
						),
					),
				),
				Error("Unexpected token aheads near by the 'order by' clause"),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			astsLen := len(asts) / 3
			z := make([]SoqlOrderByInfo, astsLen, astsLen)
			for i := 0; i < astsLen; i++ {
				z[i] = SoqlOrderByInfo{
					Field: SoqlFieldInfo{
						Type: SoqlFieldInfo_Field,
						Name: asts[i*3].Value.([]string),
					},
					Desc:      asts[i*3+1].Value.(bool),
					NullsLast: asts[i*3+2].Value.(bool),
				}
			}
			return AstSlice{{
				ClassName: "soql:OrderBy",
				Type:      AstType_Any,
				Value:     z,
			}}, nil
		},
	)
}

func offsetClause() ParserFn {
	return Trans(
		FlatGroup(
			erase(SeqI("offset")),
			sp1(),
			First(
				decimalIntegerValue(),
				parameterizedValue(),
				Error("Unexpected token aheads near by the 'offset' clause"),
			),
			sp0(),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			return AstSlice{{
				ClassName: "soql:Offset",
				Type:      AstType_Any,
				Value:     asts[0].Value,
			}}, nil
		},
	)
}

func limitClause() ParserFn {
	return Trans(
		FlatGroup(
			erase(SeqI("limit")),
			sp1(),
			First(
				decimalIntegerValue(),
				parameterizedValue(),
				Error("Unexpected token aheads near by the 'limit' clause"),
			),
			sp0(),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			return AstSlice{{
				ClassName: "soql:Limit",
				Type:      AstType_Any,
				Value:     asts[0].Value,
			}}, nil
		},
	)
}

func forViewClause() ParserFn {
	return Trans(
		FlatGroup(
			erase(SeqI("for")),
			sp1(),
			First(
				FlatGroup(
					SeqI("view"),
					wordBoundary(),
					sp0(),
					ZeroOrOnce(
						erase(CharClass(",")),
						sp0(),
						SeqI("reference"),
						wordBoundary(),
						sp0(),
					),
				),
				FlatGroup(
					SeqI("reference"),
					wordBoundary(),
					sp0(),
					ZeroOrOnce(
						erase(CharClass(",")),
						sp0(),
						SeqI("view"),
						wordBoundary(),
						sp0(),
					),
				),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			astsLen := len(asts)
			z := SoqlForClause{
				Update: true,
			}
			for i := 0; i < astsLen; i++ {
				v := strings.ToLower(asts[i].Value.(string))
				switch v {
				case "view":
					z.View = true
				case "reference":
					z.Reference = true
				}
			}
			return AstSlice{{
				ClassName: "soql:For",
				Type:      AstType_Any,
				Value:     z,
			}}, nil
		},
	)
}

func forUpdateClause() ParserFn {
	return Trans(
		FlatGroup(
			erase(SeqI("for")),
			sp1(),
			erase(SeqI("update")),
			wordBoundary(),
			sp0(),
			ZeroOrOnce(
				First(
					FlatGroup(
						SeqI("tracking"),
						wordBoundary(),
						sp0(),
						ZeroOrOnce(
							erase(CharClass(",")),
							sp0(),
							SeqI("viewstat"),
							wordBoundary(),
							sp0(),
						),
					),
					FlatGroup(
						SeqI("viewstat"),
						wordBoundary(),
						sp0(),
						ZeroOrOnce(
							erase(CharClass(",")),
							sp0(),
							SeqI("tracking"),
							wordBoundary(),
							sp0(),
						),
					),
				),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			astsLen := len(asts)
			z := SoqlForClause{
				Update: true,
			}
			for i := 0; i < astsLen; i++ {
				v := strings.ToLower(asts[i].Value.(string))
				switch v {
				case "tracking":
					z.UpdateTracking = true
				case "viewstat":
					z.UpdateViewstat = true
				}
			}
			return AstSlice{{
				ClassName: "soql:For",
				Type:      AstType_Any,
				Value:     z,
			}}, nil
		},
	)
}

func selectStatementInner(isSubQuery bool) ParserFn {
	return Trans(
		FlatGroup(
			erase(SeqI("select")), // NOTE: Do not check for errors here.
			sp1(),                 // It has not been determined whether it is a query/subquery or some other expression.
			selectFieldList(),
			First(
				fromClause(),
				Error("The 'from' clause is expected"),
			),
			// TODO: `using scope` clause (USING SCOPE filterScope)
			First(
				whereClause(),
				Zero(Ast{
					ClassName: "soql:Where",
				}),
			),
			// TODO: `with` clause (WITH SECURITY_ENFORCED, WITH RecordVisibilityContext (...), WITH _filteringExpression_)
			// TODO: `with` clause (WITH DATA CATEGORY _filteringExpression_)
			First(
				FlatGroup(
					// TODO: `group by rollup`, `group by cube` clause
					groupByClause(),
					First(
						havingClause(),
						Zero(Ast{
							ClassName: "soql:Having",
						}),
					),
				),
				FlatGroup(
					Zero(Ast{
						ClassName: "soql:GroupBy",
					}),
					Zero(Ast{
						ClassName: "soql:Having",
					}),
				),
			),
			First(
				orderByClause(),
				Zero(Ast{
					ClassName: "soql:OrderBy",
				}),
			),
			First(
				Trans(
					FlatGroup(
						offsetClause(),
						First(
							limitClause(),
							Zero(Ast{Value: int64(0)}),
						),
					),
					func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
						offsetAndLimit := SoqlOffsetAndLimitClause{}

						switch asts[0].Value.(type) {
						case string:
							offsetAndLimit.OffsetParamName = asts[0].Value.(string)
						default:
							offsetAndLimit.Offset = asts[0].Value.(int64)
						}

						switch asts[1].Value.(type) {
						case string:
							offsetAndLimit.LimitParamName = asts[1].Value.(string)
						default:
							offsetAndLimit.Limit = asts[1].Value.(int64)
						}

						return AstSlice{{
							ClassName: "soql:LimitAndOffset",
							Type:      AstType_Any,
							Value:     offsetAndLimit,
						}}, nil
					},
				),
				Trans(
					FlatGroup(
						limitClause(),
						First(
							offsetClause(),
							Zero(Ast{Value: int64(0)}),
						),
					),
					func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
						offsetAndLimit := SoqlOffsetAndLimitClause{}

						switch asts[1].Value.(type) {
						case string:
							offsetAndLimit.OffsetParamName = asts[1].Value.(string)
						default:
							offsetAndLimit.Offset = asts[1].Value.(int64)
						}

						switch asts[0].Value.(type) {
						case string:
							offsetAndLimit.LimitParamName = asts[0].Value.(string)
						default:
							offsetAndLimit.Limit = asts[0].Value.(int64)
						}

						return AstSlice{{
							ClassName: "soql:LimitAndOffset",
							Type:      AstType_Any,
							Value:     offsetAndLimit,
						}}, nil
					},
				),
				Zero(Ast{
					ClassName: "soql:LimitAndOffset",
					Type:      AstType_Any,
					Value:     SoqlOffsetAndLimitClause{},
				}),
			),
			First(
				forViewClause(),
				forUpdateClause(),
				FlatGroup(
					erase(SeqI("for")),
					wordBoundary(),
					Error("Unexpected token aheads near by the 'for' clause"),
				),
				Zero(Ast{
					ClassName: "soql:For",
					Type:      AstType_Any,
					Value:     SoqlForClause{},
				}),
			),
			LookAhead(
				sp0(),
				First(
					If(isSubQuery,
						erase(CharClass(")")),
						End(),
					),
					Error("Unexpected token aheads"),
				),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			var qFields []SoqlFieldInfo
			var qFrom []SoqlObjectInfo
			var qWhere []SoqlCondition
			var qGroupBy []SoqlFieldInfo
			var qHaving []SoqlCondition
			var qOrderBy []SoqlOrderByInfo
			var isAggregation bool

			if asts[0].Value != nil {
				qFields = asts[0].Value.([]SoqlFieldInfo)
			}
			if asts[1].Value != nil {
				qFrom = asts[1].Value.([]SoqlObjectInfo)
			}
			if asts[2].Value != nil {
				qWhere = asts[2].Value.([]SoqlCondition)
			}
			if asts[3].Value != nil {
				qGroupBy = asts[3].Value.([]SoqlFieldInfo)
				isAggregation = true
			}
			if asts[4].Value != nil {
				qHaving = asts[4].Value.([]SoqlCondition)
			}
			if asts[5].Value != nil {
				qOrderBy = asts[5].Value.([]SoqlOrderByInfo)
			}

			return AstSlice{{
				ClassName: "soql:Query",
				Type:      AstType_Any,
				Value: SoqlQuery{
					Fields:         qFields,
					From:           qFrom,
					Where:          qWhere,
					GroupBy:        qGroupBy,
					Having:         qHaving,
					OrderBy:        qOrderBy,
					OffsetAndLimit: asts[6].Value.(SoqlOffsetAndLimitClause),
					For:            asts[7].Value.(SoqlForClause),
					IsAggregation:  isAggregation,
				},
			}}, nil
		},
	)
}

func selectStatement() ParserFn {
	return selectStatementInner(false)
}

func subQuerySelectStatement() ParserFn {
	return selectStatementInner(true)
}
