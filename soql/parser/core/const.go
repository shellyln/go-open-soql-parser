package core

import (
	"math"

	"github.com/shellyln/go-open-soql-parser/soql/parser/core/class"
	. "github.com/shellyln/takenoco/base"
)

// Parser constant value
var nilAst = Ast{
	OpCode:    0,
	ClassName: class.Null,
	Type:      AstType_Nil,
	Value:     nil,
}

// Parser constant value
var trueAst = Ast{
	OpCode:    0,
	ClassName: class.Bool,
	Type:      AstType_Bool,
	Value:     true,
}

// Parser constant value
var falseAst = Ast{
	OpCode:    0,
	ClassName: class.Bool,
	Type:      AstType_Bool,
	Value:     false,
}

// Parser constant value
// var zeroStrAst = Ast{
// 	OpCode:    0,
// 	ClassName: class.String,
// 	Type:      AstType_String,
// 	Value:     "",
// }

// Parser constant value
var positiveInfinityAst = Ast{
	OpCode:    0,
	ClassName: class.Float,
	Type:      AstType_Float,
	Value:     math.Inf(1),
}

// Parser constant value
var negativeInfinityAst = Ast{
	OpCode:    0,
	ClassName: class.Float,
	Type:      AstType_Float,
	Value:     math.Inf(-1),
}

// Parser constant value
var nanAst = Ast{
	OpCode:    0,
	ClassName: class.Float,
	Type:      AstType_Float,
	Value:     math.NaN(),
}
