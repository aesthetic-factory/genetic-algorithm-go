package threadpool

import (
	"sync"
)

type ThreadPool[IN any, OUT any] struct {
	wg       sync.WaitGroup
	in       chan IN
	out      chan OUT
	poolSize int
}

func (tp ThreadPool[IN, OUT]) Run(poolSize int, data []IN, task func(arg IN) OUT) []OUT {
	tp.poolSize = poolSize
	tp.in = make(chan IN, len(data))
	tp.out = make(chan OUT, len(data))

	for _, d := range data {
		tp.in <- d
	}

	t := 0
	for t < tp.poolSize {
		tp.wg.Add(1)
		go func() {
			defer tp.wg.Done()
			for {
				if len(tp.in) > 0 {
					arg := <-tp.in
					var res = task(arg)
					tp.out <- res
				} else {
					break
				}
			}
		}()
		t++
	}
	tp.wg.Wait()
	close(tp.in)
	close(tp.out)
	var results []OUT
	for {
		if len(tp.out) > 0 {
			o := <-tp.out
			results = append(results, o)
		} else {
			break
		}
	}
	return results
}

// // test
// func main() {
// 	var testThreadPool = new(ThreadPool[int, int])
// 	// testThreadPool.Init(4)
// 	// data := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97}
// 	data := []int{45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45}
// 	var results = testThreadPool.Run(6, data, FibonacciRecursion)

// 	fmt.Println(results)
// }

// func FibonacciRecursion(n int) int {
// 	if n <= 1 {
// 		return n
// 	}
// 	return FibonacciRecursion(n-1) + FibonacciRecursion(n-2)
// }
