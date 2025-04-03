package main

type Expression interface {
	exprNode()
	String() string
}
