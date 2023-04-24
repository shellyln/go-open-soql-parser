package core

import (
	. "github.com/shellyln/takenoco/base"
	objparser "github.com/shellyln/takenoco/object"
)

func setStr(s string) TransformerFn {
	return func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
		return AstSlice{{
			ClassName: "setStr",
			Type:      AstType_String,
			Value:     s,
		}}, nil
	}
}

func replaceStr(fn ParserFn, s string) ParserFn {
	return Trans(fn, setStr(s))
}

func unwrapOperandItem(ctx ParserContext, asts AstSlice) (AstSlice, error) {
	return AstSlice{asts[0].Value.(Ast)}, nil
}

func makeOpMatcher(className string, ops []string) func(c interface{}) bool {
	return func(c interface{}) bool {
		ast, ok := c.(Ast)
		if !ok || ast.ClassName != className {
			return false
		}
		val := ast.Value.(string)
		for _, op := range ops {
			if op == val {
				return true
			}
		}
		return false
	}
}

func anyOperand() ParserFn {
	return Trans(objparser.Any(), unwrapOperandItem)
}

func isOperator(className string, ops []string) ParserFn {
	return Trans(objparser.ObjClassFn(makeOpMatcher(className, ops)), unwrapOperandItem)
}
