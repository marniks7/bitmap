// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package bitmap

import (
	"math/bits"
)

// Bitmap represents a scalar-backed bitmap index
type Bitmap []uint64

// Set sets the bit x in the bitmap and grows it if necessary.
func (dst *Bitmap) Set(x uint32) {
	blkAt := int(x >> 6)
	bitAt := int(x % 64)
	if size := len(*dst); blkAt >= size {
		dst.grow(blkAt)
	}

	(*dst)[blkAt] |= (1 << bitAt)
}

// Remove removes the bit x from the bitmap, but does not shrink it.
func (dst *Bitmap) Remove(x uint32) {
	if blkAt := int(x >> 6); blkAt < len(*dst) {
		bitAt := int(x % 64)
		(*dst)[blkAt] &^= (1 << bitAt)
	}
}

// Contains checks whether a value is contained in the bitmap or not.
func (dst Bitmap) Contains(x uint32) bool {
	blkAt := int(x >> 6)
	if size := len(dst); blkAt >= size {
		return false
	}

	bitAt := int(x % 64)
	return (dst[blkAt] & (1 << bitAt)) > 0
}

// Ones sets the entire bitmap to one
func (dst Bitmap) Ones() {
	for i := 0; i < len(dst); i++ {
		dst[i] = 0xffffffffffffffff
	}
}

// Min get the smallest value stored in this bitmap, assuming the bitmap is not empty.
func (dst Bitmap) Min() (uint32, bool) {
	for blkAt, blk := range dst {
		if blk != 0x0 {
			return uint32(blkAt<<6 + bits.TrailingZeros64(blk)), true
		}
	}

	return 0, false
}

// Max get the largest value stored in this bitmap, assuming the bitmap is not empty.
func (dst Bitmap) Max() (uint32, bool) {
	var blk uint64
	for blkAt := len(dst) - 1; blkAt >= 0; blkAt-- {
		if blk = dst[blkAt]; blk != 0x0 {
			return uint32(blkAt<<6 + (63 - bits.LeadingZeros64(blk))), true
		}
	}
	return 0, false
}

// MinZero finds the first zero bit and returns its index, assuming the bitmap is not empty.
func (dst Bitmap) MinZero() (uint32, bool) {
	for blkAt, blk := range dst {
		if blk != 0xffffffffffffffff {
			return uint32(blkAt<<6 + bits.TrailingZeros64(^blk)), true
		}
	}
	return 0, false
}

// MaxZero get the last zero bit and return its index, assuming bitmap is not empty
func (dst Bitmap) MaxZero() (uint32, bool) {
	var blk uint64
	for blkAt := len(dst) - 1; blkAt >= 0; blkAt-- {
		if blk = dst[blkAt]; blk != 0xffffffffffffffff {
			return uint32(blkAt<<6 + (63 - bits.LeadingZeros64(^blk))), true
		}
	}
	return 0, false
}

// CountTo counts the number of elements in the bitmap up until the specified index. If until
// is math.MaxUint32, it will return the count. The count is non-inclusive of the index.
func (dst Bitmap) CountTo(until uint32) int {
	if len(dst) == 0 {
		return 0
	}

	// Figure out the index of the last block
	blkUntil := int(until >> 6)
	bitUntil := int(until % 64)
	if blkUntil >= len(dst) {
		blkUntil = len(dst) - 1
	}

	// Count the bits right before the last block
	sum := dst[:blkUntil].Count()

	// Count the bits at the end
	sum += bits.OnesCount64(dst[blkUntil] << (64 - uint64(bitUntil)))
	return sum
}

// Grow grows the bitmap size until we reach the desired bit.
func (dst *Bitmap) Grow(desiredBit uint32) {
	dst.grow(int(desiredBit >> 6))
}

// balance grows the destination bitmap to match the size of the source bitmap.
func (dst *Bitmap) balance(src Bitmap) {
	if len(*dst) == len(src) {
		return
	}

	// Grow if necessary
	if len(*dst) < len(src) {
		dst.grow(len(src) - 1)
		return
	}
}

// grow grows the size of the bitmap until we reach the desired block offset
func (dst *Bitmap) grow(blkAt int) {
	// Note that a bitmap is automatically initialized with zeros.

	// If blkAt is no greater that the current length, do nothing.
	if len(*dst) > blkAt {
		return
	}

	// If blkAt is no greater than the current capacity, resize the slice without copying.
	if cap(*dst) > blkAt {
		*dst = (*dst)[:blkAt+1]
		return
	}

	old := *dst
	*dst = make(Bitmap, blkAt+1, capacityFor(blkAt+1))
	copy(*dst, old)
}

// shrink shrinks the size of the bitmap and resets to zero
func (dst *Bitmap) shrink(length int) {
	until := len(*dst)
	for i := length; i < until; i++ {
		(*dst)[i] = 0
	}

	// Trim without reallocating
	*dst = (*dst)[:length]
}

// capacityFor computes the next power of 2 for a given index
func capacityFor(v int) int {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return int(v)
}
