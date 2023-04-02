package gobenchmark

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestNewBenchmark(t *testing.T) {
	b := NewBenchmark(NewContext(context.TODO(), 3*time.Second), 1, Uniform(0, 1000000, 10000), func(t context.Context, b *Benchmark) (err error) {
		time.Sleep(time.Duration(rand.Int()%1) * time.Millisecond)
		return nil
	})
	b.Start()

	fmt.Println(b.String(3, 1))
}

func TestSearch(t *testing.T) {

	a := []float64{0, 1, 4, 6, 7, 10}
	i := sort.Search(len(a), func(i int) bool {
		return a[i] >= -1
	})
	fmt.Println(a[i])
}

func TestF(t *testing.T) {
	fmt.Printf("%.6f", 1.0/1e5)
}
