package core

import (
	"strings"
	"time"

	"github.com/shellyln/go-open-soql-parser/soql/parser/core/class"
	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
	. "github.com/shellyln/takenoco/base"
	"github.com/shellyln/takenoco/extra"
	. "github.com/shellyln/takenoco/string"
)

func erase(fn ParserFn) ParserFn {
	return Trans(fn, Erase)
}

func sp0() ParserFn {
	return erase(ZeroOrMoreTimes(First(Whitespace(), comment())))
}

func sp1() ParserFn {
	return erase(OneOrMoreTimes(First(Whitespace(), comment())))
}

func lineComment() ParserFn {
	return erase(FlatGroup(
		Seq("--"),
		FlatGroup(
			ZeroOrMoreTimes(CharClassN("\r\n", "\n", "\r")),
			First(
				CharClass("\r\n", "\n", "\r"),
				LookAhead(End()),
			),
		),
	))
}

func blockComment() ParserFn {
	return erase(FlatGroup(
		Seq("/*"),
		ZeroOrMoreTimes(CharClassN("*/")),
		Seq("*/"),
	))
}

func comment() ParserFn {
	return First(lineComment(), blockComment())
}

func reservedKeywords() ParserFn {
	return Trans(
		First(
			FlatGroup(SeqI("select"), WordBoundary()),
			FlatGroup(SeqI("from"), WordBoundary()),
			FlatGroup(SeqI("where"), WordBoundary()),
			FlatGroup(SeqI("order"), sp1(), SeqI("by"), WordBoundary()),
			FlatGroup(SeqI("group"), sp1(), SeqI("by"), WordBoundary()),
			FlatGroup(SeqI("having"), WordBoundary()),
			FlatGroup(SeqI("offset"), WordBoundary()),
			FlatGroup(SeqI("limit"), WordBoundary()),
		),
		Concat,
	)
}

func notAheadReservedKeywords() ParserFn {
	return LookAheadN(
		reservedKeywords(),
		wordBoundary(),
	)
}

func wordBoundary() ParserFn {
	return WordBoundary()
}

func trueValue() ParserFn {
	return FlatGroup(
		erase(SeqI("true")),
		wordBoundary(),
		Zero(trueAst),
	)
}

func falseValue() ParserFn {
	return FlatGroup(
		erase(SeqI("false")),
		wordBoundary(),
		Zero(falseAst),
	)
}

func nullValue() ParserFn {
	return FlatGroup(
		erase(SeqI("null")),
		wordBoundary(),
		Zero(nilAst),
	)
}

func positiveInfinityValue() ParserFn {
	return FlatGroup(
		erase(FlatGroup(
			ZeroOrOnce(Seq("+")),
			SeqI("Infinity"),
		)),
		wordBoundary(),
		Zero(positiveInfinityAst),
	)
}

func negativeInfinityValue() ParserFn {
	return FlatGroup(
		erase(SeqI("-Infinity")),
		wordBoundary(),
		Zero(negativeInfinityAst),
	)
}

func nanValue() ParserFn {
	return FlatGroup(
		erase(SeqI("NaN")),
		wordBoundary(),
		Zero(nanAst),
	)
}

func decimalIntegerValue() ParserFn {
	return FlatGroup(
		Trans(extra.IntegerNumberStr(), ParseInt, ChangeClassName(class.Int)),
		wordBoundary(),
	)
}

func decimalPositiveIntegerValue() ParserFn {
	return FlatGroup(
		LookAhead(CharClassN("-")),
		Trans(extra.IntegerNumberStr(), ParseInt, ChangeClassName(class.Int)),
		wordBoundary(),
	)
}

func numberValue() ParserFn {
	return First(
		FlatGroup(
			First(
				Trans(
					FlatGroup(erase(SeqI("0b")), extra.BinaryNumberStr()),
					ParseIntRadix(2),
					ChangeClassName(class.Int),
				),
				Trans(
					FlatGroup(erase(SeqI("0o")), extra.OctalNumberStr()),
					ParseIntRadix(8),
					ChangeClassName(class.Int),
				),
				Trans(
					FlatGroup(erase(SeqI("0x")), extra.HexNumberStr()),
					ParseIntRadix(16),
					ChangeClassName(class.Int),
				),
				// TODO: Big decimal number
				Trans(
					extra.FloatNumberStr(),
					ParseFloat,
					ChangeClassName(class.Float),
				),
				Trans(
					extra.IntegerNumberStr(),
					ParseInt,
					ChangeClassName(class.Int),
				),
			),
			wordBoundary(),
		),
		positiveInfinityValue(),
		negativeInfinityValue(),
		nanValue(),
	)
}

