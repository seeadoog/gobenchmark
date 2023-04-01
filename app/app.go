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
	"os"
	"reflect"
	"runtime"
	"time"
)

type App struct {
	cmd *cobra.Command
	ctx context.Context
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

func (c *App) Start() {
	cmd := c.cmd
	if err := cmd.Execute(); err != nil {
		fmt.Println("exec cmd error=>", err)
	}
}

func (a *App) SetTask(task gobenchmark.Task, bucket []float64, metrics ...*gobenchmark.Histogram) {

	cmd := a.cmd
	var (
		concurrency int
		procs       int
		duration    time.Duration
	)
	cmd.Flags().IntVarP(&concurrency, "concurrency", "n", 1, "concurrency per process, the final concurrency num is concurrency * proc")
	cmd.Flags().IntVarP(&procs, "proc", "p", 1, "process num")
	cmd.Flags().DurationVarP(&duration, "duration", "d", 3*time.Second, "benchmark duration")

	cmd.Flags().Bool(fork.ForkFlag, false, "forked flag,used to mark process  as children process .do not use it")

	cmd.Run = func(cmd *cobra.Command, args []string) {
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
			b := gobenchmark.NewBenchmark(gobenchmark.NewContext(a.ctx, duration), concurrency, bucket, task)
			b.Start()
			met := Metrics{
				Metrics: append([]*gobenchmark.HistogramMetric{b.Metrics().Metrics()}),
			}
			for _, metric := range metrics {
				met.Metrics = append(met.Metrics, metric.Metrics())
			}
			return json.NewEncoder(os.Stdout).Encode(met)
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

		printTable(tm)

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

		if metric.Max > sumMetrics.Max {
			sumMetrics.Max = metric.Max
		}
	}
	n := float64(len(m))
	sumMetrics.T50 /= n
	sumMetrics.T90 /= n
	sumMetrics.T95 /= n
	sumMetrics.T99 /= n
	sumMetrics.Avg /= n
	sumMetrics.T999 /= n
	sumMetrics.T9999 /= n
	return sumMetrics
}

func printTable[T any](data []T) {
	if len(data) == 0 {
		return
	}
	tlp := data[0]
	tw := table.NewWriter(os.Stdout)
	tw.SetAutoWrapText(true)
	//tw.SetBorder(false)
	headers := parseTableHeaders(reflect.TypeOf(tlp))
	tw.SetHeader(headers)
	for _, v := range data {
		tw.Append(parseTableValue(reflect.ValueOf(v)))
	}
	tw.Render()
}

func parseTableHeaders(a reflect.Type) (res []string) {
	switch a.Kind() {
	case reflect.Ptr:
		return parseTableHeaders(a.Elem())
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			ft := a.Field(i)
			fieldName := ft.Tag.Get("table")
			if fieldName != "" {
				res = append(res, fieldName)
			}
		}
	default:
		panic("cannot read table header of type:" + a.String())
	}
	return res
}

func parseTableValue(a reflect.Value) (res []string) {
	switch a.Kind() {
	case reflect.Ptr:
		return parseTableValue(a.Elem())
	case reflect.Struct:
		t := a.Type()
		for i := 0; i < a.NumField(); i++ {
			ft := t.Field(i)
			fv := a.Field(i)
			fieldName := ft.Tag.Get("table")
			if fieldName != "" {
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
