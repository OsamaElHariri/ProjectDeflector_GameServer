package gamemechanics

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand"
	"strconv"
)

type VarianceFactory interface {
	GeneratePawnVariant(str string, turns int) []string
	GenerateDeflectionSource(gameBoard GameBoard, turn int) DirectedPosition
}

type RandomVarianceFactory struct{}

func (factory RandomVarianceFactory) GeneratePawnVariant(str string, turns int) []string {
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

func (factory RandomVarianceFactory) GenerateDeflectionSource(gameBoard GameBoard, turn int) DirectedPosition {
	hashGen := md5.New()
	hash := hashGen.Sum([]byte(gameBoard.defenition.Id + strconv.Itoa(turn)))
	var seed uint64 = binary.BigEndian.Uint64(hash)

	rand.Seed(int64(seed))
	if rand.Float64() < 0.5 {
		return DirectedPosition{
			Position:  position(gameBoard.defenition.XMax/2, gameBoard.defenition.YMax+1),
			Direction: DOWN,
		}
	} else {
		return DirectedPosition{
			Position:  position(gameBoard.defenition.XMax/2, -1),
			Direction: UP,
		}
	}
}