func stringLiteralInner(cc string, multiline bool) ParserFn {
	return FlatGroup(
		erase(Seq(cc)),
		ZeroOrMoreTimes(
			First(
				FlatGroup(
					erase(Seq("\\")),
					First(
						CharClass("\\", "'", "\"", "`"),
						replaceStr(CharClass("n", "N"), "\n"),
						replaceStr(CharClass("r", "R"), "\r"),
						replaceStr(CharClass("v", "V"), "\v"),
						replaceStr(CharClass("t", "T"), "\t"),
						replaceStr(CharClass("b", "B"), "\b"),
						replaceStr(CharClass("f", "F"), "\f"),
						replaceStr(CharClass("_"), "\\_"), // for Like expression
						replaceStr(CharClass("%"), "\\%"), // for Like expression
						Trans(
							FlatGroup(
								erase(CharClass("u")),
								Repeat(Times{Min: 4, Max: 4}, HexNumber()),
							),
							ParseIntRadix(16),
							StringFromInt,
						),
						Trans(
							FlatGroup(
								erase(CharClass("u{")),
								Repeat(Times{Min: 1, Max: 6}, HexNumber()),
								erase(CharClass("}")),
							),
							ParseIntRadix(16),
							StringFromInt,
						),
						Trans(
							FlatGroup(
								erase(CharClass("x")),
								Repeat(Times{Min: 2, Max: 2}, HexNumber()),
							),
							ParseIntRadix(16),
							StringFromInt,
						),
						Trans(
							FlatGroup(
								Repeat(Times{Min: 3, Max: 3}, OctNumber()),
							),
							ParseIntRadix(8),
							StringFromInt,
						),
					),
				),
				If(multiline,
					OneOrMoreTimes(CharClassN(cc, "\\")),
					OneOrMoreTimes(
						First(
							FlatGroup(
								CharClass("\r", "\n"),
								Error("An unexpected newline has appeared in the string literal."),
							),
							CharClassN(cc, "\\"),
						),
					),
				),
			),
		),
		First(
			FlatGroup(End(), Error("An unexpected termination has appeared in the string literal.")),
			erase(Seq(cc)),
		),
	)
}

func stringValue() ParserFn {
	return Trans(
		stringLiteralInner("'", true),
		Concat,
		ChangeClassName("soql:StringValue"),
	)
}

func dateValue() ParserFn {
	return Trans(
		FlatGroup(
			extra.DateStr(),
			wordBoundary(),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			value := asts[len(asts)-1].Value.(string)
			t, err := time.Parse("2006-01-02", value)
			if err != nil {
				return nil, err
			}
			return AstSlice{{
				ClassName: class.Date,
				Type:      AstType_Any,
				Value:     t.UTC(),
			}}, nil
		},
	)
}

func dateTimeValue() ParserFn {
	return Trans(
		FlatGroup(
			extra.DateTimeStr(),
			wordBoundary(),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			// TODO: BUG: Cannot parse years with negative values or years greater than or equal to 10000.
			value := asts[len(asts)-1].Value.(string)
			t, err := time.Parse("2006-01-02T15:04:05.000000000-07:00", value)
			if err != nil {
				return nil, err
			}
			return AstSlice{{
				ClassName: class.DateTime,
				Type:      AstType_Any,
				Value:     t.UTC(),
			}}, nil
		},
	)
}

func timeValue() ParserFn {
	return Trans(
		FlatGroup(
			extra.TimeStr(),
			wordBoundary(),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			value := "1970-01-01T" + asts[len(asts)-1].Value.(string) + "+00:00"
			t, err := time.Parse("2006-01-02T15:04:05.000000000-07:00", value)
			if err != nil {
				return nil, err
			}
			return AstSlice{{
				ClassName: class.Time,
				Type:      AstType_Any,
				Value:     t.UTC(),
			}}, nil
		},
	)
}

func symbolStringValue() ParserFn {
	return Trans(
		stringLiteralInner("\"", false),
		Concat,
		ChangeClassName(class.SymbolString),
	)
}

// TODO: hinting string value: `...`
// e.g. SELECT (
//          SELECT
//              id          xid    `type:"id"   description:"..."`
//            , name        xname  `type:"text" description:"..."`
//            , fz.id       fzid   `type:"id"   description:"..."`
//            , qux.id      qxid   `type:"id"   description:"..."`
//            , qux.quux.id quid   `type:"id"   description:"..."`
//            , category           `type:"picklist" values:"value1;text1;value2;text2"`
//            , score              `type:"decimal" precision:"16" scale:"0" notNull:"true"`
//          FROM
//              contacts con `
//                  uri:"contact.csv"
//                  references:"Account.Contact"
//                  description:"..."`
//            , con.foo `
//                  type:"csv"
//                  uri:"https://example.com/foo"
//                  references:"Foo"
//                  description:"..."`
//            , con.foo.bar `
//                  uri:"bar.csv"
//                  references:"Bar"
//                  description:"..."`
//            , con.foo.bar.baz fz `
//                  uri:"baz.csv"
//                  references:"Baz"
//                  description:"..."`
//            , con.qux `
//                  uri:"qux.csv"
//                  references:"Qux"
//                  description:"..."`
//            , con.qux.quux `
//                  uri:"quux.csv"
//                  references:"Quux"
//                  description:"..."`
//      ) FROM account acc `uri:"account.csv"`

