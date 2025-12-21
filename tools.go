package toolkit

import (
	"math/rand/v2"
)

const randStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type used to instantiate this module. Any variable of this type will 
// have access to all the methods with the reciever *Tools
type Tools struct{}

// RandomString generates a safe random string of length l, using randStringSource as source
// for the string.
func (t *Tools) RandomString(l int) string {
	res := make([]byte, l)
	for i := range res {
		n := rand.IntN(len(randStringSource))
		res[i] = randStringSource[n]
	}
	return string(res)
}


