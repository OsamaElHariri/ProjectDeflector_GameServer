package colors

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

func GetPlayerColors(playerId string, colorCount int) []string {

	hashGen := md5.New()
	hash := hashGen.Sum([]byte(playerId + time.Now().Format("YYYY-MM-DD")))

	var seed uint64 = binary.BigEndian.Uint64(hash)
	rand.Seed(int64(seed))

	colors := make([]string, colorCount)
	for i := 0; i < colorCount; i++ {
		color := colorful.Hcl(rand.Float64()*360.0, 90/150.0, 0.6+rand.Float64()*0.4)
		colors[i] = color.Hex()
	}
	return colors
}
