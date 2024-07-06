// Code generated by ogen, DO NOT EDIT.

package client

import (
	"math/bits"
	"strconv"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/json"
	"github.com/ogen-go/ogen/validate"
)

// Encode implements json.Marshaler.
func (s *AppInstallation) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *AppInstallation) encodeFields(e *jx.Encoder) {
	{
		e.FieldStart("sys")
		s.Sys.Encode(e)
	}
	{
		if s.Parameters.Set {
			e.FieldStart("parameters")
			s.Parameters.Encode(e)
		}
	}
}

var jsonFieldsNameOfAppInstallation = [2]string{
	0: "sys",
	1: "parameters",
}

// Decode decodes AppInstallation from json.
func (s *AppInstallation) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode AppInstallation to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "sys":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				if err := s.Sys.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"sys\"")
			}
		case "parameters":
			if err := func() error {
				s.Parameters.Reset()
				if err := s.Parameters.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"parameters\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode AppInstallation")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00000001,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfAppInstallation) {
					name = jsonFieldsNameOfAppInstallation[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *AppInstallation) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *AppInstallation) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s AppInstallationParameters) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields implements json.Marshaler.
func (s AppInstallationParameters) encodeFields(e *jx.Encoder) {
	for k, elem := range s {
		e.FieldStart(k)

		if len(elem) != 0 {
			e.Raw(elem)
		}
	}
}

// Decode decodes AppInstallationParameters from json.
func (s *AppInstallationParameters) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode AppInstallationParameters to nil")
	}
	m := s.init()
	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		var elem jx.Raw
		if err := func() error {
			v, err := d.RawAppend(nil)
			elem = jx.Raw(v)
			if err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return errors.Wrapf(err, "decode field %q", k)
		}
		m[string(k)] = elem
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode AppInstallationParameters")
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s AppInstallationParameters) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *AppInstallationParameters) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *AppInstallationSys) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *AppInstallationSys) encodeFields(e *jx.Encoder) {
	{
		e.FieldStart("type")
		s.Type.Encode(e)
	}
}

var jsonFieldsNameOfAppInstallationSys = [1]string{
	0: "type",
}

// Decode decodes AppInstallationSys from json.
func (s *AppInstallationSys) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode AppInstallationSys to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "type":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				if err := s.Type.Decode(d); err != nil {
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
		return errors.Wrap(err, "decode AppInstallationSys")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00000001,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfAppInstallationSys) {
					name = jsonFieldsNameOfAppInstallationSys[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *AppInstallationSys) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *AppInstallationSys) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes AppInstallationSysType as json.
func (s AppInstallationSysType) Encode(e *jx.Encoder) {
	e.Str(string(s))
}

// Decode decodes AppInstallationSysType from json.
func (s *AppInstallationSysType) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode AppInstallationSysType to nil")
	}
	v, err := d.StrBytes()
	if err != nil {
		return err
	}
	// Try to use constant string.
	switch AppInstallationSysType(v) {
	case AppInstallationSysTypeAppInstallation:
		*s = AppInstallationSysTypeAppInstallation
	default:
		*s = AppInstallationSysType(v)
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s AppInstallationSysType) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *AppInstallationSysType) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *Error) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *Error) encodeFields(e *jx.Encoder) {
	{
		e.FieldStart("sys")
		s.Sys.Encode(e)
	}
	{
		e.FieldStart("message")
		e.Str(s.Message)
	}
	{
		if s.Details.Set {
			e.FieldStart("details")
			s.Details.Encode(e)
		}
	}
}

var jsonFieldsNameOfError = [3]string{
	0: "sys",
	1: "message",
	2: "details",
}

// Decode decodes Error from json.
func (s *Error) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode Error to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "sys":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				if err := s.Sys.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"sys\"")
			}
		case "message":
			requiredBitSet[0] |= 1 << 1
			if err := func() error {
				v, err := d.Str()
				s.Message = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"message\"")
			}
		case "details":
			if err := func() error {
				s.Details.Reset()
				if err := s.Details.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"details\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode Error")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00000011,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfError) {
					name = jsonFieldsNameOfError[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *Error) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *Error) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *ErrorSys) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *ErrorSys) encodeFields(e *jx.Encoder) {
	{
		e.FieldStart("type")
		s.Type.Encode(e)
	}
	{
		e.FieldStart("id")
		e.Str(s.ID)
	}
}

