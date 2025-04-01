package main

type expression interface {
	exprNode()
	string() string
}
