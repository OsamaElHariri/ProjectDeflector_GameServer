package gamemechanics

import (
	"math/rand"
)

type PawnVariantFactory interface {
	Generate(seed int, turns int) []string
}

type RandomVariantFactory struct{}

func (factory RandomVariantFactory) Generate(seed int, turns int) []string {
	rand.Seed(int64(seed))

	variants := make([]string, turns)
	for i := 0; i < turns; i++ {
		rand := rand.Float64()
		if rand < 0.5 {
			variants[i] = SLASH
		} else {
			variants[i] = BACKSLASH
		}
	}
	return variants
}
