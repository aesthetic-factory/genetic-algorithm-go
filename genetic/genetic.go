package genetic

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

const MAX_POPULATION = 10000
const THREAD = 12

type Attributes []int

type AttributeRange struct {
	Min int
	Max int
}

type AttributeConstraint struct {
	Ranges []AttributeRange
	Length int
}

type Child struct {
	Score   int
	Attr    Attributes
	AttrLen int
}

var g_seed int64 = time.Now().UnixNano()

var g_population []Child

type FitnessFunction func(child *Child)

func Mutation(attributeConstraint *AttributeConstraint) []int {
	var result = make([]int, attributeConstraint.Length)
	idx := 0
	for idx < attributeConstraint.Length {
		Max := attributeConstraint.Ranges[idx].Max
		Min := attributeConstraint.Ranges[idx].Min
		Range := Max - Min
		result[idx] = rand.Intn(Range+1) - Range/2
		idx++
	}
	return result
}

func Crossover(attributeConstraint *AttributeConstraint, p1 Attributes, p2 Attributes, mutationProb int, offsetMultiplier int, idxShifter int) []int {

	// check args
	if mutationProb < 0 {
		mutationProb = 0
	}
	if mutationProb > 10 {
		mutationProb = 10
	}

	if offsetMultiplier < 1 {
		offsetMultiplier = 1
	}
	if offsetMultiplier > 10 {
		offsetMultiplier = 10
	}

	const update_parent_prob_interval = 10
	const update_offset_prob_interval = 3

	var result = make(Attributes, attributeConstraint.Length)

	// prob range from 0% to 110%
	parentProb := rand.Intn(100) + mutationProb

	// prob range from 0% to 100%
	offsetProb := rand.Intn(100)

	idx := 0
	for idx < attributeConstraint.Length {

		Max := attributeConstraint.Ranges[idx].Max
		Min := attributeConstraint.Ranges[idx].Min
		Range := Max - Min

		// update probability values
		if (idxShifter+idx)%update_parent_prob_interval == 0 {
			parentProb = rand.Intn(100+parentProb) % 100
			parentProb += mutationProb
		}
		if (idxShifter+idx)%update_offset_prob_interval == 0 {
			offsetProb = rand.Intn(100+offsetProb) % 100
		}

		// Mutation
		if parentProb >= 100 {
			result[idx] = rand.Intn(Range) - Range/2
		} else {
			// calc offset
			var offset int = 0
			if offsetProb < 7 {
				offset = (rand.Intn(3) - 1) * offsetMultiplier
			}

			if parentProb < 50 {
				// use parent 1
				result[idx] = p1[idx] + offset
			} else {
				// use parent 2
				result[idx] = p2[idx] + offset
			}

		}
		idx++
	}
	return result
}

// generate completely random children
func GeneratePopulation(attributeConstraint *AttributeConstraint, calcFitness FitnessFunction) []Child {
	var population []Child
	for len(population) < MAX_POPULATION {
		child := Child{Score: 0, Attr: Mutation(attributeConstraint)}
		calcFitness(&child)
		population = append(population, child)
	}
	sort.Slice(population, func(i, j int) bool {
		return population[i].Score > population[j].Score
	})
	return population
}

func BreedPopulationClassic(attributeConstraint *AttributeConstraint, population []Child, calcFitness FitnessFunction, progress float32) []Child {
	var populationNextGen []Child

	child := Child{Score: 0, Attr: Mutation(attributeConstraint)}
	calcFitness(&child)
	population = append(population, child)
	offsetMultiplier := 5
	if progress > 0.4 {
		offsetMultiplier = 1
	} else if progress > 0.3 {
		offsetMultiplier = 2
	} else if progress > 0.2 {
		offsetMultiplier = 3
	}
	for len(populationNextGen) < MAX_POPULATION {
		for idxA, a := range population {
			for idxB, b := range population {
				if len(populationNextGen) > MAX_POPULATION {
					break
				}
				if idxA == idxB {
					continue
				}
				child := Child{Score: 0, Attr: Crossover(attributeConstraint, a.Attr, b.Attr, 3, offsetMultiplier, 0)}
				calcFitness(&child)
				populationNextGen = append(populationNextGen, child)
			}
		}
	}
	return populationNextGen
}

