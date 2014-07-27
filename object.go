package riakpb

import (
	//"bytes"
	"github.com/philhofer/riakpb/rpbc"
)

// byte compare
func bcmpr(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, av := range a {
		if av != b[i] {
			return false
		}
	}
	return true
}

// Object is the interface that must
// be satisfied in order to store and retrieve
// an object from Riak.
type Object interface {
	// Satisfy this by
	// putting an Info{}
	// into your object.
	Info() *Info

	// Marshal should return the encoded
	// value of the object. It may use
	// the bytes passed to it as an argument
	// in order to reduce allocations, although
	// it should not count on that slice not being
	// nil.
	Marshal([]byte) ([]byte, error)

	// Unmarshal should unmarshal the object
	// from a []byte. It can safely use
	// zero-copy methods.
	Unmarshal([]byte) error
}

// Info contains information
// about a riak object. You can use
// it to satisfy the Object interface.
type Info struct {
	key    []byte          // key
	bucket []byte          // bucket
	links  []*rpbc.RpbLink // Links
	idxs   []*rpbc.RpbPair // Indexes
	meta   []*rpbc.RpbPair // Meta
	ctype  []byte          // Content-Type
	vclock []byte          // Vclock
	value  []byte          // value
}

func readHeader(o Object, ctnt *rpbc.RpbContent) {
	if o.Info() == nil {
		panic("nil Info")
	}
	o.Info().ctype = ctnt.GetContentType()
	o.Info().links = ctnt.GetLinks()
	o.Info().idxs = ctnt.GetIndexes()
	o.Info().meta = ctnt.GetUsermeta()
}

// read into 'o' from content
func readContent(o Object, ctnt *rpbc.RpbContent) error {
	// just in case
	if o.Info() == nil {
		panic("nil Info")
	}

	o.Info().ctype = ctnt.GetContentType()
	o.Info().links = ctnt.GetLinks()
	o.Info().idxs = ctnt.GetIndexes()
	o.Info().meta = ctnt.GetUsermeta()

	// catch the ReturnHead case
	if ctnt.GetValue() == nil {
		return nil
	}
	// read content
	o.Info().value = ctnt.GetValue() // save reference
	return o.Unmarshal(ctnt.GetValue())
}

// write into content from 'o'
func writeContent(o Object, ctnt *rpbc.RpbContent) error {
	if o.Info() == nil {
		panic("nil Info")
	}

	var err error
	ctnt.Value, err = o.Marshal(nil)
	if err != nil {
		return err
	}
	ctnt.ContentType = o.Info().ctype
	ctnt.Links = o.Info().links
	ctnt.Usermeta = o.Info().meta
	ctnt.Indexes = o.Info().idxs
	return nil
}

func set(l *[]*rpbc.RpbPair, key, value []byte) {
	if l == nil || len(*l) == 0 {
		goto add
	}
	for _, item := range *l {
		if bcmpr(key, item.GetKey()) {
			item.Key = key
			item.Value = value
			return
		}
	}
add:
	*l = append(*l, &rpbc.RpbPair{
		Key:   key,
		Value: value,
	})
	return
}

func get(l *[]*rpbc.RpbPair, key []byte) []byte {
	if l == nil || len(*l) == 0 {
		return nil
	}
	for _, item := range *l {
		if bcmpr(key, item.GetKey()) {
			return item.Value
		}
	}
	return nil
}

func add(l *[]*rpbc.RpbPair, key, value []byte) bool {
	if l == nil || len(*l) == 0 {
		goto add
	}
	for _, item := range *l {
		if bcmpr(key, item.GetKey()) {
			if bcmpr(value, item.GetValue()) {
				return true
			}
			return false
		}
	}
add:
	*l = append(*l, &rpbc.RpbPair{
		Key:   key,
		Value: value,
	})
	return true
}

func del(l *[]*rpbc.RpbPair, key []byte) {
	if l == nil || len(*l) == 0 {
		return
	}
	nl := len(*l)
	for i, item := range *l {
		if bcmpr(key, item.GetKey()) {
			(*l)[i], (*l)[nl-1], *l = (*l)[nl-1], nil, (*l)[:nl-1]
			return
		}
	}
}

// Key is the canonical riak key
func (info *Info) Key() string { return string(info.key) }

// Bucket is the canonical riak bucket
func (info *Info) Bucket() string { return string(info.bucket) }

