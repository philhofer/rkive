// +build !race

package rkive

// finish node (success)
func (c *Client) done(n *conn) {
	if c.closed() {
		n.Close()
		return
	}
	c.pool.Put(n)
}

// finish node (err)
func (c *Client) err(n *conn) {
	if c.closed() {
		n.Close()
		return
	}
	err := ping(n)
	if err != nil {
		n.Close()
	}
	c.pool.Put(n)
}
