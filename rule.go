package main

type rule struct {
	level     int        // alert level
	condition expression // the 'when' condition expression
}
