package rkive

import (
	"errors"
	"fmt"
)

var (
	ErrDone = errors.New("done")
)

// PushChangeset pushes a changeset to an object, handling the case
// in which the object has been updated in the database since the last
// local fetch. The 'chng' function should check if the change that it wanted
// already happened, and return ErrDone in that case. The 'chng' function is allowed
// to type-assert its argument to the underlying type of 'o'.
func (c *Client) PushChangeset(o Object, chng func(Object) error, opts *WriteOpts) error {
	err := chng(o)
	if err != nil {
		return err
	}
	nmerge := 0
push:
	err = c.Push(o, opts)
	if err == ErrModified {
		var upd bool
		nmerge++
		if nmerge > maxMerges {
			return fmt.Errorf("exceeded max merges: %s", err)
		}
		upd, err = c.Update(o, nil)
		if err != nil {
			return err
		}
		if !upd {
			return errors.New("failure updating...")
		}
		err = chng(o)
		if err == ErrDone {
			return nil
		}
		goto push
	}
	return err
}
