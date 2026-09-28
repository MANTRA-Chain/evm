package usbwallet

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// hidReplyDevice answers any request with the given HID frames.
type hidReplyDevice struct {
	frames []byte
}

func (d *hidReplyDevice) Write(p []byte) (int, error) {
	return len(p), nil
}

func (d *hidReplyDevice) Read(p []byte) (int, error) {
	n := copy(p, d.frames)
	d.frames = d.frames[n:]
	return n, nil
}

func TestLedgerExchangeShortReply(t *testing.T) {
	// a device announcing a reply too short for the status word used to panic
	for _, reply := range [][]byte{nil, {0x90}} {
		driver := &ledgerDriver{device: &hidReplyDevice{frames: hidFrames(reply)}}
		_, err := driver.ledgerVersion()
		require.ErrorIs(t, err, errLedgerReplyLacksStatusWord)
	}
}
