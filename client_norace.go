// +build race

package rkive

// finish node (success)
func (c *Client) done(n *conn) {
	n.Close()
}

// finish node (err)
func (c *Client) err(n *conn) {
	n.Close()
}
