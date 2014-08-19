package rkive

import (
	"errors"
)

var (
	ErrDone = errors.New("done")
)

// PushChangeset pushes a changeset to an object, handling the case
// in which the object has been updated in the database since the last
// local fetch. The 'chng' function should check if the change that it wanted
// already happened, and return ErrDone in that case. The 'chng' function is allowed
// to type-assert its argument to the underlying type of 'o'.
func (c *Client) PushChangset(o Object, chng func(Object) error, opts *WriteOpts) error {
	err := chng(o)
	if err != nil {
		return err
	}

	nmerge := 0
	for err = c.Push(o, opts); err != nil; nmerge++ {
		if nmerge > maxMerges {
			return err
		}
		if err == ErrModified {
			err = chng(o)
			if err != nil {
				if err == ErrDone {
					return nil
				}
				return err
			}
			continue
		}
		return err
	}
	return nil
}
