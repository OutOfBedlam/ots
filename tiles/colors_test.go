package tiles_test

import (
	"fmt"
	"testing"

	"github.com/OutOfBedlam/ots/tiles"
	"github.com/stretchr/testify/assert"
)

func TestColor(t *testing.T) {
	//c := tiles.RgbHex("#f44336")
	c := tiles.Red
	assert.NotNil(t, c)
	r, g, b, a := c.RGBA()
	fmt.Printf("%x %x %x %x\n", r, g, b, a)
	assert.Equal(t, uint32(0xf4), r)
	assert.Equal(t, uint32(0x43), g)
	assert.Equal(t, uint32(0x36), b)
	assert.Equal(t, uint32(0xff), a)
}
