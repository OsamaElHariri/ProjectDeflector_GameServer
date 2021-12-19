package gamemechanics

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand"
)

type PawnVariantFactory interface {
	Generate(str string, turns int) []string
}

type RandomVariantFactory struct{}

func (factory RandomVariantFactory) Generate(str string, turns int) []string {
	hashGen := md5.New()
	hash := hashGen.Sum([]byte(str))
	var seed uint64 = binary.BigEndian.Uint64(hash)

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
