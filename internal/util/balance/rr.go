package balance

import (
	"sync/atomic"
	"unsafe"

	"github.com/chenlx0/echoes/internal/config"
)

type RRFunc func(domain string) *config.Upstream

func getMaxWeightIndex(ups []int) int {
	index := 0
	max := -1
	for i := 0; i < len(ups); i++ {
		if ups[i] > max {
			max = ups[i]
			index = i
		}
	}
	return index
}

func getSumWeight(w []*config.Upstream) int {
	sum := 0
	for _, ww := range w {
		sum += ww.Weight
	}
	return sum
}

func copyWeights(ups []*config.Upstream) []int {
	list := make([]int, 0)
	for _, up := range ups {
		list = append(list, up.Weight)
	}
	return list
}

// GetRoundRobin return wrr function with closure
func GetRoundRobin(rrconf map[string][]*config.Upstream) RRFunc {
	currentWeightsMap := make(map[string][]int, len(rrconf))
	sumWeightsMap := make(map[string]int, len(rrconf))
	for domain, ups := range rrconf {
		// init currentWeights
		list := make([]int, 0)
		for _, up := range ups {
			list = append(list, up.Weight)
		}
		currentWeightsMap[domain] = list

		// calculate domains weight sum
		sumWeightsMap[domain] = getSumWeight(ups)
	}

	return func(domain string) *config.Upstream {
		sumWeight := sumWeightsMap[domain]
		currentWeight := currentWeightsMap[domain]

		for i := 0; i < len(currentWeight); i++ {
			atomic.AddInt32((*int32)(unsafe.Pointer(&currentWeight[i])), int32(rrconf[domain][i].Weight))
		}

		// sub max weight to current upstream
		index := getMaxWeightIndex(currentWeight)
		newWeight := int32(currentWeight[index] - sumWeight)
		atomic.StoreInt32((*int32)(unsafe.Pointer(&currentWeight[index])), newWeight)
		return rrconf[domain][index]
	}
}
