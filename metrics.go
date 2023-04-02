package gobenchmark

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

type MetricValue comparable

type bucket struct {
	value float64
	num   int
}

type Histogram struct {
	bucket []*bucket
	name   string
	counts int
	lock   sync.Mutex
	sum    float64
	max    float64
	unit   string
}

type HistogramMetric struct {
	Name  string  `table:"name"`
	Unit  string  `table:"unit"`
	Total int     `table:"total"`
	Rps   float64 `table:"rps"`
	Max   float64 `table:"max"`
	Avg   float64 `table:"avg"`

	T9999 float64 `table:"t9999"`
	T999  float64 `table:"t999"`
	T99   float64 `table:"t99"`
	T95   float64 `table:"t95"`
	T90   float64 `table:"t90"`
	T50   float64 `table:"t50"`

	StdDev      float64 `table:"stddev"`
	SuccessRate float64 `table:"success"`
}

func NewHistogram(name string, rawBuckets []float64, unit string) *Histogram {

	if len(rawBuckets) == 0 {
		panic("bucket cannot be empty")
	}

	sort.Slice(rawBuckets, func(i, j int) bool {
		return rawBuckets[i] < rawBuckets[j]
	})
	bks := make([]*bucket, len(rawBuckets))
	for i, rawBucket := range rawBuckets {
		bks[i] = &bucket{
			value: rawBucket,
			num:   0,
		}
	}

	return &Histogram{
		bucket: bks,
		name:   name,
		counts: 0,
		lock:   sync.Mutex{},
		sum:    0,
		max:    0,
		unit:   unit,
	}
}

// 1/3/5/8/9
func (h *Histogram) Add(data float64) {
	h.lock.Lock()
	h.lock.Unlock()
	if h.max < data {
		h.max = data
	}
	i := sort.Search(len(h.bucket), func(i int) bool {
		return h.bucket[i].value >= data
	})

	h.counts++
	h.sum += data
	i--
	if i >= len(h.bucket) || i < 0 {
		return
	}

	h.bucket[i].num++

}

func (h *Histogram) Top(percent float64) float64 {
	if percent >= 1 {
		return h.bucket[len(h.bucket)-1].value
	}
	if percent <= 0 {
		return h.bucket[0].value
	}

	total := 0
	for _, b := range h.bucket {
		total += b.num
	}

	n := int(float64(total) * percent)
	total = 0
	for _, b := range h.bucket {
		total += b.num
		if total >= n {
			return b.value
		}
	}

	return h.bucket[len(h.bucket)-1].value
}

func (h *Histogram) Metrics(costSecond float64, success float64) *HistogramMetric {
	m := &HistogramMetric{
		Name:        h.name,
		SuccessRate: success,

		Max:  h.max,
		Unit: h.unit,
		//Avg:    h.sum / float64(h.counts),
		T9999:  h.Top(0.9999),
		T999:   h.Top(0.999),
		T99:    h.Top(0.99),
		T95:    h.Top(0.95),
		T90:    h.Top(0.9),
		T50:    h.Top(0.5),
		Total:  h.counts,
		Rps:    float64(h.counts) / costSecond,
		StdDev: h.stdDev(),
	}

	if h.counts != 0 {
		m.Avg = h.sum / float64(h.counts)
	}
	return m
}

func (h *HistogramMetric) String() string {
	return fmt.Sprintf("%s max:%.2f%v avg:%.2f%v t99:%.2f%v t95:%.2f%v t90:%.2f%v   counts:%v", h.Name, h.Max, h.Unit, h.Avg, h.Unit, h.T99, h.Unit, h.T90, h.Unit, h.T50, h.Unit, h.Total)
}

func (h *Histogram) stdDev() float64 {
	if h.counts == 0 {
		return 0
	}
	avg := h.sum / float64(h.counts)

	sum := 0.0
	total := 0.0
	for _, b := range h.bucket {
		sum += float64(b.num) * math.Abs(b.value-avg) * math.Abs(b.value-avg)
		total += float64(b.num)
	}
	return math.Sqrt(sum / total)
}
