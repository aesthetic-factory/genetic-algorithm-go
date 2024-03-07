package main

import (
	"fmt"
	"math"
	"root/genetic"
)

func CalcScore(child *genetic.Child) {

	for idx, a := range child.Attr {
		if idx%2 == 0 {
			child.Score += (10 - int(math.Abs(5-float64(a/1000))))
		} else {
			child.Score += (10 - int(math.Abs(-5-float64(a/1000))))
		}
	}
}

func main() {
	success, population := genetic.Task(50, CalcScore, 50*9.98, 5000)
	if success {
		fmt.Println("success")
		for idx, val := range population {
			fmt.Println(idx)
			fmt.Println(val.Score)
			fmt.Println(val.Attr)
		}
	}

}
