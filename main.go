package main

import (
	"fmt"
	"math"
	"root/genetic"
)

func CalcScore(child *genetic.Child) {

	for idx, a := range child.Attr {
		if idx%2 == 0 {
			child.Score += (10 - int(math.Pow(1005-float64(a), 2)))
		} else {
			child.Score += (10 - int(math.Pow(-5-float64(a), 2)))
		}
	}
}

func main() {
	var constraint genetic.AttributeConstraint
	constraint.Length = 50
	constraint.Ranges = make([]genetic.AttributeRange, constraint.Length)

	idx := 0
	for idx < constraint.Length {
		if idx%2 == 0 {
			constraint.Ranges[idx].Min = 0
			constraint.Ranges[idx].Max = 1500
		} else {
			constraint.Ranges[idx].Min = -100
			constraint.Ranges[idx].Max = 0
		}

		idx++
	}
	success, population := genetic.Task(constraint, CalcScore, float64(constraint.Length)*9.97, 500)
	if success {
		fmt.Println("success")
		for idx, val := range population {
			fmt.Println(idx)
			fmt.Println(val.Score)
			fmt.Println(val.Attr)
		}
	} else {
		fmt.Println(population[0].Score)
		fmt.Println(population[0].Attr)
	}
}
