// +build riak

package riakpb

import (
	"sync"
	"testing"
)

func TestRiakPing(t *testing.T) {
	cl, err := Dial([]string{"localhost:8087", "localhost:8087"}, "testClient")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Performing 3 x 50 pings...")

	wg := new(sync.WaitGroup)
	wg.Add(3)
	for g := 0; g < 3; g++ {
		go func(t *testing.T) {
			for i := 0; i < 50; i++ {
				err := cl.Ping()
				if err != nil {
					t.Fatal(err)
				}
			}
			wg.Done()
		}(t)
	}
	wg.Wait()
	cl.Close()
}
