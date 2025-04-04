package dslalert

func (p *parser) parseMonitor() *monitor {
	var result monitor

	// 1. Monitor keyword
	if !p.expectWTokenAdvance(
		&paramsExpect{
			Caller:       "parseMonitor - 1",
			KindExpected: tokenMonitor,
		},
	) {
		return nil
	}

	// 2. Monitor name (string)
	if !p.expectNoTokenAdvance(
		&paramsExpect{
			Caller:       "parseMonitor - 2",
			KindExpected: tokenStringLiteral,
		},
	) {
		return nil
	}

	result.ColumnName = p.tokenCurrent.valueLiteral
	p.advanceToken()

	// 3. Opening brace
	if !p.expectWTokenAdvance(
		&paramsExpect{
			Caller:       "parseMonitor - 3",
			KindExpected: tokenLeftBrace,
		},
	) {
		return nil
	}

	if p.currentTokenIs(tokenRightBrace) {
		p.errorf("monitor must contain at least one level rule")

		return nil
	}

	// 4. Body parsing with strict advancement control
	for !p.currentTokenIs(tokenRightBrace) && !p.currentTokenIs(tokenEOF) {
		switch {
		case p.currentTokenIs(tokenLevel):
			if r := p.parseRule(); r != nil {
				result.Rules = append(result.Rules, r)

				continue
			}
			// Fallthrough to error handling if parseRule failed

		default:
			p.errorf(
				"parseMonitor - 4: unexpected token %v (%s)",
				p.tokenCurrent.kind,
				p.tokenCurrent.valueLiteral,
			)
		}

		// Critical: ensure token advancement in all cases
		if !p.currentTokenIs(tokenRightBrace) && !p.currentTokenIs(tokenEOF) {
			p.advanceToken()
		}
	}

	// 5. Closing brace validation
	if !p.expectWTokenAdvance(
		&paramsExpect{
			Caller:       "parseMonitor - 5",
			KindExpected: tokenRightBrace,
		},
	) {
		p.errorf(
			"missing closing brace '}' for monitor '%s'",
			result.ColumnName,
		)

		return nil
	}

	return &result
}
