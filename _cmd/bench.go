// +build bench,riak

package main

import (
	"fmt"
	"github.com/philhofer/rkive"
	"os"
	"testing"
)

// To run:
// `go run rkive/bench.go -tags 'bench riak'
//

type Benchmarker struct {
	Client  *rkive.Client
	LastAvg uint64
}

type TestObject struct {
	info rkive.Info
	Data []byte
}

func (t *TestObject) Info() *rkive.Info        { return &t.info }
func (t *TestObject) Marshal() ([]byte, error) { return t.Data, nil }
func (t *TestObject) Unmarshal(b []byte) error { t.Data = append(t.Data[0:0], b...); return nil }

func (t *Benchmarker) Store() func(b *testing.B) {
	return func(b *testing.B) {
		b.N /= 100
		ob := &TestObject{
			Data: []byte("Hello World"),
		}

		err := t.Client.New(ob, "tesbucket", nil, nil)
		if err != nil {
			b.Fatal(err)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err = t.Client.Store(ob, nil)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()
		t.LastAvg = t.Client.AvgWait()
	}
}

func (t *Benchmarker) Fetch() func(b *testing.B) {
	return func(b *testing.B) {
		b.N /= 100
		ob := &TestObject{
			Data: []byte("Hello World"),
		}

		err := t.Client.New(ob, "testbucket", nil, nil)
		if err != nil {
			b.Fatal(err)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err = t.Client.Fetch(ob, "testbucket", ob.Info().Key(), nil)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()
		t.LastAvg = t.Client.AvgWait()
	}
}

func main() {
	cl, err := rkive.DialOne("localhost:8087", "bench-client")
	if err != nil {
		fmt.Printf("FATAL: %s\n", err)
		os.Exit(1)
	}
	b := Benchmarker{
		Client: cl,
	}
	fmt.Println("-------- STORE --------")
	storeRes := testing.Benchmark(b.Store())
	fmt.Printf("I/O time/op: %dns\n", b.LastAvg)
	fmt.Printf("Client time/op: %dns\n", storeRes.NsPerOp()-int64(b.LastAvg))
	fmt.Printf("Total time/op: %dns\n", storeRes.NsPerOp())
	fmt.Printf("Allocs/op: %d\n", storeRes.AllocsPerOp())
	fmt.Printf("Mem/op: %dB\n", storeRes.AllocedBytesPerOp())
	fmt.Print("\n")

	fmt.Println("-------- FETCH --------")
	fetchRes := testing.Benchmark(b.Fetch())
	fmt.Printf("I/O time/op: %dns\n", b.LastAvg)
	fmt.Printf("Client time/op: %dns\n", fetchRes.NsPerOp()-int64(b.LastAvg))
	fmt.Printf("Total time/op: %dns\n", fetchRes.NsPerOp())
	fmt.Printf("Allocs/op: %d\n", fetchRes.AllocsPerOp())
	fmt.Printf("Mem/op: %dB\n", fetchRes.AllocedBytesPerOp())

	b.Client.Close()
}
