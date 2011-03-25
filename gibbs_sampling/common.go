package lda

import "rand"

type Distribution []float64
type Histogram    []int

func NewDistribution(dim int) Distribution {
	return make(Distribution, dim)
}

func NewHistogram(dim int) Histogram {
	return make(Histogram, dim)
}

func (d Distribution) IsValid() bool {
	var sum float64 = 0.0
	for _, v := range d {
		sum += v
	}
	return (sum - 1)*(sum - 1) < 0.00001
}

func GetAccumulativeSample(distribution Distribution) int {
	distribution_sum := 0.0
	for _, v := range distribution {
		distribution_sum += v
	}
	choice := rand.Float64() * float64(distribution_sum)

	sum_so_far := 0.0
	for i, v := range distribution {
		sum_so_far += v
		if sum_so_far >= choice {
			return i;
		}
	}
	return -1;
}
