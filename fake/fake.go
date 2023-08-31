package fake

import (
	"encoding/json"
	"time"

	"github.com/brianvoe/gofakeit/v5"
)

var currentSeed int64

func init() {
	Seed(time.Now().UnixMicro())
}

func Seed(seed int64) {
	currentSeed = seed
	gofakeit.Seed(currentSeed)
}

func Name() string {
	return gofakeit.Name()
}

func FirstName() string {
	return gofakeit.FirstName()
}

func LastName() string {
	return gofakeit.LastName()
}

func Email() string {
	return gofakeit.Email()
}

func Sentence(words int) string {
	return gofakeit.Sentence(words)
}

func GetInfo() string {
	bb, _ := json.MarshalIndent(struct{ Seed int64 }{Seed: currentSeed}, "", "    ")
	return string(bb)
}
