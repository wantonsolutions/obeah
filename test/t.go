package main

import (
	"github.com/wantonsolutions/obeah/obeah"
	"log"
	"math/rand"
	"os"
)

const (
	RUNS = 20
	MOD  = 20
)

func main() {
	logger := log.New(os.Stdout, "[obeah test]", log.Lshortfile)
	obeah.Taboo()

	for i := 0; i < RUNS; i++ {
		a := rand.Int() % MOD
		b := rand.Int() % MOD
		obeah.Taboo(a, b)
        if a == a {
           b--
        }
		if a < 5 {
			if a < b {
				a++
			} else if a == b {
				logger.Fatalf("CRASH!!!")
			} else {
				b++
			}
			a = 5
		} else if a > 10 || b > 10 {
			a, b = 100, 100
		} else {
			a = 7
			b = 10
		}
	}
}
