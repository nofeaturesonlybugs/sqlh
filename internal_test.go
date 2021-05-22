package sqlh

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScannerDestType(t *testing.T) {
	chk := assert.New(t)
	//
	chk.Equal("Invalid", scannerDestType(0).String())
}
