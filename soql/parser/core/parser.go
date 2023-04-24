package core

import (
	. "github.com/shellyln/takenoco/base"
	. "github.com/shellyln/takenoco/string"
)

func Query() ParserFn {
	return FlatGroup(
		Start(),
		sp0(),
		First(
			selectStatement(),
			Error("Unexpected token aheads"),
		),
		sp0(),
		End(),
	)
}
