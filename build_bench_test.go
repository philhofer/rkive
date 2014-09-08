// +build riak

package rkive

import (
	"runtime"
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

func BenchmarkParallelStore(b *testing.B) {
	b.Skip("Doesn't run by default.")
	cl, err := DialOne("localhost:8087", "bench-client")
	if err != nil {
		b.Fatal(err)
	}
	defer cl.Close()
	b.N /= 10
	b.SetParallelism(maxConns / runtime.GOMAXPROCS(0))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		obj := &TestObject{
			Data: []byte("Hello World"),
		}
		err = cl.New(obj, "testbucket", nil, nil)
		if err != nil {
			b.Fatal(err)
		}
		for pb.Next() {
			cl.Store(obj, nil)
		}
	})
	b.StopTimer()
}

func BenchmarkFetch(b *testing.B) {
	cl, err := DialOne("localhost:8087", "bench-client")
	if err != nil {
		b.Fatal(err)
	}

	b.N /= 10
	ob := &TestObject{
		Data: []byte("Hello World"),
	}

	err = cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		b.Fatal(err)
	}

	key := ob.Info().Key()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = cl.Fetch(ob, "testbucket", key, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.Logf("Avg iowait: %s", time.Duration(cl.AvgWait()))
	cl.Close()
}

func BenchmarkParallelFetch(b *testing.B) {
	b.Skip("Doesn't run by default.")
	cl, err := DialOne("localhost:8087", "bench-client")
	if err != nil {
		b.Fatal(err)
	}
	defer cl.Close()
	b.N /= 10
	b.SetParallelism(maxConns / runtime.GOMAXPROCS(0))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		obj := &TestObject{
			Data: []byte("Hello World"),
		}
		err = cl.New(obj, "testbucket", nil, nil)
		if err != nil {
			b.Fatal(err)
		}
		for pb.Next() {
			cl.Fetch(obj, "testbucket", obj.Info().Key(), nil)
		}
	})
	b.StopTimer()
}

func BenchmarkCacheFetch(b *testing.B) {
	cl, err := DialOne("localhost:8087", "bench-client")
	if err != nil {
		b.Fatal(err)
	}
	defer cl.Close()
	b.N /= 10
	cache := cl.Bucket("test-cache")
	err = cache.MakeCache()
	if err != nil {
		b.Fatal(err)
	}

	// make test object
	ob := &TestObject{
		Data: []byte("Hello, World!"),
	}

	err = cache.New(ob, nil)
	if err != nil {
		b.Fatal(err)
	}

	key := ob.Info().Key()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = cache.Fetch(ob, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCounterInc(b *testing.B) {
	b.Skip("Doesn't run by default.")

	cl, err := DialOne("localhost:8087", "bench-client")
	if err != nil {
		b.Fatal(err)
	}
	defer cl.Close()
	b.N /= 10
	ctr, err := cl.Bucket("testbucket").NewCounter("bench-counter", 0)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = ctr.Add(1)
		if err != nil {
			b.Fatal(err)
		}
	}
}
