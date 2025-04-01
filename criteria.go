package main

type criteria struct {
	name     string
	settings []*setting // settings defined within this criteria
	monitors []*monitor
}
