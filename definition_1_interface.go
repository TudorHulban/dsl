package main

type Expression interface {
	exprNode()
	string() string
}
