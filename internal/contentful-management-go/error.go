package client

import (
	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

type ErrorWithDetailsWithReasonsString struct {
	Reasons OptString `json:"reasons"`
}

func (s *ErrorWithDetailsWithReasonsString) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode ErrorWithDetailsWithReasonsString to nil")
	}

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "reasons":
			if err := func() error {
				if err := s.Reasons.Decode(d); err != nil {
					return err
				}

				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"type\"")
			}
		default:
			return d.Skip()
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "decode ErrorWithDetailsWithReasonsString")
	}

	return nil
}
