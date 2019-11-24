package cip

const ErrorInvalidNetwork Error = "Network must be tcp or tcp4"
const ErrorInvalidIPID Error = "IPID must be between 0x03 and 0xFE inclusive"

// Dial initiates a CIP connection to a listening CIP host. `network` must be
// either "tcp" or "tcp4" and will always use IPv4. `address` should be the
// the host device, generally with port 41794 (eg "10.0.0.1:41794")
func Dial(network, address string, ipid IPID) (*Conn, error) {
	if network != "tcp" && network != "tcp4" {
		return nil, ErrorInvalidNetwork
	}
	if ipid < 0x03 || ipid > 0xfe {
		return nil, ErrorInvalidIPID
	}

	return nil, nil
}