func symbolName() ParserFn {
	return Trans(
		FlatGroup(
			CharRange(
				RuneRange{Start: 'A', End: 'Z'},
				RuneRange{Start: 'a', End: 'z'},
				RuneRange{Start: '$', End: '$'},
				RuneRange{Start: '_', End: '_'},
			),
			ZeroOrMoreTimes(
				CharRange(
					RuneRange{Start: '0', End: '9'},
					RuneRange{Start: 'A', End: 'Z'},
					RuneRange{Start: 'a', End: 'z'},
					RuneRange{Start: '$', End: '$'},
					RuneRange{Start: '_', End: '_'},
				),
			),
			wordBoundary(),
		),
		Concat,
		ChangeClassName(class.SymbolName),
	)
}

func complexSymbolName() ParserFn {
	return Trans(
		FlatGroup(
			First(
				symbolName(),
				symbolStringValue(),
			),
			sp0(),
			ZeroOrMoreTimes(
				erase(Seq(".")),
				sp0(),
				First(
					symbolName(),
					symbolStringValue(),
				),
				sp0(),
			),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			astsLen := len(asts)
			z := make([]string, astsLen, astsLen)
			for i := 0; i < astsLen; i++ {
				z[i] = asts[i].Value.(string)
			}
			return AstSlice{{
				ClassName: class.ComplexSymbol,
				Type:      AstType_Any,
				Value:     z,
			}}, nil
		},
	)
}

func parameterizedValue() ParserFn {
	return Trans(
		FlatGroup(
			erase(Seq(":")),
			sp0(),
			symbolName(),
		),
		ChangeClassName(class.ParameterizedValue),
	)
}

func dateTimeLiteralName() ParserFn {
	return Trans(
		FlatGroup(
			First(
				FlatGroup(SeqI("NEXT_N_FISCAL_QUARTERS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("LAST_N_FISCAL_QUARTERS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("NEXT_N_FISCAL_YEARS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("LAST_N_FISCAL_YEARS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("THIS_FISCAL_QUARTER"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("LAST_FISCAL_QUARTER"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("NEXT_FISCAL_QUARTER"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("THIS_FISCAL_YEAR"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("LAST_FISCAL_YEAR"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("NEXT_FISCAL_YEAR"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("NEXT_N_QUARTERS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("LAST_N_QUARTERS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("NEXT_N_MONTHS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("LAST_N_MONTHS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("NEXT_N_YEARS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("LAST_N_YEARS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("NEXT_N_WEEKS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("LAST_N_WEEKS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("LAST_N_DAYS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("NEXT_N_DAYS"), erase(Seq(":")), decimalPositiveIntegerValue()),
				FlatGroup(SeqI("LAST_90_DAYS"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("NEXT_90_DAYS"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("THIS_QUARTER"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("LAST_QUARTER"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("NEXT_QUARTER"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("LAST_MONTH"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("THIS_MONTH"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("NEXT_MONTH"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("LAST_WEEK"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("THIS_WEEK"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("NEXT_WEEK"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("THIS_YEAR"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("LAST_YEAR"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("NEXT_YEAR"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("YESTERDAY"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("TOMORROW"), Zero(Ast{Value: int64(0)})),
				FlatGroup(SeqI("TODAY"), Zero(Ast{Value: int64(0)})),
			),
			wordBoundary(),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			return AstSlice{{
				ClassName: class.DateTimeLiteralName,
				Type:      AstType_Any,
				Value: SoqlDateTimeLiteralName{
					Name: strings.ToUpper(asts[0].Value.(string)),
					N:    int(asts[1].Value.(int64)),
				},
			}}, nil
		},
	)
}

func literalValue() ParserFn {
	return FlatGroup(
		First(
			dateTimeValue(),
			dateValue(),
			timeValue(),
			numberValue(),
			stringValue(),
			trueValue(),
			falseValue(),
			nullValue(),
			parameterizedValue(),
			dateTimeLiteralName(),
		),
		sp0(),
	)
}

func listValue() ParserFn {
	return Trans(
		FlatGroup(
			erase(CharClass("(")),
			sp0(),
			literalValue(),
			ZeroOrMoreTimes(
				sp0(),
				erase(CharClass(",")),
				sp0(),
				literalValue(),
			),
			sp0(),
			erase(CharClass(")")),
			sp0(),
		),
		func(ctx ParserContext, asts AstSlice) (AstSlice, error) {
			astsLen := len(asts)
			z := make([]SoqlListItem, astsLen, astsLen)
			for i := 0; i < astsLen; i++ {
				z[i] = SoqlListItem{
					Type:  fieldInfoTypeFromClassName(asts[i].ClassName),
					Value: asts[i].Value,
				}
			}
			return AstSlice{{
				ClassName: class.List,
				Type:      AstType_Any,
				Value:     z,
			}}, nil
		},
	)
}
