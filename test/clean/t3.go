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
        switch a {
        case 5:
            a = 6
            break
        case 6: 
            a = 7
        default:
            a = 0
            break
        }
        logger.Println(a)
	}
}
