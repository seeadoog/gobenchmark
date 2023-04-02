package gobenchmark

func Uniform(lo, hi float64, n int) []float64 {
	bucket := make([]float64, n)
	b := (hi - lo) / float64(n)
	for i := 0; i < n; i++ {
		bucket[i] = float64(i) * b
	}
	return bucket
}

func Combine(bjkss ...[]float64) []float64 {
	res := make([]float64, 0)
	for _, float64s := range bjkss {
		res = append(res, float64s...)
	}
	return res
}

var (
	DefaultBuckets = Combine(Uniform(0, 100, 1000), Uniform(100, 1000, 1000), Uniform(1000, 10000, 100), Uniform(10000, 100000, 100), Uniform(100000, 1000000, 200))
)
