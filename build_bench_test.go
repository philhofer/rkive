// +build riak

package rkive

import (
	"testing"
	"time"
)

func BenchmarkStore(b *testing.B) {
	cl, err := DialOne("localhost:8087", "bench-client")
	if err != nil {
		b.Fatal(err)
	}

	b.N /= 10
	ob := &TestObject{
		Data: []byte("Hello World"),
	}

	err = cl.New(ob, "tesbucket", nil, nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = cl.Store(ob, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.Logf("Avg iowait: %s", time.Duration(cl.AvgWait()))
	cl.Close()
}

func BenchmarkFetch(b *testing.B) {
	cl, err := DialOne("localhost:8087", "bench-client")

	b.N /= 10
	ob := &TestObject{
		Data: []byte("Hello World"),
	}

	err = cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = cl.Fetch(ob, "testbucket", ob.Info().Key(), nil)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.Logf("Avg iowait: %s", time.Duration(cl.AvgWait()))
	cl.Close()
}
