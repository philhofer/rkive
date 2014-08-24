// +build race

package rkive

import "sync/atomic"

// finish node (success)
func (c *Client) done(n *conn) {
	n.Close()
	atomic.AddInt32(&c.inuse, -1)
}

// finish node (err)
func (c *Client) err(n *conn) {
	n.Close()
	atomic.AddInt32(&c.inuse, -1)
}
