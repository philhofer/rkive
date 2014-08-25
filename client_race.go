// +build !race

package rkive

import "sync/atomic"

// finish node (success)
func (c *Client) done(n *conn) {
	if c.closed() {
		n.Close()
	} else {
		c.pool.Put(n)
	}
	atomic.AddInt32(&c.inuse, -1)
}

// finish node (err)
func (c *Client) err(n *conn) {
	if c.closed() {
		n.Close()
	} else {
		err := ping(n)
		if err != nil {
			n.Close()
		} else {
			c.pool.Put(n)
		}
	}
	atomic.AddInt32(&c.inuse, -1)
}
