package dslalert

import "strconv"

func (p *parser) parseRule() *rule {
	var result rule

	// 1. Level keyword
	if !p.expectWTokenAdvance(
		&paramsExpect{
			Caller:       "parseRule - 1",
			KindExpected: tokenLevel,
		},
	) {
		return nil
	}

	// 2. Number
	if !p.expectNoTokenAdvance(
		&paramsExpect{
			Caller:       "parseRule - 2",
			KindExpected: tokenNumber,
		},
	) {
		return nil
	}

	level, err := strconv.Atoi(p.tokenCurrent.valueLiteral)
	if err != nil {
		p.errorf(
			"invalid level number '%s': %v",
			p.tokenCurrent.valueLiteral,
			err,
		)

		return nil
	}

	result.Level = level

	p.advanceToken()

	// 3. When
	if !p.expectWTokenAdvance(
		&paramsExpect{
			Caller:       "parseRule - 3",
			KindExpected: tokenWhen,
		},
	) {
		return nil
	}

	p.logTokenState()

	result.Condition = p.parseExpression(0) // parse the condition expression
	if result.Condition == nil {
		p.errorf("invalid rule condition expression")

		return nil
	}

	if !p.expectWTokenAdvance(
		&paramsExpect{
			Caller:       "parseRule - 4",
			KindExpected: tokenSemicolon,
		},
	) {
		return nil
	}

	return &result
}