// ContentType is the content-type
func (info *Info) ContentType() string { return string(info.ctype) }

func (info *Info) SetContentType(s string) { info.ctype = []byte(s) }

// Vclock is the vector clock value as a string
func (info *Info) Vclock() string { return string(info.vclock) }

// SetVclock sets the vector clock
func (info *Info) SetVclock(s string) { info.vclock = []byte(s) }

// SetVclockBytes sets the vector clock
func (info *Info) SetVclockBytes(b []byte) { info.vclock = b }

// Add adds a key-value pair to an Indexes
// object, but returns false if a key already
// exists under that name and has a different value.
// Returns true if the index already has this exact key-value
// pair, or if the pair is written in with no conflicts.
func (info *Info) AddIndex(key string, value string) bool {
	return add(&info.idxs, []byte(key), []byte(value))
}

// Set sets a key-value pair in an Indexes object
func (info *Info) SetIndex(key string, value string) {
	set(&info.idxs, []byte(key), []byte(value))
}

// Get gets a key-value pair in an indexes object
func (info *Info) GetIndex(key string) (val string) {
	return string(get(&info.idxs, []byte(key)))
}

// Remove removes a key from an indexes object
func (info *Info) RemoveIndex(key string) {
	del(&info.idxs, []byte(key))
}

// AddMeta conditionally adds a key-value pair
// if it didn't exist already
func (info *Info) AddMeta(key string, value string) bool {
	return add(&info.meta, []byte(key), []byte(value))
}

// AddMetaBytes adds a key-value pair
// if it didn't exist already. Note that
// changes to the key-value byte slices
// will be reflected in the underlying object (!)
func (info *Info) AddMetaBytes(key []byte, value []byte) bool {
	return add(&info.meta, key, value)
}

// SetMeta sets a key-value pair
func (info *Info) SetMeta(key string, value string) {
	set(&info.meta, []byte(key), []byte(value))
}

// SetMetabytes sets a key-value pair.
// []byte caveats apply(!)
func (info *Info) SetMetaBytes(key []byte, value []byte) {
	set(&info.meta, key, value)
}

// GetMeta gets a meta value
func (info *Info) GetMeta(key string) (val string) {
	return string(get(&info.meta, []byte(key)))
}

// GetMetaBytes returns the meta value
// as a []byte. []byte caveats apply
func (info *Info) GetMetaBytes(key []byte) []byte {
	return get(&info.meta, key)
}

// RemoveMeta deletes the meta value
// at a key
func (info *Info) RemoveMeta(key string) {
	del(&info.meta, []byte(key))
}

// AddLink adds a link conditionally. It returns true
// if the value was already set to this bucket-key pair,
// or if no value existed at 'name'. It returns false otherwise.
func (info *Info) AddLink(name string, bucket string, key string) bool {
	nm := []byte(name)

	// don't duplicate
	for _, link := range info.links {
		if bcmpr(nm, link.GetTag()) {
			return false
		}
	}

	info.links = append(info.links, &rpbc.RpbLink{
		Bucket: []byte(bucket),
		Key:    []byte(key),
		Tag:    nm,
	})
	return true
}

// SetLink sets a link
func (info *Info) SetLink(name string, bucket string, key string) {
	nm := []byte(name)
	for _, link := range info.links {
		if bcmpr(nm, link.GetTag()) {
			link.Bucket = []byte(bucket)
			link.Key = []byte(key)
			return
		}
	}
	info.links = append(info.links, &rpbc.RpbLink{
		Bucket: []byte(bucket),
		Key:    []byte(key),
		Tag:    nm,
	})
	return
}

// RemoveLink removes a link (if it exists)
func (info *Info) RemoveLink(name string) {
	nm := []byte(name)
	nl := len(info.links)
	if nl == 0 {
		return
	}
	for i, link := range info.links {
		if bcmpr(nm, link.GetTag()) {
			// swap and don't preserve order
			info.links[i], info.links[nl-1], info.links = info.links[nl-1], nil, info.links[:nl-1]
		}
	}
}

// GetLink gets a link bucket-key pari
func (info *Info) GetLink(name string) (bucket string, key string) {
	nm := []byte(name)

	for _, link := range info.links {
		if bcmpr(nm, link.GetTag()) {
			bucket = string(link.GetBucket())
			key = string(link.GetKey())
			return
		}
	}
	return
}
