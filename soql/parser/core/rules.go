package core

import (
	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
	. "github.com/shellyln/takenoco/base"
	objparser "github.com/shellyln/takenoco/object"
)

var conditionExpressionExprRule3 = Precedence{
	Rules: []ParserFn{
		Trans(
			FlatGroup(
				isOperator("soql:UnaryOp", []string{"not"}),
				anyOperand(),
			),
			func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
				cond1 := asts[1].Value.([]SoqlCondition)

				condLen := len(cond1) + 1
				cond := make([]SoqlCondition, 0, condLen)

				cond = append(cond, cond1...)
				cond = append(cond, SoqlCondition{
					Opcode: SoqlConditionOpcode_Not,
				})

				return AstSlice{{
					ClassName: "soql:Condition",
					Value:     cond,
				}}, nil
			},
		),
	},
	Rtol: false,
}

var conditionExpressionExprRule2 = Precedence{
	Rules: []ParserFn{
		Trans(
			FlatGroup(
				anyOperand(),
				isOperator("soql:BinaryOp", []string{"and"}),
				anyOperand(),
			),
			func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
				cond1 := asts[0].Value.([]SoqlCondition)
				cond2 := asts[2].Value.([]SoqlCondition)

				condLen := len(cond1) + len(cond2) + 1
				cond := make([]SoqlCondition, 0, condLen)

				cond = append(cond, cond1...)
				cond = append(cond, cond2...)
				cond = append(cond, SoqlCondition{
					Opcode: SoqlConditionOpcode_And,
				})

				return AstSlice{{
					ClassName: "soql:Condition",
					Value:     cond,
				}}, nil
			},
		),
	},
	Rtol: false,
}

var conditionExpressionExprRule1 = Precedence{
	Rules: []ParserFn{
		Trans(
			FlatGroup(
				anyOperand(),
				isOperator("soql:BinaryOp", []string{"or"}),
				anyOperand(),
			),
			func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
				cond1 := asts[0].Value.([]SoqlCondition)
				cond2 := asts[2].Value.([]SoqlCondition)

				condLen := len(cond1) + len(cond2) + 1
				cond := make([]SoqlCondition, 0, condLen)

				cond = append(cond, cond1...)
				cond = append(cond, cond2...)
				cond = append(cond, SoqlCondition{
					Opcode: SoqlConditionOpcode_Or,
				})

				return AstSlice{{
					ClassName: "soql:Condition",
					Value:     cond,
				}}, nil
			},
		),
	},
	Rtol: false,
}

var conditionExpressionExprProdRule TransformerFn = ProductionRule(
	[]Precedence{
		conditionExpressionExprRule3,
		conditionExpressionExprRule2,
		conditionExpressionExprRule1,
	},
	FlatGroup(Start(), objparser.Any(), objparser.End()),
)
