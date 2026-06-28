package utils

import (
	"math/rand"
	"time"
)

func RandomCapsString() string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	rand.Seed(time.Now().UnixNano())

	result := make([]rune, 5)
	for i := 0; i < 5; i++ {
		result[i] = letters[rand.Intn(len(letters))]
	}

	return string(result)
}
