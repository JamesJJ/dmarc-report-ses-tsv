package main

import (
	"math/rand"
)

const randRunes = "abcdefghjklmnpqrstuvwxyz"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randRunes[rand.Intn(len(randRunes))]
	}
	return string(b)
}
