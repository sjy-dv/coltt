package backup

import (
	"encoding/binary"
	"io"
)

const (
	zero bit = false
	one  bit = true
)

type bitstream struct {
	stream []byte
	count  uint8
}

func (self *bitstream) bytes() []byte {
	return self.stream
}

func (self *bitstream) reset() {
	self.stream = self.stream[:0]
	self.count = 0
}

type bit bool

func (self *bitstream) writeByte(b byte) {
	if self.count == 0 {
		self.stream = append(self.stream, 0)
		self.count = 8
	}

	i := len(self.stream) - 1

	self.stream[i] |= b >> (8 - self.count)

	self.stream = append(self.stream, 0)
	i++
	self.stream[i] = b << self.count
}

func (self *bitstream) writeBit(bit bit) {
	if self.count == 0 {
		self.stream = append(self.stream, 0)
		self.count = 8
	}

	i := len(self.stream) - 1

	if bit {
		self.stream[i] |= 1 << (self.count - 1)
	}

	self.count--
}

func (self *bitstream) writeBits(x uint64, nbits int) {
	x <<= (64 - uint(nbits))
	for nbits >= 8 {
		b := byte(x >> 56)
		self.writeByte(b)
		x <<= 8
		nbits -= 8
	}

	for nbits > 0 {
		self.writeBit((x >> 63) == 1)
		x <<= 1
		nbits--
	}
}

type bitstreamReader struct {
	stream       []byte
	streamOffset int
	buffer       uint64
	valid        uint8
}

func newBitReader(b []byte) bitstreamReader {
	return bitstreamReader{
		stream: b,
	}
}

func (self *bitstreamReader) readBit() (bit, error) {
	if self.valid == 0 {
		if !self.loadNextBuffer(1) {
			return false, io.EOF
		}
	}
	return self.readBitFast()
}

func (self *bitstreamReader) readBitFast() (bit, error) {
	if self.valid == 0 {
		return false, io.EOF
	}

	self.valid--
	bitmask := uint64(1) << self.valid
	return (self.buffer & bitmask) != 0, nil
}

func (self *bitstreamReader) readBits(nbits uint8) (uint64, error) {
	if self.valid == 0 {
		if !self.loadNextBuffer(nbits) {
			return 0, io.EOF
		}
	}

	if nbits <= self.valid {
		return self.readBitsFast(nbits)
	}

	// We have to read all remaining valid bits from the current buffer and a part from the next one.
	bitmask := (uint64(1) << self.valid) - 1
	nbits -= self.valid
	v := (self.buffer & bitmask) << nbits
	self.valid = 0

	if !self.loadNextBuffer(nbits) {
		return 0, io.EOF
	}

	bitmask = (uint64(1) << nbits) - 1
	v = v | ((self.buffer >> (self.valid - nbits)) & bitmask)
	self.valid -= nbits

	return v, nil
}

func (self *bitstreamReader) readBitsFast(nbits uint8) (uint64, error) {
	if nbits > self.valid {
		return 0, io.EOF
	}

	bitmask := (uint64(1) << nbits) - 1
	self.valid -= nbits

	return (self.buffer >> self.valid) & bitmask, nil
}

func (self *bitstreamReader) ReadByte() (byte, error) {
	v, err := self.readBits(8)
	if err != nil {
		return 0, err
	}
	return byte(v), nil
}

func (self *bitstreamReader) loadNextBuffer(nbits uint8) bool {
	if self.streamOffset >= len(self.stream) {
		return false
	}

	if self.streamOffset+8 < len(self.stream) {
		self.buffer = binary.BigEndian.Uint64(self.stream[self.streamOffset:])
		self.streamOffset += 8
		self.valid = 64
		return true
	}

	nbytes := int((nbits / 8) + 1)
	if self.streamOffset+nbytes > len(self.stream) {
		nbytes = len(self.stream) - self.streamOffset
	}

	buffer := uint64(0)
	for i := 0; i < nbytes; i++ {
		buffer = buffer | (uint64(self.stream[self.streamOffset+i]) << uint(8*(nbytes-i-1)))
	}
	self.buffer = buffer
	self.streamOffset += nbytes
	self.valid = uint8(nbytes * 8)
	return true
}
