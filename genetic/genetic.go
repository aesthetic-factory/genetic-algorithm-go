package genetic

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type Attributes []int

type Child struct {
	Score int
	Attr  Attributes
}

const ATTR_LEN = 200
const MAX_POPULATION = 8000
const THREAD = 20

var g_seed int64 = time.Now().UnixNano()

var g_population []Child

type FitnessFunction func(child *Child)

func Mutation(attributeLen int) []int {
	var result = make([]int, attributeLen)
	for idx, _ := range result {
		result[idx] = rand.Intn(400) - 200
	}
	return result
}

func Crossover(attributeLen int, p1 Attributes, p2 Attributes) []int {

	var result = make(Attributes, attributeLen)

	parentProb := rand.Intn(100)
	offsetProb := rand.Intn(100)

	for idx, _ := range result {
		if idx%15 == 0 {
			parentProb = rand.Intn(200+parentProb) % 100
		}
		if idx%3 == 0 {
			offsetProb = rand.Intn(200+offsetProb) % 100
		}
		var offset int
		if offset = 0; offsetProb < 3 {
			offset = rand.Intn(8) - 4
		} else if offsetProb < 6 {
			offset = rand.Intn(4) - 2
		}

		if parentProb < 48 {
			result[idx] = p1[idx] + offset
		} else if parentProb < 96 {
			result[idx] = p2[idx] + offset
		} else {
			result[idx] = rand.Intn(400) - 200
		}
	}
	return result
}

func GeneratePopulation(calcFitness FitnessFunction) []Child {
	var population []Child
	for len(population) < MAX_POPULATION {
		child := Child{Score: 0, Attr: Mutation(ATTR_LEN)}
		// for i := range child.Attr {
		// 	child.Attr[i] = rand.Intn(400) - 200
		// }
		calcFitness(&child)
		population = append(population, child)
	}
	sort.Slice(population, func(i, j int) bool {
		return population[i].Score > population[j].Score
	})
	return population
}

func BreedPopulationClassic(population []Child, calcFitness FitnessFunction) []Child {
	var populationNextGen []Child

	child := Child{Score: 0, Attr: Mutation(ATTR_LEN)}
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
				child := Child{Score: 0, Attr: Crossover(ATTR_LEN, a.Attr, b.Attr)}
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

func BreedPopulationWorker(population *[]Child, ch chan []Child, calcFitness FitnessFunction, wg *sync.WaitGroup) {
	defer wg.Done()

	var populationNextGen []Child
	var outputSize = MAX_POPULATION / THREAD

	var end = false
	for !end {
		for idxA, a := range *population {
			for idxB, b := range *population {
				if len(populationNextGen) > outputSize*1 {
					end = true
					break
				}
				if idxA == idxB {
					continue
				}
				child := Child{Score: 0, Attr: Crossover(ATTR_LEN, a.Attr, b.Attr)}
				calcFitness(&child)
				populationNextGen = append(populationNextGen, child)
			}
		}
		if len(*population) > 15 {
			*population = (*population)[0:15]
		}
	}

	sort.Slice(populationNextGen, func(i, j int) bool {
		return populationNextGen[i].Score > populationNextGen[j].Score
	})

	if len(populationNextGen) > 100 {
		ch <- populationNextGen[0:100]
	} else {
		ch <- populationNextGen
	}

}

func SpawnWorkers(thread int, population []Child, calcFitness FitnessFunction) []Child {
	var wg sync.WaitGroup
	ch := make(chan []Child, thread)

	child := Child{Score: 0, Attr: Mutation(ATTR_LEN)}
	calcFitness(&child)
	population = append(population, child)

	// prepare works
	for k := 0; k < thread; k++ {
		wg.Add(1)
		go BreedPopulationWorker(&population, ch, calcFitness, &wg)
	}
	wg.Wait()
	close(ch)

	var populationNextGen []Child
	for res := range ch {
		populationNextGen = append(populationNextGen, res...)
	}
	return populationNextGen
}

func BreedPopulation(population []Child, calcFitness FitnessFunction) []Child {
	g_seed = time.Now().UnixNano()
	rand.Seed(g_seed)

	// var populationScore = CalcPopulationScore(population)

	var bound = 30 //int(math.Sqrt(MAX_POPULATION)) // 15
	populationNextGen := SpawnWorkers(THREAD, population[0:bound], calcFitness)

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

func Task(calcFitness FitnessFunction) {
	g_seed = time.Now().UnixNano() + int64(rand.Intn(100000))
	rand.Seed(g_seed)
	targetScore := ATTR_LEN * 9.95
	g_population = GeneratePopulation(calcFitness)
	step := 0
	for step < 5000 {
		step += 1
		g_population = BreedPopulation(g_population, calcFitness)
		if step%10 == 0 {
			fmt.Println("BreedPopulation")
			fmt.Println("Reach Goal:", step, ", Score: ", g_population[0].Score, "Target score", targetScore)
		}
		if float64(g_population[0].Score) >= targetScore {
			fmt.Println("Reach Goal:", step, ", Score: ", g_population[0].Score)
			fmt.Println(g_population[0:2])
			break
		}
	}
}