var jsonFieldsNameOfErrorSys = [2]string{
	0: "type",
	1: "id",
}

// Decode decodes ErrorSys from json.
func (s *ErrorSys) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode ErrorSys to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "type":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				if err := s.Type.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"type\"")
			}
		case "id":
			requiredBitSet[0] |= 1 << 1
			if err := func() error {
				v, err := d.Str()
				s.ID = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"id\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode ErrorSys")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00000011,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfErrorSys) {
					name = jsonFieldsNameOfErrorSys[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *ErrorSys) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *ErrorSys) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes ErrorSysType as json.
func (s ErrorSysType) Encode(e *jx.Encoder) {
	e.Str(string(s))
}

// Decode decodes ErrorSysType from json.
func (s *ErrorSysType) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode ErrorSysType to nil")
	}
	v, err := d.StrBytes()
	if err != nil {
		return err
	}
	// Try to use constant string.
	switch ErrorSysType(v) {
	case ErrorSysTypeError:
		*s = ErrorSysTypeError
	default:
		*s = ErrorSysType(v)
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s ErrorSysType) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *ErrorSysType) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes AppInstallationParameters as json.
func (o OptAppInstallationParameters) Encode(e *jx.Encoder) {
	if !o.Set {
		return
	}
	o.Value.Encode(e)
}

// Decode decodes AppInstallationParameters from json.
func (o *OptAppInstallationParameters) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptAppInstallationParameters to nil")
	}
	o.Set = true
	o.Value = make(AppInstallationParameters)
	if err := o.Value.Decode(d); err != nil {
		return err
	}
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s OptAppInstallationParameters) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *OptAppInstallationParameters) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes time.Time as json.
func (o OptDateTime) Encode(e *jx.Encoder, format func(*jx.Encoder, time.Time)) {
	if !o.Set {
		return
	}
	format(e, o.Value)
}

// Decode decodes time.Time from json.
func (o *OptDateTime) Decode(d *jx.Decoder, format func(*jx.Decoder) (time.Time, error)) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptDateTime to nil")
	}
	o.Set = true
	v, err := format(d)
	if err != nil {
		return err
	}
	o.Value = v
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s OptDateTime) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e, json.EncodeDateTime)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *OptDateTime) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d, json.DecodeDateTime)
}

// Encode encodes PutAppInstallationReqParameters as json.
func (o OptPutAppInstallationReqParameters) Encode(e *jx.Encoder) {
	if !o.Set {
		return
	}
	o.Value.Encode(e)
}

// Decode decodes PutAppInstallationReqParameters from json.
func (o *OptPutAppInstallationReqParameters) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptPutAppInstallationReqParameters to nil")
	}
	o.Set = true
	o.Value = make(PutAppInstallationReqParameters)
	if err := o.Value.Decode(d); err != nil {
		return err
	}
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s OptPutAppInstallationReqParameters) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *OptPutAppInstallationReqParameters) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes string as json.
func (o OptString) Encode(e *jx.Encoder) {
	if !o.Set {
		return
	}
	e.Str(string(o.Value))
}

// Decode decodes string from json.
func (o *OptString) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptString to nil")
	}
	o.Set = true
	v, err := d.Str()
	if err != nil {
		return err
	}
	o.Value = string(v)
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s OptString) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *OptString) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *PutAppInstallationReq) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *PutAppInstallationReq) encodeFields(e *jx.Encoder) {
	{
		if s.Parameters.Set {
			e.FieldStart("parameters")
			s.Parameters.Encode(e)
		}
	}
}

var jsonFieldsNameOfPutAppInstallationReq = [1]string{
	0: "parameters",
}

// Decode decodes PutAppInstallationReq from json.
func (s *PutAppInstallationReq) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode PutAppInstallationReq to nil")
	}

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "parameters":
			if err := func() error {
				s.Parameters.Reset()
				if err := s.Parameters.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"parameters\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode PutAppInstallationReq")
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *PutAppInstallationReq) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *PutAppInstallationReq) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s PutAppInstallationReqParameters) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields implements json.Marshaler.
func (s PutAppInstallationReqParameters) encodeFields(e *jx.Encoder) {
	for k, elem := range s {
		e.FieldStart(k)

		if len(elem) != 0 {
			e.Raw(elem)
		}
	}
}

