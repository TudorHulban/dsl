package main

type Expression interface {
	exprNode() // TODO: needed?
	String() string
}

var _ Expression = &ExpressionBinary{}
var _ Expression = &ExpressionLiteral{}
var _ Expression = &ExpressionVariable{}
