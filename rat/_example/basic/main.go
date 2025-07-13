package main

import (
	"log"

	"github.com/aca/qwer/rat"
)

func main() {
	// log.Println(rat.Rat("0.1").Add("0.2").Equal("0.3"))
	// log.Println(!rat.Rat("0.1").Add("0.1", 0.1).Equal("0.3"))
	// log.Println(rat.Rat("0.1").Add("0.1", "0.1").Equal("0.3"))
	// log.Println(rat.Rat("0.1").Add("0.1", rat.Rat("0.1")).Equal("0.3"))
	// log.Println(rat.Rat("-5").Ceil())
	// log.Println(rat.Rat("5").Ceil())
	// log.Println(rat.Rat("5.1").Ceil())
	// log.Println(rat.Rat("-5.1").Ceil())
	// log.Println(rat.Rat("3%").Mul(10))

	log.Println(rat.Rat("1").Quo(10))
}
