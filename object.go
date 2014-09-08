package rkive

import (
	"bytes"
	"github.com/philhofer/rkive/rpbc"
	"strconv"
	"unsafe"
)

// unsafe string-to-byte
// only use this when 's' has the same scope
// as the returned byte slice, and there are guarantees
// that the slice will not be mutated.
func ustr(s string) []byte { return *(*[]byte)(unsafe.Pointer(&s)) }

// Object is the interface that must
// be satisfied in order to fetch or
// store an object in Riak.
type Object interface {
	// Objects must maintain
	// a reference to an Info
	// struct, which contains
	// this object's riak
	// metadata. Info() must
	// never return nil, or it
	// will cause a panic.
	Info() *Info

	// Marshal should return the encoded
	// value of the object, and any
	// relevant errors.
	Marshal() ([]byte, error)

	// Unmarshal should unmarshal the object
	// from a []byte. It can safely use
	// zero-copy methods, as the byte slice
	// passed to it will "belong" to the
	// object.
	Unmarshal([]byte) error
}

// Duplicator types know how to return
// an empty copy of themselves, on top of
// fulfilling the Object interface.
type Duplicator interface {
	Object
	// Empty should return an initialized
	// (zero-value) object of the same underlying
	// type as the parent.
	NewEmpty() Object
}

// ObjectM is an object that also knows how to merge
// itself with siblings. If an object has this interface
// defined, this package will use the Merge method to transparently
// handle siblings returned from Riak.
type ObjectM interface {
	Duplicator

	// Merge should merge the argument object into the method receiver. It
	// is safe to type-assert the argument of Merge to the same type
	// as the type of the object satisfying the inteface. (Under the hood,
	// the argument passed to Merge is simply the value of NewEmpty() after
	// data has been read into it.) Merge is used to iteratively merge many sibling objects.
	Merge(o Object)
}

// sibling merge - object should be Store()d after call
func handleMerge(om ObjectM, ct []*rpbc.RpbContent) error {
	var err error
	for i, ctt := range ct {
		if i == 0 {
			err = readContent(om, ctt)
			if err != nil {
				return err
			}
			continue
		}

		// read into new empty
		nom := om.NewEmpty()
		err = readContent(nom, ctt)
		nom.Info().vclock = append(nom.Info().vclock[0:0], ctt.Vtag...)
		if err != nil {
			return err
		}
		om.Merge(nom)

		// transfer vclocks if we didn't have one before
		if len(om.Info().vclock) == 0 && len(nom.Info().vclock) > 0 {
			om.Info().vclock = append(om.Info().vclock, nom.Info().vclock...)
		}
	}
	return nil
}

// Info contains information
// about a specific Riak object. You can use
// it to satisfy the Object interface.
// Info's zero value (Info{}) is valid.
// You can use the Info object to add
// links, seconary indexes, and user metadata
// to the object referencing this Info object.
// Calls to Fetch(), Push(), Store(), New(),
// etc. will changes the contents of this struct.
type Info struct {
	key    []byte          // key
	bucket []byte          // bucket
	links  []*rpbc.RpbLink // Links
	idxs   []*rpbc.RpbPair // Indexes
	meta   []*rpbc.RpbPair // Meta
	ctype  []byte          // Content-Type
	vclock []byte          // Vclock
}

func readHeader(o Object, ctnt *rpbc.RpbContent) {
	o.Info().ctype = append(o.Info().ctype[0:0], ctnt.ContentType...)
	o.Info().links = append(o.Info().links[0:0], ctnt.Links...)
	o.Info().idxs = append(o.Info().idxs[0:0], ctnt.Indexes...)
	o.Info().meta = append(o.Info().meta[0:0], ctnt.Usermeta...)
}

// read into 'o' from content
func readContent(o Object, ctnt *rpbc.RpbContent) error {
	if ctnt.GetDeleted() {
		return ErrDeleted
	}
	readHeader(o, ctnt)
	return o.Unmarshal(ctnt.Value)
}

// write into content from 'o'
func writeContent(o Object, ctnt *rpbc.RpbContent) error {
	var err error
	ctnt.Value, err = o.Marshal()
	if err != nil {
		return err
	}
	ctnt.ContentType = append(ctnt.ContentType[0:0], o.Info().ctype...)
	ctnt.Links = append(ctnt.Links[0:0], o.Info().links...)
	ctnt.Usermeta = append(ctnt.Usermeta[0:0], o.Info().meta...)
	ctnt.Indexes = append(ctnt.Indexes[0:0], o.Info().idxs...)
	return nil
}

