package rkive

import (
	"testing"
)

func TestAddRemoveLink(t *testing.T) {
	info := Info{}

	info.AddLink("testlink", "testbucket", "k")

	bucket, key := info.GetLink("testlink")
	if bucket != "testbucket" || key != "k" {
		t.Errorf("Bucket: %q; key: %q", bucket, key)
	}

	info.RemoveLink("testlink")
	bucket, key = info.GetLink("testlink")
	if bucket != "" || key != "" {
		t.Errorf("Bucket: %q; key: %q", bucket, key)
	}

	info.AddLink("testlink", "testbucket", "k1")
	info.SetLink("testlink", "newbucket", "k2")

	bucket, key = info.GetLink("testlink")
	if bucket != "newbucket" || key != "k2" {
		t.Errorf("Bucket: %q; key: %q", bucket, key)
	}
}

func TestAddRemoveIndex(t *testing.T) {
	info := Info{}

	info.AddIndex("testidx", "blah")

	val := info.GetIndex("testidx")
	if val != "blah" {
		t.Errorf("Val: %q", val)
		t.Errorf("Indexes: %v", info.idxs)
	}

	info.SetIndex("testidx", "newblah")
	val = info.GetIndex("testidx")
	if val != "newblah" {
		t.Errorf("Val: %q", val)
	}

	info.RemoveIndex("testidx")
	val = info.GetIndex("testidx")
	if val != "" {
		t.Errorf("Val: %q", val)
	}

	info.AddIndexInt("myNum", 300)

	ival := info.GetIndexInt("myNum")
	if ival == nil || *ival != 300 {
		t.Errorf("Ival is %d; expected %d", *ival, 300)
	}

	info.SetIndexInt("myNum", -84)
	ival = info.GetIndexInt("myNum")
	if ival == nil || *ival != -84 {
		t.Errorf("Ival is %d; expected %d", *ival, -84)
	}

	info.RemoveIndexInt("myNum")
	ival = info.GetIndexInt("myNum")
	if ival != nil {
		t.Errorf("Expected nil; got %d", *ival)
	}
}
