package main

import "github.com/KitchenMishap/pudding-shed/jobs"

func main() {
	//	err := jobs.SeveralYearsPrimaries(16, "delegated")
	err := jobs.SeveralYearsPrimaries(16, "separate files")
	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	println("End of main()")
}