func set(l *[]*rpbc.RpbPair, key, value []byte) {
	if l == nil || len(*l) == 0 {
		goto add
	}
	for _, item := range *l {
		if bytes.Equal(key, item.Key) {
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
		if bytes.Equal(key, item.Key) {
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
		if bytes.Equal(key, item.Key) {
			if bytes.Equal(value, item.Value) {
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
		if bytes.Equal(key, item.Key) {
			(*l)[i], (*l)[nl-1], *l = (*l)[nl-1], nil, (*l)[:nl-1]
			return
		}
	}
}

func all(l *[]*rpbc.RpbPair) [][2]string {
	nl := len(*l)
	if nl == 0 {
		return nil
	}
	out := make([][2]string, nl)
	for i, item := range *l {
		out[i] = [2]string{string(item.Key), string(item.Value)}
	}
	return out
}

// Key is the canonical riak key
func (in *Info) Key() string { return string(in.key) }

// Bucket is the canonical riak bucket
func (in *Info) Bucket() string { return string(in.bucket) }

// ContentType is the content-type
func (in *Info) ContentType() string { return string(in.ctype) }

// SetContentType sets the content-type
// to 's'.
func (in *Info) SetContentType(s string) { in.ctype = []byte(s) }

// Vclock is the vector clock value as a string
func (in *Info) Vclock() string { return string(in.vclock) }

// format key as key_bin
func fmtbin(key string) []byte {
	kl := len(key)
	kv := make([]byte, kl+4)
	copy(kv[0:], key)
	copy(kv[kl:], []byte("_bin"))
	kv = bytes.ToLower(kv)
	return kv
}

// format key as key_int
func fmtint(key string) []byte {
	kl := len(key)
	kv := make([]byte, kl+4)
	copy(kv[0:], key)
	copy(kv[kl:], []byte("_int"))
	kv = bytes.ToLower(kv)
	return kv
}

// Add adds a key-value pair to an Indexes
// object, but returns false if a key already
// exists under that name and has a different value.
// Returns true if the index already has this exact key-value
// pair, or if the pair is written in with no conflicts.
// (All XxxIndex operations append "_bin" to key values
// internally in order to comply with the Riak secondary
// index specification, so the user does not have to
// include it.)
func (in *Info) AddIndex(key string, value string) bool {
	return add(&in.idxs, fmtbin(key), []byte(value))
}

// AddIndexInt sets an integer secondary index value
// using the same conditional rules as AddIndex
func (in *Info) AddIndexInt(key string, value int64) bool {
	return add(&in.idxs, fmtint(key), ustr(strconv.FormatInt(value, 10)))
}

// Set sets a key-value pair in an Indexes object
func (in *Info) SetIndex(key string, value string) {
	set(&in.idxs, fmtbin(key), []byte(value))
}

// SetIndexInt sets a integer secondary index value
func (in *Info) SetIndexInt(key string, value int64) {
	set(&in.idxs, fmtint(key), ustr(strconv.FormatInt(value, 10)))
}

// Get gets a key-value pair in an indexes object
func (in *Info) GetIndex(key string) (val string) {
	return string(get(&in.idxs, fmtbin(key)))
}

// GetIndexInt gets an integer index value
func (in *Info) GetIndexInt(key string) *int64 {
	bts := get(&in.idxs, fmtint(key))
	if bts == nil {
		return nil
	}
	val, _ := strconv.ParseInt(string(bts), 10, 64)
	return &val
}

// RemoveIndex removes a key from the object
func (in *Info) RemoveIndex(key string) {
	del(&in.idxs, fmtbin(key))
}

// RemoveIndexInt removes an integer index key
// from an object
func (in *Info) RemoveIndexInt(key string) {
	del(&in.idxs, fmtint(key))
}

// Indexes returns a list of all of the
// key-value pairs in this object. (Key first,
// then value.) Note that string-string
// indexes will have keys postfixed with
// "_bin", and string-int indexes will
// have keys postfixed with "_int", per the
// Riak secondary index specification.
func (in *Info) Indexes() [][2]string {
	return all(&in.idxs)
}

// AddMeta conditionally adds a key-value pair
// if it didn't exist already
func (in *Info) AddMeta(key string, value string) bool {
	return add(&in.meta, []byte(key), []byte(value))
}

// SetMeta sets a key-value pair
func (in *Info) SetMeta(key string, value string) {
	set(&in.meta, []byte(key), []byte(value))
}

// GetMeta gets a meta value
func (in *Info) GetMeta(key string) (val string) {
	return string(get(&in.meta, []byte(key)))
}

// RemoveMeta deletes the meta value
// at a key
func (in *Info) RemoveMeta(key string) {
	del(&in.meta, []byte(key))
}

// Metas returns all of the metadata
// key-value pairs. (Key first, then value.)
func (in *Info) Metas() [][2]string {
	return all(&in.idxs)
}

// AddLink adds a link conditionally. It returns true
// if the value was already set to this bucket-key pair,
// or if no value existed at 'name'. It returns false otherwise.
func (in *Info) AddLink(name string, bucket string, key string) bool {
	nm := []byte(name)

	// don't duplicate
	for _, link := range in.links {
		if bytes.Equal(nm, link.GetTag()) {
			return false
		}
	}

	in.links = append(in.links, &rpbc.RpbLink{
		Bucket: []byte(bucket),
		Key:    []byte(key),
		Tag:    nm,
	})
	return true
}

// SetLink sets a link for an object
func (in *Info) SetLink(name string, bucket string, key string) {
	nm := []byte(name)
	for _, link := range in.links {
		if bytes.Equal(nm, link.GetTag()) {
			link.Bucket = []byte(bucket)
			link.Key = []byte(key)
			return
		}
	}
	in.links = append(in.links, &rpbc.RpbLink{
		Bucket: []byte(bucket),
		Key:    []byte(key),
		Tag:    nm,
	})
	return
}

// RemoveLink removes a link (if it exists)
func (in *Info) RemoveLink(name string) {
	nm := []byte(name)
	nl := len(in.links)
	if nl == 0 {
		return
	}
	for i, link := range in.links {
		if bytes.Equal(nm, link.GetTag()) {
			// swap and don't preserve order
			in.links[i], in.links[nl-1], in.links = in.links[nl-1], nil, in.links[:nl-1]
		}
	}
}

// GetLink gets a link from the object
func (in *Info) GetLink(name string) (bucket string, key string) {
	nm := []byte(name)

	for _, link := range in.links {
		if bytes.Equal(nm, link.GetTag()) {
			bucket = string(link.GetBucket())
			key = string(link.GetKey())
			return
		}
	}
	return
}
