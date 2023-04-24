package core

import (
	"errors"
	"fmt"
	"strings"

	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
	. "github.com/shellyln/takenoco/base"
	. "github.com/shellyln/takenoco/string"
)

func conditionalOperator() ParserFn {
	return Trans(
		FlatGroup(
			First(
				Seq("!="),
				Seq("<="),
				Seq(">="),
				Seq("="),
				Seq("<"),
				Seq(">"),
				Trans(
					FlatGroup(
						SeqI("not"),
						sp1(),
						First(
							SeqI("like"),
							SeqI("in"),
						),
						wordBoundary(),
					),
				),
				Trans(
					FlatGroup(
						SeqI("like"),
						wordBoundary(),
					),
				),
				Trans(
					FlatGroup(
						SeqI("in"),
						wordBoundary(),
					),
				),
				Trans(
					FlatGroup(
						SeqI("includes"),
						wordBoundary(),
					),
				),
				Trans(
					FlatGroup(
						SeqI("excludes"),
						wordBoundary(),
					),
				),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			ast := asts[0]
			ast.ClassName = "soql:ConditionOperator"
			ast.Value = strings.ToLower(ast.Value.(string))
			return AstSlice{ast}, nil
		},
	)
}

func transWhereFieldExpression(ctx ParserContext, asts AstSlice) (AstSlice, error) {
	cond := make([]SoqlCondition, 3, 3)

	cond[0] = SoqlCondition{
		Opcode: SoqlConditionOpcode_FieldInfo,
		Value:  asts[0].Value.(SoqlFieldInfo),
	}
	cond[1] = SoqlCondition{
		Opcode: SoqlConditionOpcode_FieldInfo,
		Value:  asts[2].Value.(SoqlFieldInfo),
	}

	opcode := SoqlConditionOpcode_Noop
	switch asts[1].Value.(string) {
	case "not":
		opcode = SoqlConditionOpcode_Not
	case "and":
		opcode = SoqlConditionOpcode_And
	case "or":
		opcode = SoqlConditionOpcode_Or
	case "=":
		opcode = SoqlConditionOpcode_Eq
	case "!=":
		opcode = SoqlConditionOpcode_NotEq
	case "<":
		opcode = SoqlConditionOpcode_Lt
	case "<=":
		opcode = SoqlConditionOpcode_Le
	case ">":
		opcode = SoqlConditionOpcode_Gt
	case ">=":
		opcode = SoqlConditionOpcode_Ge
	case "like":
		opcode = SoqlConditionOpcode_Like
	case "notlike":
		opcode = SoqlConditionOpcode_NotLike
	case "in":
		opcode = SoqlConditionOpcode_In
	case "notin":
		opcode = SoqlConditionOpcode_NotIn
	case "includes":
		opcode = SoqlConditionOpcode_Includes
	case "excludes":
		opcode = SoqlConditionOpcode_Excludes
	default:
		return nil, errors.New("Unexpected token at conditional expression operator: " + fmt.Sprint(asts[1].Value))
	}
	cond[2] = SoqlCondition{
		Opcode: opcode,
	}

	return AstSlice{{
		ClassName: "soql:Condition",
		Type:      AstType_Any,
		Value:     cond,
	}}, nil
}

func whereFieldExpression() ParserFn {
	return Trans(
		FlatGroup(
			notAheadReservedKeywords(),
			Trans(
				FlatGroup(
					First(
						selectFieldFunctionCall(),
						complexSymbolName(),
						Error("Unexpected token aheads near by the 'where' clause (unknown operand1)"),
					),
					// Dummy alias name
					Zero(Ast{Value: ""}),
				),
				transComplexSelectFieldName,
			),
			First(
				conditionalOperator(),
				Error("Unexpected token aheads near by the 'where' clause (unknown operator)"),
			),
			sp0(),
			Trans(
				FlatGroup(
					// TODO: operators (* / + - ||) and parentheses
					First(
						literalValue(),
						subQuery(),
						selectFieldFunctionCall(),
						complexSymbolName(),
						listValue(),
						Error("Unexpected token aheads near by the 'where' clause (unknown operand2)"),
					),
					// Dummy alias name
					Zero(Ast{Value: ""}),
				),
				transComplexSelectFieldName,
			),
		),
		transWhereFieldExpression,
	)
}

func whereConditionExpressionInnerRoot() ParserFn {
	return FlatGroup(
		ZeroOrOnce(
			Trans(
				SeqI("not"),
				ChangeClassName("soql:UnaryOp"),
			),
			wordBoundary(),
			sp0(),
		),
		First(
			FlatGroup(
				erase(CharClass("(")),
				First(
					FlatGroup(
						sp0(),
						Indirect(whereConditionExpression),
						erase(CharClass(")")),
						sp0(),
					),
					Error("Unexpected token aheads near by the 'where' clause"),
				),
			),
			whereFieldExpression(),
		),
		ZeroOrMoreTimes(
			Trans(
				First(
					SeqI("and"),
					SeqI("or"),
				),
				ChangeClassName("soql:BinaryOp"),
			),
			wordBoundary(),
			First(
				FlatGroup(
					sp0(),
					Indirect(whereConditionExpression),
				),
				Error("Unexpected token aheads near by the 'where' clause"),
			),
		),
	)
}

func whereConditionExpression() ParserFn {
	return Trans(
		whereConditionExpressionInnerRoot(),
		conditionExpressionExprProdRule,
	)
}

func whereClause() ParserFn {
	return Trans(
		FlatGroup(
			erase(SeqI("where")),
			wordBoundary(),
			First(
				FlatGroup(
					First(
						LookAhead(CharClass("(")),
						sp1(),
						wordBoundary(),
					),
					sp0(),
					whereConditionExpression(),
				),
				Error("Unexpected token aheads near by the 'where' clause"),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			return AstSlice{{
				ClassName: "soql:Where",
				Type:      AstType_Any,
				Value:     asts[0].Value,
			}}, nil
		},
	)
}

func groupByClause() ParserFn {
	return Trans(
		FlatGroup(
			erase(SeqI("group")),
			sp1(),
			First(
				FlatGroup(
					erase(SeqI("by")),
					sp1(),
					complexSymbolName(), // TODO: scaler functions e.g. HOUR_IN_DAY(convertTimezone(CreatedDate))
					sp0(),
					ZeroOrMoreTimes(
						erase(CharClass(",")),
						sp0(),
						complexSymbolName(), // TODO: scaler functions e.g. HOUR_IN_DAY(convertTimezone(CreatedDate))
						sp0(),
					),
				),
				Error("Unexpected token aheads near by the 'group by' clause"),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			astsLen := len(asts)
			fields := make([]SoqlFieldInfo, astsLen, astsLen)
			for i := 0; i < astsLen; i++ {
				fields[i] = SoqlFieldInfo{
					Type: SoqlFieldInfo_Field,
					Name: asts[i].Value.([]string),
				}
			}
			return AstSlice{{
				ClassName: "soql:GroupBy",
				Type:      AstType_Any,
				Value:     fields,
			}}, nil
		},
	)
}

func havingFieldExpression() ParserFn {
	return Trans(
		FlatGroup(
			notAheadReservedKeywords(),
			First(
				selectFieldFunctionCall(),
				Error("Unexpected token aheads near by the 'having' clause (unknown operand1)"),
			),
			First(
				conditionalOperator(),
				Error("Unexpected token aheads near by the 'having' clause (unknown operator)"),
			),
			sp0(),
			Trans(
				FlatGroup(
					First(
						literalValue(),
						subQuery(),
						selectFieldFunctionCall(),
						listValue(),
						Error("Unexpected token aheads near by the 'having' clause (unknown operand2)"),
					),
					// Dummy alias name
					Zero(Ast{Value: ""}),
				),
				transComplexSelectFieldName,
			),
		),
		transWhereFieldExpression,
	)
}

func havingConditionExpressionInnerRoot() ParserFn {
	return Trans(
		FlatGroup(
			ZeroOrOnce(
				Trans(
					SeqI("not"),
					ChangeClassName("soql:UnaryOp"),
				),
				wordBoundary(),
				sp0(),
			),
			First(
				FlatGroup(
					erase(CharClass("(")),
					First(
						FlatGroup(
							sp0(),
							Indirect(havingConditionExpression),
							erase(CharClass(")")),
							sp0(),
						),
						Error("Unexpected token aheads near by the 'having' clause"),
					),
				),
				havingFieldExpression(),
			),
			ZeroOrMoreTimes(
				Trans(
					First(
						SeqI("and"),
						SeqI("or"),
					),
					ChangeClassName("soql:BinaryOp"),
				),
				wordBoundary(),
				First(
					FlatGroup(
						sp0(),
						Indirect(havingConditionExpression),
					),
					Error("Unexpected token aheads near by the 'having' clause"),
				),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			return asts, nil
		},
	)
}

func havingConditionExpression() ParserFn {
	return Trans(
		havingConditionExpressionInnerRoot(),
		conditionExpressionExprProdRule,
	)
}

func havingClause() ParserFn {
	return Trans(
		FlatGroup(
			erase(SeqI("having")),
			wordBoundary(),
			First(
				FlatGroup(
					First(
						LookAhead(CharClass("(")),
						sp1(),
						wordBoundary(),
					),
					sp0(),
					havingConditionExpression(),
				),
				Error("Unexpected token aheads near by the 'having' clause"),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			return AstSlice{{
				ClassName: "soql:Having",
				Type:      AstType_Any,
				Value:     asts[0].Value,
			}}, nil
		},
	)
}
