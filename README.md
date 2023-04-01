## golang 自定义性能压测框架

## USAGE
```go
package main

import (
	"context"
	"github.com/seeadoog/gobenchmark"
	"github.com/seeadoog/gobenchmark/app"
	"math/rand"
	"time"
)

// main
func main() {
	myMetric := gobenchmark.NewHistogram("custom cost", gobenchmark.Uniform(0, 1000000000, 10000), "ns")
	a := app.New("benchmark")

	bentime := a.NewApp("time")
	bentime.SetTask(func(t context.Context, b *gobenchmark.Benchmark) (err error) {
		st := time.Now()
		time.Sleep(time.Duration(rand.Int()%1) * time.Millisecond)
		myMetric.Add(float64(time.Since(st)))
		return nil
	}, gobenchmark.Uniform(0, 1000000, 10000), myMetric)

	benchCost := a.NewApp("cost")

	benchCost.SetTask(func(t context.Context, b *gobenchmark.Benchmark) (err error) {
		st := time.Now()
		time.Sleep(time.Duration(rand.Int()%1) * time.Millisecond)
		myMetric.Add(float64(time.Since(st)))
		return nil
	}, gobenchmark.Uniform(0, 1000000, 10000), myMetric)

	a.Start()
}

```

#### output
````text
test durations: 3.0478885s
+-------------+------+----------+------------+------+------------+------------+------------+------------+------------+------------+
|    NAME     | UNIT |  TOTAL   |    MAX     | AVG  |   T9999    |    T999    |    T99     |    T95     |    T90     |    T50     |
+-------------+------+----------+------------+------+------------+------------+------------+------------+------------+------------+
| cost        | us   | 36932328 |    1004.00 | 0.04 |    1000.00 |    1000.00 |    1000.00 |     900.00 |     900.00 |     900.00 |
| custom cost | ns   | 36932328 | 1004300.00 | 9.67 | 1000000.00 | 1000000.00 | 1000000.00 | 1000000.00 | 1000000.00 | 1000000.00 |
+-------------+------+----------+------------+------+------------+------------+------------+------------+------------+------------+

````