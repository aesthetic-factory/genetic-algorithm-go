package main

import (
	"math"
	"root/genetic"
)

func CalcScore(child *genetic.Child) {

	for idx, a := range child.Attr {
		if idx%2 == 0 {
			child.Score += (10 - int(math.Abs(105-float64(a))))
		} else {
			child.Score += (10 - int(math.Abs(-5-float64(a))))
		}
	}
}

func main() {
	cnt := 0
	for cnt < 1 {
		cnt += 1
		genetic.Task(CalcScore)
	}
}
