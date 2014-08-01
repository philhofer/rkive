package riakpb

import (
	"github.com/philhofer/riakpb/rpbc"
)

// GetBucketTypeProperties gets the bucket properties
// associated with a given bucket type.
// *NOTE* bucket types are a Riak 2.0 feature.
func (c *Client) GetBucketTypeProperties(typeName string) (*rpbc.RpbBucketProps, error) {
	req := &rpbc.RpbGetBucketTypeReq{}
	req.Type = ustr(typeName)
	res := &rpbc.RpbBucketProps{}
	_, err := c.req(req, 31, res)
	return res, err
}

// SetBucketTypeProperties sets the bucket properties
// associated with a given bucket type.
// *NOTE* bucket types are a Riak 2.0 feature.
func (c *Client) SetBucketTypeProperties(typeName string, props *rpbc.RpbBucketProps) error {
	req := &rpbc.RpbSetBucketReq{}
	req.Props = props
	_, err := c.req(req, 32, nil)
	return err
}
