package main

import (
	"encoding/json"
	"fmt"
	"github.com/philhofer/rkive"
)

// This example demonstrates
// how to use the Duplicator
// interface to use the "Async" methods
// defined by the client.

// "Person" will be the data-type used for this example
type Person struct {
	// we need to embed the 'Info' field
	// in order to fulfill the rkive.Object interface (see person.Info())
	info rkive.Info

	First string `json:"first_name"` // first name
	Last  string `json:"last_name"`  // last name
}

// Implementing the Info method is as simple as
// returning an inline reference to the embedded field
func (p *Person) Info() *rkive.Info { return &p.info }

// Marshal needs to marshal an object into bytes.
// Here we simply use json.Marshal
func (p *Person) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// Unmarshal needs to unmarshal an object from bytes.
// Since we used json.Marshal, we need json.Unmarshal here.
func (p *Person) Unmarshal(b []byte) error {
	return json.Unmarshal(b, p)
}

// NewEmpty is the method required to satisfy the Duplicator
// interface. It is almost always as simple as returning a reference
// to a zero-initialized value of the same type as the method receiver.
func (p *Person) NewEmpty() rkive.Object { return &Person{} }

func main() {

	// first, we need to set up a client. make sure
	// you have the database running locally first.
	riak, err := rkive.DialOne("localhost:8087", "demo-client")

	// we'll use the "people" bucket for Person structs
	people := riak.Bucket("people")

	// let's make some people
	// and put them in the database
	bob := &Person{
		First: "Bob",
		Last:  "Johnson",
	}

	joe := &Person{
		First: "Joe",
		Last:  "Johnson",
	}

	// here we're putting "bob" in the "people"
	// bucket under the key "bob"
	err = people.New(bob, &bob.First)
	if err != nil {
		panic(err)
	}

	// ... and we'll do the same with joe
	err = people.New(joe, &joe.First)
	if err != nil {
		panic(err)
	}

	// Now this is where things get more interesting.
	// We can retrieve both the "bob" and "joe" objects
	// asynchronously:
	results := people.MultiFetchAsync(&Person{}, 2, "Bob", "Joe")

	// and now we can iterate through
	// the results and print them out.
	for res := range results {
		// results return a "Value" field
		// and an "Error" field.
		if res.Error != nil {
			panic(err)
		}

		// res.Value can always be type-asserted
		// to the same type as the result returned
		// from NewEmpty()
		person := res.Value.(*Person)
		fmt.Printf("%s %s\n", person.First, person.Last)
	}

	// Here's another trick: let's make
	// people query-able by their last name.
	// We'll add a secondary index field called
	// "lastname" that contains the last name of the person.
	bob.Info().AddIndex("lastname", bob.Last)
	joe.Info().AddIndex("lastname", joe.Last)

	// now we need to push the changes
	// to those objects back to the database
	err = people.Push(bob)
	if err != nil {
		panic(err)
	}
	err = people.Push(joe)
	if err != nil {
		panic(err)
	}

	// now we can fetch all people
	// with the last name "Johnson":
	res, err := people.IndexLookup("lastname", "Johnson")
	if err != nil {
		panic(err)
	}

	// Riak only returns keys for secondary
	// index queries. However, rkive gives you
	// a quick way to fetch all of them that
	// looks a lot like the one we used before:
	stream := res.FetchAsync(&Person{}, 2)

	// ... and we can use it the same way:
	fmt.Print("\n")
	fmt.Println("All the Johnsons: ")
	fmt.Println("------------------")
	for v := range stream {
		person := v.Value.(*Person)
		fmt.Printf("%s %s\n", person.First, person.Last)
	}

	// now let's do some cleanup.
	riak.Delete(bob, nil)
	riak.Delete(joe, nil)
}
