package gobenchmark

func Uniform(lo, hi float64, n int) []float64 {
	bucket := make([]float64, n)
	b := (hi - lo) / float64(n)
	for i := 0; i < n; i++ {
		bucket[i] = float64(i) * b
	}
	return bucket
}
