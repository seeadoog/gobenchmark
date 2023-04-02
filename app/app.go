package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	table "github.com/olekukonko/tablewriter"
	"github.com/seeadoog/fork"
	"github.com/seeadoog/gobenchmark"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"math"
	"os"
	"reflect"
	"runtime"
	"time"
)

type App struct {
	cmd           *cobra.Command
	ctx           context.Context
	PreRunFunc    func(a *App, args []string)
	Concurrency   int
	Proc          int
	Duration      time.Duration
	IgnoreHeaders []string
}

func New(name string) *App {
	return &App{
		cmd: &cobra.Command{Use: name},
		ctx: context.TODO(),
	}
}

func NewWithContext(name string, ctx context.Context) *App {
	return &App{
		cmd: &cobra.Command{Use: name},
		ctx: ctx,
	}
}

func (a *App) Cmd() *cobra.Command {
	return a.cmd
}

func (a *App) Flags() *flag.FlagSet {
	return a.cmd.Flags()
}

func (a *App) PersistentFlags() *flag.FlagSet {
	return a.cmd.PersistentFlags()
}

func (c *App) NewApp(name string) *App {
	a := New(name)
	c.cmd.AddCommand(a.cmd)
	return a
}

func (c *App) NewAppWithContext(name string, ctx context.Context) *App {
	a := NewWithContext(name, ctx)
	c.cmd.AddCommand(a.cmd)
	return a
}

func (c *App) SetContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *App) Start() {
	cmd := c.cmd
	if err := cmd.Execute(); err != nil {
		fmt.Println("exec cmd error=>", err)
	}
}

func (a *App) init() {
	if a.Concurrency <= 0 {
		a.Concurrency = 1
	}

	if a.Proc <= 0 {
		a.Proc = 1
	}

	if a.Duration <= 0 {
		a.Duration = 3 * time.Second
	}

}

func (a *App) SetTask(task gobenchmark.Task, bucket []float64, metrics ...*gobenchmark.Histogram) {
	a.init()
	cmd := a.cmd
	var (
		concurrency int           = a.Concurrency
		procs       int           = a.Proc
		duration    time.Duration = a.Duration
		count       int64
	)
	cmd.Flags().IntVarP(&concurrency, "concurrency", "n", concurrency, "concurrency per process, the final concurrency num is concurrency * proc")
	cmd.Flags().IntVarP(&procs, "proc", "p", procs, "process num")
	cmd.Flags().DurationVarP(&duration, "duration", "d", duration, "benchmark duration")

	cmd.Flags().Bool(fork.ForkFlag, false, "forked flag,used to mark process  as children process .do not use it")

	cmd.Flags().Int64VarP(&count, "count", "c", math.MaxInt64, "run times per process ,default max int64 value")

	cmd.Flags().StringSliceVar(&a.IgnoreHeaders, "ignore-h", a.IgnoreHeaders, "ignore Headers in metrics table")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if len(a.IgnoreHeaders) == 0 {
			a.IgnoreHeaders = defaultDisableMetricsHeaders
		}
		if a.PreRunFunc != nil {
			a.PreRunFunc(a, args)
		}

		start := time.Now()
		pf := fork.NewForker(procs)
		outputs := make([]*bytes.Buffer, 0, procs)
		pf.SetPreForkChild(func(child *fork.Cmd) error {
			bf := bytes.NewBuffer(nil)
			child.Cmd().Stdout = bf
			outputs = append(outputs, bf)
			return nil
		})
		err := pf.ForkProcess(func(f *fork.MasterTool) error {
			return nil
		}, func(c *fork.ChildrenTool) error {
			runtime.GOMAXPROCS(1)
			ctx, counter := gobenchmark.NewCounterContext(gobenchmark.NewContext(a.ctx, duration), count)
			b := gobenchmark.NewBenchmark(ctx, concurrency, bucket, func(t context.Context, b *gobenchmark.Benchmark) (err error) {
				err = task(t, b)
				counter.Add(1)
				return err
			})
			b.Start()
			met := Metrics{
				Metrics: append([]*gobenchmark.HistogramMetric{b.Metrics().Metrics(time.Since(start).Seconds(), b.SuccessRate())}),
			}
			for _, metric := range metrics {
				met.Metrics = append(met.Metrics, metric.Metrics(time.Since(start).Seconds(), b.SuccessRate()))
			}
			err := json.NewEncoder(os.Stdout).Encode(met)
			if err != nil {
				b.Println("err json:", met)
				return err
			}
			return err
		})
		if err != nil {
			panic(err)
		}

		if !pf.IsMaster() {
			return
		}
		pf.Wait()

		sumMetrics := make([][]*gobenchmark.HistogramMetric, len(metrics)+1)
		fmt.Println("test durations:", time.Since(start))
		tm := []*gobenchmark.HistogramMetric{}
		for _, output := range outputs {
			metric := new(Metrics)
			err = json.NewDecoder(output).Decode(metric)
			if err != nil {
				panic(err)
			}
			for i, histogramMetric := range metric.Metrics {
				sumMetrics[i] = append(sumMetrics[i], histogramMetric)
			}
		}
		for _, metric := range sumMetrics {
			tm = append(tm, SumMetrics(metric))
			//fmt.Println(SumMetrics(metric).String())
		}

		printTable(tm, a.IgnoreHeaders...)

	}

}

