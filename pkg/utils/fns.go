package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateSixDigitCode() string {
    n, _ := rand.Int(rand.Reader, big.NewInt(900000))
    return fmt.Sprintf("%06d", n.Int64()+100000)
}
