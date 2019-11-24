package cip

// An Error is a basic type implementing the error interface intended for simple
// constant error values.
type Error string

// An IPID is an 8 bit value representing how a supplicant device relates to the
// host device. In most cases, a CIP host is IPID 02, while a CIP client is 03
// through FE, though sometimes these roles are not well defined.
type IPID uint8

// A Digital value is either high or low, analogous to a boolean.
type Digital bool

// An Analog value is a 16 bit numeric value, whose signedness is context-
// dependent.
type Analog uint16

// A Serial value is a series of bytes. In theory the platform only supports
// serial sequences up to 255 bytes. Serials are also unique in that they are
// defined by the platform to be ephemeral.
type Serial []byte

// A Join is a 16 bit 1-based index into an array of Digitals, Analogs, or
// Serials. All CIP values have a Join number.
type Join uint16

func (e Error) Error() string {
	return string(e)
}
