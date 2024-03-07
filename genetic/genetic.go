package genetic

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

const MAX_POPULATION = 10000
const THREAD = 20
const ATTR_MAX = 10000
const ATTR_MIN = -10000

type Attributes []int

type Child struct {
	Score   int
	Attr    Attributes // range from ATTR_MIN to ATTR_MAX
	AttrLen int
}

var g_seed int64 = time.Now().UnixNano()

var g_population []Child

type FitnessFunction func(child *Child)

func Mutation(attributeLen int) []int {
	var result = make([]int, attributeLen)
	for idx, _ := range result {
		result[idx] = rand.Intn(ATTR_MAX*2) - ATTR_MAX
	}
	return result
}

func Crossover(attributeLen int, p1 Attributes, p2 Attributes, mutationProb int, offsetMultiplier int) []int {

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

	var result = make(Attributes, attributeLen)

	// prob range from 0% to 100%
	parentProb := rand.Intn(100) + mutationProb
	offsetProb := rand.Intn(100)

	for idx, _ := range result {

		// update probability values
		if idx%update_parent_prob_interval == 0 {
			parentProb = rand.Intn(100+parentProb) % 100
			parentProb += mutationProb
		}
		if idx%update_offset_prob_interval == 0 {
			offsetProb = rand.Intn(100+offsetProb) % 100
		}

		if parentProb > 100 {
			// Mutation
			result[idx] = rand.Intn(ATTR_MAX*2) - ATTR_MAX
		} else {
			// calc offset
			var offset int = 0
			if offsetProb < 10 {
				offset = offsetMultiplier * (rand.Intn(100) - 50)
			}

			if parentProb < 50 {
				// use parent 1
				result[idx] = p1[idx] + offset
			} else {
				// use parent 2
				result[idx] = p2[idx] + offset
			}

		}
	}
	return result
}

// generate completely random children
func GeneratePopulation(attributeLen int, calcFitness FitnessFunction) []Child {
	var population []Child
	for len(population) < MAX_POPULATION {
		child := Child{Score: 0, Attr: Mutation(attributeLen)}
		calcFitness(&child)
		population = append(population, child)
	}
	sort.Slice(population, func(i, j int) bool {
		return population[i].Score > population[j].Score
	})
	return population
}

func BreedPopulationClassic(attributeLen int, population []Child, calcFitness FitnessFunction) []Child {
	var populationNextGen []Child

	child := Child{Score: 0, Attr: Mutation(attributeLen)}
	calcFitness(&child)
	population = append(population, child)

	for len(populationNextGen) < MAX_POPULATION {
		for idxA, a := range population {
			for idxB, b := range population {
				if len(populationNextGen) > MAX_POPULATION {
					break
				}
				if idxA == idxB {
					continue
				}
				child := Child{Score: 0, Attr: Crossover(attributeLen, a.Attr, b.Attr, 3, idxB%10+1)}
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

func BreedPopulationWorker(attributeLen int, population *[]Child, ch chan []Child, calcFitness FitnessFunction, wg *sync.WaitGroup) {
	defer wg.Done()

	var populationNextGen []Child
	var outputSize = MAX_POPULATION/THREAD + 1

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
				child := Child{Score: 0, Attr: Crossover(attributeLen, a.Attr, b.Attr, 3, idxB)}
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

func SpawnWorkers(thread int, attributeLen int, population []Child, calcFitness FitnessFunction) []Child {
	var wg sync.WaitGroup
	ch := make(chan []Child, thread)

	child := Child{Score: 0, Attr: Mutation(attributeLen)}
	calcFitness(&child)
	population = append(population, child)

	// prepare workers
	for k := 0; k < thread; k++ {
		wg.Add(1)
		go BreedPopulationWorker(attributeLen, &population, ch, calcFitness, &wg)
	}
	wg.Wait()
	close(ch)

	var populationNextGen []Child
	for res := range ch {
		populationNextGen = append(populationNextGen, res...)
	}
	return populationNextGen
}

func BreedPopulation(attributeLen int, population []Child, calcFitness FitnessFunction) []Child {

	g_seed = time.Now().UnixNano()

	var bound = 50 // only pick the top 50 in the population
	populationNextGen := SpawnWorkers(THREAD, attributeLen, population[0:bound], calcFitness)
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

func Task(attributeLen int, calcFitness FitnessFunction, targetScore float64, maxStep int) (bool, []Child) {
	g_seed = time.Now().UnixNano() + int64(rand.Intn(100000))

	g_population = GeneratePopulation(attributeLen, calcFitness)
	step := 0
	for step < maxStep {
		step += 1
		g_population = BreedPopulation(attributeLen, g_population, calcFitness)
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
