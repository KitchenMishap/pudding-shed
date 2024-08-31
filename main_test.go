package main

import (
	"github.com/KitchenMishap/pudding-shed/jobs"
	"testing"
)

func TestMain(t *testing.T) {
	//	err := jobs.SeveralYearsPrimaries(16, "delegated")
	err := jobs.SeveralYearsPrimaries(2, "delegated")
	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	println("End of main()")
}
