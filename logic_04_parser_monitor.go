package main

func (p *Parser) parseMonitor() *Monitor {
	var result Monitor

	// 1. Monitor keyword
	if !p.expect(
		&paramsExpect{
			Caller:       "parseMonitor - 1",
			KindExpected: tokenMonitor,
		},
	) {
		return nil
	}

	// 2. Monitor name (string)
	if !p.expect(
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
	if !p.expect(
		&paramsExpect{
			Caller:       "parseMonitor - 3",
			KindExpected: tokenLeftBrace,
		},
	) {
		return nil
	}

	// 4. Body parsing (improved keyword detection)
	for !p.currentTokenIs(tokenRightBrace) {
		switch {
		case p.currentTokenIs(tokenLevel):
			if r := p.parseRule(); r != nil {
				result.Rules = append(result.Rules, r)

				continue
			}
			p.skipToIdentifierRightBrace(_dslLevel)

		default:
			p.errorf(
				"unexpected token in monitor: %v (%s)",
				p.tokenCurrent.kind,
				p.tokenCurrent.valueLiteral,
			)

			p.skipToIdentifierRightBrace(_dslLevel)
		}
	}

	// 5. Closing brace
	if !p.expect(
		&paramsExpect{
			Caller:       "parseMonitor - 5",
			KindExpected: tokenRightBrace,
		},
	) {
		p.tryRecoverAtBlockEnd()

		return nil
	}

	return &result
}
