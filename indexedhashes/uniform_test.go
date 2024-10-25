package indexedhashes

import "testing"

func TestUniform(t *testing.T) {

	const gigabytesMem = 1
	uhc, _, err := NewUniformHashStoreCreatorAndPreloaderFromFile(
		"F:/Data/UniformHashes", "Transactions", gigabytesMem)
	if err != nil {
		println(err.Error())
		println("There was an error :-O")
	} else {
		hs, err := uhc.openHashStorePrivate()
		if err != nil {
			println(err.Error())
			println("There was an error :-O")
		} else {
			success, err := hs.Test()
			if err != nil {
				println(err.Error())
				println("There was an error :-O")
			} else if success {
				println("success")
			} else {
				println("failure")
			}
		}
	}
}