type Metrics struct {
	Metrics []*gobenchmark.HistogramMetric
}

func SumMetrics(m []*gobenchmark.HistogramMetric) *gobenchmark.HistogramMetric {

	sumMetrics := &gobenchmark.HistogramMetric{Name: m[0].Name, Unit: m[0].Unit}
	for _, metric := range m {

		//fmt.Println(metric.String())
		sumMetrics.T50 += metric.T50
		sumMetrics.T99 += metric.T99
		sumMetrics.T95 += metric.T95
		sumMetrics.T90 += metric.T90
		sumMetrics.T999 += metric.T999
		sumMetrics.T9999 += metric.T9999

		sumMetrics.Total += metric.Total

		sumMetrics.Avg += metric.Avg

		sumMetrics.Rps += metric.Rps

		if metric.Max > sumMetrics.Max {
			sumMetrics.Max = metric.Max
		}
		sumMetrics.StdDev += metric.StdDev
		sumMetrics.SuccessRate += metric.SuccessRate
	}
	n := float64(len(m))
	sumMetrics.T50 /= n
	sumMetrics.T90 /= n
	sumMetrics.T95 /= n
	sumMetrics.T99 /= n
	sumMetrics.Avg /= n
	sumMetrics.T999 /= n
	sumMetrics.T9999 /= n
	sumMetrics.StdDev /= n
	sumMetrics.SuccessRate /= n
	return sumMetrics
}

func printTable[T any](data []T, igs ...string) {
	if len(data) == 0 {
		return
	}
	tlp := data[0]
	tw := table.NewWriter(os.Stdout)
	tw.SetAutoWrapText(true)
	tw.SetCenterSeparator("|")
	tw.SetBorder(false)
	headers := parseTableHeaders(reflect.TypeOf(tlp), igs...)
	tw.SetHeader(headers)
	for _, v := range data {
		tw.Append(parseTableValue(reflect.ValueOf(v), igs...))
	}
	tw.Render()
}

func parseTableHeaders(a reflect.Type, igs ...string) (res []string) {
	switch a.Kind() {
	case reflect.Ptr:
		return parseTableHeaders(a.Elem(), igs...)
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			ft := a.Field(i)
			fieldName := ft.Tag.Get("table")
			if fieldName != "" && !isIgnore(igs, fieldName) {
				res = append(res, fieldName)
			}
		}
	default:
		panic("cannot read table header of type:" + a.String())
	}
	return res
}

func parseTableValue(a reflect.Value, igs ...string) (res []string) {
	switch a.Kind() {
	case reflect.Ptr:
		return parseTableValue(a.Elem(), igs...)
	case reflect.Struct:
		t := a.Type()
		for i := 0; i < a.NumField(); i++ {
			ft := t.Field(i)
			fv := a.Field(i)
			fieldName := ft.Tag.Get("table")
			if fieldName != "" && !isIgnore(igs, fieldName) {
				vv := fv.Interface()
				switch v := vv.(type) {
				case float64, float32:
					res = append(res, fmt.Sprintf("%.2f", v))

				default:
					res = append(res, fmt.Sprintf("%v", v))

				}
			}
		}
	default:
		panic("cannot read table header of type:" + a.String())
	}
	return res
}

var (
	defaultDisableMetricsHeaders = []string{"t9999", "t999"}
)

func isIgnore(table []string, name string) bool {
	for _, s := range table {
		if s == name {
			return true
		}
	}
	return false
}