// Decode decodes PutAppInstallationReqParameters from json.
func (s *PutAppInstallationReqParameters) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode PutAppInstallationReqParameters to nil")
	}
	m := s.init()
	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		var elem jx.Raw
		if err := func() error {
			v, err := d.RawAppend(nil)
			elem = jx.Raw(v)
			if err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return errors.Wrapf(err, "decode field %q", k)
		}
		m[string(k)] = elem
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode PutAppInstallationReqParameters")
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s PutAppInstallationReqParameters) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *PutAppInstallationReqParameters) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *User) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *User) encodeFields(e *jx.Encoder) {
	{
		e.FieldStart("sys")
		s.Sys.Encode(e)
	}
	{
		e.FieldStart("email")
		e.Str(s.Email)
	}
	{
		e.FieldStart("firstName")
		e.Str(s.FirstName)
	}
	{
		e.FieldStart("lastName")
		e.Str(s.LastName)
	}
}

var jsonFieldsNameOfUser = [4]string{
	0: "sys",
	1: "email",
	2: "firstName",
	3: "lastName",
}

// Decode decodes User from json.
func (s *User) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode User to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "sys":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				if err := s.Sys.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"sys\"")
			}
		case "email":
			requiredBitSet[0] |= 1 << 1
			if err := func() error {
				v, err := d.Str()
				s.Email = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"email\"")
			}
		case "firstName":
			requiredBitSet[0] |= 1 << 2
			if err := func() error {
				v, err := d.Str()
				s.FirstName = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"firstName\"")
			}
		case "lastName":
			requiredBitSet[0] |= 1 << 3
			if err := func() error {
				v, err := d.Str()
				s.LastName = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"lastName\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode User")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00001111,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfUser) {
					name = jsonFieldsNameOfUser[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *User) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *User) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *UserSys) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *UserSys) encodeFields(e *jx.Encoder) {
	{
		e.FieldStart("type")
		s.Type.Encode(e)
	}
	{
		e.FieldStart("id")
		e.Str(s.ID)
	}
	{
		e.FieldStart("version")
		e.Int(s.Version)
	}
	{
		if s.CreatedAt.Set {
			e.FieldStart("createdAt")
			s.CreatedAt.Encode(e, json.EncodeDateTime)
		}
	}
	{
		if s.UpdatedAt.Set {
			e.FieldStart("updatedAt")
			s.UpdatedAt.Encode(e, json.EncodeDateTime)
		}
	}
}

var jsonFieldsNameOfUserSys = [5]string{
	0: "type",
	1: "id",
	2: "version",
	3: "createdAt",
	4: "updatedAt",
}

// Decode decodes UserSys from json.
func (s *UserSys) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode UserSys to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "type":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				if err := s.Type.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"type\"")
			}
		case "id":
			requiredBitSet[0] |= 1 << 1
			if err := func() error {
				v, err := d.Str()
				s.ID = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"id\"")
			}
		case "version":
			requiredBitSet[0] |= 1 << 2
			if err := func() error {
				v, err := d.Int()
				s.Version = int(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"version\"")
			}
		case "createdAt":
			if err := func() error {
				s.CreatedAt.Reset()
				if err := s.CreatedAt.Decode(d, json.DecodeDateTime); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"createdAt\"")
			}
		case "updatedAt":
			if err := func() error {
				s.UpdatedAt.Reset()
				if err := s.UpdatedAt.Decode(d, json.DecodeDateTime); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"updatedAt\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode UserSys")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00000111,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfUserSys) {
					name = jsonFieldsNameOfUserSys[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *UserSys) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *UserSys) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes UserSysType as json.
func (s UserSysType) Encode(e *jx.Encoder) {
	e.Str(string(s))
}

// Decode decodes UserSysType from json.
func (s *UserSysType) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode UserSysType to nil")
	}
	v, err := d.StrBytes()
	if err != nil {
		return err
	}
	// Try to use constant string.
	switch UserSysType(v) {
	case UserSysTypeUser:
		*s = UserSysTypeUser
	default:
		*s = UserSysType(v)
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s UserSysType) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *UserSysType) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}
