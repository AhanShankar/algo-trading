package utils

import (
	"math/rand"
)

func RandomCapsString() string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	result := make([]rune, 5)
	for i := 0; i < 5; i++ {
		result[i] = letters[rand.Intn(len(letters))]
	}

	return string(result)
}
