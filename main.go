package main

import "github.com/KitchenMishap/pudding-shed/jobs"

func main() {
	err := jobs.SeveralYearsPrimaries(16, "delegated")
	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	}
	println("End of main()")
}
