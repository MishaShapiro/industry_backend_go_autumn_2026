package main

import "math"

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	if len(nums) < 2 {
		return Stats{}
	}

	res := Stats{Min: math.MaxInt64, Max: math.MinInt64}
	for i := 1; i < len(nums); i++ {
		diff := nums[i] - nums[i-1]
		res.Count++
		res.Sum += diff
		if diff < res.Min {
			res.Min = diff
		}
		if diff > res.Max {
			res.Max = diff
		}
	}
	return res
}