func CalcPopulationScore(population []Child) int {
	var Score = 0
	for _, child := range population {
		Score += child.Score
	}
	return Score
}

func BreedPopulationWorker(threadId int, attributeConstraint *AttributeConstraint, population *[]Child, ch chan []Child, calcFitness FitnessFunction, progress float32, wg *sync.WaitGroup) {
	defer wg.Done()

	var populationNextGen []Child
	var outputSize = MAX_POPULATION/THREAD + 1
	offsetMultiplier := 5
	if progress > 0.4 {
		offsetMultiplier = 1
	} else if progress > 0.3 {
		offsetMultiplier = 2
	} else if progress > 0.2 {
		offsetMultiplier = 3
	}
	var end = false
	for !end {
		for idxA, a := range *population {
			for idxB, b := range *population {
				if len(populationNextGen) > outputSize {
					end = true
					break
				}
				if idxA == idxB {
					continue
				}
				child := Child{Score: 0, Attr: Crossover(attributeConstraint, a.Attr, b.Attr, 3, offsetMultiplier, threadId)}
				calcFitness(&child)
				populationNextGen = append(populationNextGen, child)
			}
		}

		// If not enough populationNextGen
		// breed again using the top population
		if len(*population) > 20 {
			*population = (*population)[0:20]
		}
	}

	// in this worker thread
	// calucation score in populationNextGen
	// do a sort before returning result
	sort.Slice(populationNextGen, func(i, j int) bool {
		return populationNextGen[i].Score > populationNextGen[j].Score
	})

	if len(populationNextGen) > 250 {
		ch <- populationNextGen[0:250]
	} else {
		ch <- populationNextGen
	}

}

func SpawnWorkers(thread int, progress float32, attributeConstraint *AttributeConstraint, population []Child, calcFitness FitnessFunction) []Child {
	var wg sync.WaitGroup
	ch := make(chan []Child, thread)

	// child := Child{Score: 0, Attr: Mutation(attributeConstraint)}
	// calcFitness(&child)
	// population = append(population, child)

	// prepare workers
	for k := 0; k < thread; k++ {
		wg.Add(1)
		go BreedPopulationWorker(k, attributeConstraint, &population, ch, calcFitness, progress, &wg)
	}
	wg.Wait()
	close(ch)

	var populationNextGen []Child
	for res := range ch {
		populationNextGen = append(populationNextGen, res...)
	}
	return populationNextGen
}

func BreedPopulation(progress float32, attributeConstraint *AttributeConstraint, population []Child, calcFitness FitnessFunction) []Child {

	g_seed = time.Now().UnixNano()

	var bound = 50 // only pick the top 50 in the population
	populationNextGen := SpawnWorkers(THREAD, progress, attributeConstraint, population[0:bound], calcFitness)
	//populationNextGen := BreedPopulationClassic(population[0:bound])

	sort.Slice(populationNextGen, func(i, j int) bool {
		return populationNextGen[i].Score > populationNextGen[j].Score
	})

	if len(populationNextGen) > MAX_POPULATION {
		return populationNextGen[0:MAX_POPULATION]
	} else {
		return populationNextGen
	}
}

func Task(attributeConstraint AttributeConstraint, calcFitness FitnessFunction, targetScore float64, maxStep int) (bool, []Child) {
	g_seed = time.Now().UnixNano() + int64(rand.Intn(100000))

	fmt.Println("attributeConstraint ", attributeConstraint.Length)
	g_population = GeneratePopulation(&attributeConstraint, calcFitness)
	step := 0
	for step < maxStep {
		step += 1
		g_population = BreedPopulation(float32(step)/float32(maxStep), &attributeConstraint, g_population, calcFitness)
		if step%20 == 0 {
			fmt.Println("BreedPopulation")
			fmt.Println("Reach Goal:", step, ", Score: ", g_population[0].Score, "Target score", targetScore)
		}
		if float64(g_population[0].Score) >= targetScore {
			return true, g_population[0:5]
		}
	}
	return false, g_population[0:5]
}
