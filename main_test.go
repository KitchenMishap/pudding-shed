package main

import (
	"github.com/KitchenMishap/pudding-shed/jobs"
	"testing"
)

func TestMain(m *testing.M) {
	//	err := jobs.SeveralYearsPrimaries(16, "delegated")
	err := jobs.SeveralYearsParallel(4, "delegated")
	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	println("End of main()")
}
