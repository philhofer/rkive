// +build bench

package rkive

import (
	"sync"
	"testing"
)

func BenchmarkStore(b *testing.B) {
	b.N /= 100

	cl, err := DialOne("localhost:8087", "testClient")
	if err != nil {
		b.Fatal(err)
	}

	ob := &TestObject{
		Data: []byte("Hello World"),
	}

	err = cl.New(ob, "tesbucket", nil, nil)
	if err != nil {
		b.Fatal(err)
	}

	lock := new(sync.Mutex)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = cl.Store(ob, nil)
		if err != nil {
			lock.Lock()
			b.Fatal(err)
			lock.Unlock()
		}
	}
	b.StopTimer()
	b.Logf("mean i/o time: %dns", cl.AvgWait())
	cl.Close()
}

func BenchmarkMultiStore(b *testing.B) {
	NCONN := 5           // nubmer of connections
	nSEND := b.N / NCONN // number of stores/goroutine
	cl, err := DialOne("localhost:8087", "testClient")
	if err != nil {
		b.Fatal(err)
	}

	obs := make([]*TestObject, NCONN)
	for i := range obs {
		obs[i] = &TestObject{
			Data: []byte("Hello World"),
		}
		err = cl.New(obs[i], "testbucket", nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	lock := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	b.ReportAllocs()
	b.ResetTimer()
	for _, ob := range obs {
		wg.Add(1)
		go func(ob *TestObject, wg *sync.WaitGroup) {
			for i := 0; i < nSEND; i++ {
				err := cl.Store(ob, nil)
				if err != nil {
					lock.Lock()
					b.Fatal(err)
					lock.Unlock()
				}
			}
			wg.Done()
		}(ob, wg)
	}
	wg.Wait()
	b.StopTimer()
	b.Logf("Mean i/o time: %dns", cl.AvgWait())
	cl.Close()
}

func BenchmarkFetch(b *testing.B) {
	b.N /= 100

	cl, err := DialOne("localhost:8087", "testClient")
	if err != nil {
		b.Fatal(err)
	}

	ob := &TestObject{
		Data: []byte("Hello World"),
	}

	err = cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		b.Fatal(err)
	}

	lock := new(sync.Mutex)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = cl.Fetch(ob, "testbucket", ob.Info().Key(), nil)
		if err != nil {
			lock.Lock()
			b.Fatal(err)
			lock.Unlock()
		}
	}
	b.StopTimer()
	b.Logf("Mean i/o time: %dns", cl.AvgWait())
	cl.Close()
}

func BenchmarkMultiFetch(b *testing.B) {
	NCONNS := 5
	nSEND := b.N / NCONNS

	cl, err := DialOne("localhost:8087", "testClient")
	if err != nil {
		b.Fatal(err)
	}

	obs := make([]*TestObject, NCONNS)
	for i := range obs {
		obs[i] = &TestObject{
			Data: []byte("Hello World"),
		}
		err = cl.New(obs[i], "testbucket", nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	lock := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	b.ReportAllocs()
	b.ResetTimer()
	for _, ob := range obs {
		wg.Add(1)
		go func(o *TestObject, wg *sync.WaitGroup) {
			for i := 0; i < nSEND; i++ {
				err := cl.Fetch(o, "testbucket", o.Info().Key(), nil)
				if err != nil {
					lock.Lock()
					b.Fatal(err)
					lock.Unlock()
				}
			}
			wg.Done()
		}(ob, wg)
	}
	wg.Wait()
	b.StopTimer()
	b.Logf("Mean i/o time: %dns", cl.AvgWait())
	cl.Close()
}
