package cityhash

import (
	"encoding/binary"
)

type Uint128 [2]uint64

func (ui128 *Uint128) SetLower64(l uint64) {
	ui128[0] = l
}

func (ui128 *Uint128) SetHigher64(h uint64) {
	ui128[1] = h
}

func (ui128 Uint128) Lower64() uint64 {
	return ui128[0]
}

func (ui128 Uint128) Higher64() uint64 {
	return ui128[1]
}

func (ui128 Uint128) Bytes() []byte {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, ui128[0])
	binary.LittleEndian.PutUint64(b[8:], ui128[1])
	return b
}

// Hash 128 input bits down to 64 bits of output.
// This is intended to be a reasonably good hash function.
func hash128to64(x Uint128) uint64 {
	// Murmur-inspired hashing.
	const kMul uint64 = 0x9ddfea08eb382d69
	var a uint64 = (x.Lower64() ^ x.Higher64()) * kMul
	a ^= (a >> 47)
	var b uint64 = (x.Higher64() ^ a) * kMul
	b ^= (b >> 47)
	b *= kMul
	return b
}

func swap(a, b *uint64) {
	*a, *b = *b, *a
}

func unalignedLoad64(p []byte) uint64 {
	return binary.LittleEndian.Uint64(p)
}

func unalignedLoad32(p []byte) uint32 {
	return binary.LittleEndian.Uint32(p)
}

// Some primes between 2^63 and 2^64 for various uses.
const (
	k0 uint64 = 0xc3a5c85c97cb3127
	k1 uint64 = 0xb492b66fbe98f273
	k2 uint64 = 0x9ae16a3b2f90404f
	k3 uint64 = 0xc949d7c7509e6557
)

// Bitwise right rotate.
func rotate(val uint64, shift uint32) uint64 {
	// Avoid shifting by 64: doing so yields an undefined result.
	if shift != 0 {
		return ((val >> shift) | (val << (64 - shift)))
	}

	return val
}

// Equivalent to rotate(), but requires the second arg to be non-zero.
func rotateByAtLeast1(val uint64, shift uint32) uint64 {
	return ((val >> shift) | (val << (64 - shift)))
}

func shiftMix(val uint64) uint64 {
	return val ^ (val >> 47)
}

func hashLen16(u, v uint64) uint64 {
	return hash128to64(Uint128{u, v})
}

func hashLen0to16(s []byte, length uint32) uint64 {
	if length >= 8 {
		var a uint64 = unalignedLoad64(s)
		var b uint64 = unalignedLoad64(s[length-8:])
		return hashLen16(a, rotateByAtLeast1(b+uint64(length), length)) ^ b
	}
	if length >= 4 {
		var a uint64 = uint64(unalignedLoad32(s))
		return hashLen16(uint64(length)+(a<<3), uint64(unalignedLoad32(s[length-4:])))
	}
	if length > 0 {
		var a uint8 = uint8(s[0])
		var b uint8 = uint8(s[length>>1])
		var c uint8 = uint8(s[length-1])
		var y uint32 = uint32(a) + (uint32(b) << 8)
		var z uint32 = length + (uint32(c) << 2)
		return shiftMix(uint64(y)*k2^uint64(z)*k3) * k2
	}
	return k2
}

// This probably works well for 16-byte strings as well, but it may be overkill
// in that case.
func hashLen17to32(s []byte, length uint32) uint64 {
	var a uint64 = unalignedLoad64(s) * k1
	var b uint64 = unalignedLoad64(s[8:])
	var c uint64 = unalignedLoad64(s[length-8:]) * k2
	var d uint64 = unalignedLoad64(s[length-16:]) * k0
	return hashLen16(rotate(a-b, 43)+rotate(c, 30)+d, a+rotate(b^k3, 20)-c+uint64(length))
}

// Return a 16-byte hash for 48 bytes.  Quick and dirty.
// Callers do best to use "random-looking" values for a and b.
func weakHashLen32WithSeeds(w, x, y, z, a, b uint64) Uint128 {
	a += w
	b = rotate(b+a+z, 21)
	var c uint64 = a
	a += x
	a += y
	b += rotate(a, 44)
	return Uint128{a + z, b + c}
}

// Return a 16-byte hash for s[0] ... s[31], a, and b.  Quick and dirty.
func weakHashLen32WithSeeds_3(s []byte, a, b uint64) Uint128 {
	return weakHashLen32WithSeeds(unalignedLoad64(s), unalignedLoad64(s[8:]), unalignedLoad64(s[16:]), unalignedLoad64(s[24:]), a, b)
}

// Return an 8-byte hash for 33 to 64 bytes.
func hashLen33to64(s []byte, length uint32) uint64 {
	var z uint64 = unalignedLoad64(s[24:])
	var a uint64 = unalignedLoad64(s) + (uint64(length)+unalignedLoad64(s[length-16:]))*k0
	var b uint64 = rotate(a+z, 52)
	var c uint64 = rotate(a, 37)
	a += unalignedLoad64(s[8:])
	c += rotate(a, 7)
	a += unalignedLoad64(s[16:])
	var vf uint64 = a + z
	var vs uint64 = b + rotate(a, 31) + c
	a = unalignedLoad64(s[16:]) + unalignedLoad64(s[length-32:])
	z = unalignedLoad64(s[length-8:])
	b = rotate(a+z, 52)
	c = rotate(a, 37)
	a += unalignedLoad64(s[length-24:])
	c += rotate(a, 7)
	a += unalignedLoad64(s[length-16:])
	var wf uint64 = a + z
	var ws uint64 = b + rotate(a, 31) + c
	var r uint64 = shiftMix((vf+ws)*k2 + (wf+vs)*k0)
	return shiftMix(r*k0+vs) * k2
}

// Hash function for a byte array.
func CityHash64(s []byte, length uint32) uint64 {
	if length <= 32 {
		if length <= 16 {
			return hashLen0to16(s, length)
		} else {
			return hashLen17to32(s, length)
		}
	} else if length <= 64 {
		return hashLen33to64(s, length)
	}

	// For strings over 64 bytes we hash the end first, and then as we
	// loop we keep 56 bytes of state: v, w, x, y, and z.
	var x uint64 = unalignedLoad64(s)
	var y uint64 = unalignedLoad64(s[length-16:]) ^ k1
	var z uint64 = unalignedLoad64(s[length-56:]) ^ k0
	var v Uint128 = weakHashLen32WithSeeds_3(s[length-64:], uint64(length), y)
	var w Uint128 = weakHashLen32WithSeeds_3(s[length-32:], uint64(length)*k1, k0)
	z += shiftMix(v.Higher64()) * k1
	x = rotate(z+x, 39) * k1
	y = rotate(y, 33) * k1

	// Decrease length to the nearest multiple of 64, and operate on 64-byte chunks.
	length = (length - 1) & ^uint32(63)
	for {
		x = rotate(x+y+v.Lower64()+unalignedLoad64(s[16:]), 37) * k1
		y = rotate(y+v.Higher64()+unalignedLoad64(s[48:]), 42) * k1
		x ^= w.Higher64()
		y ^= v.Lower64()
		z = rotate(z^w.Lower64(), 33)
		v = weakHashLen32WithSeeds_3(s, v.Higher64()*k1, x+w.Lower64())
		w = weakHashLen32WithSeeds_3(s[32:], z+w.Higher64(), y)
		swap(&z, &x)
		s = s[64:]
		length -= 64
		if length == 0 {
			break
		}
	}
	return hashLen16(hashLen16(v.Lower64(), w.Lower64())+shiftMix(y)*k1+z, hashLen16(v.Higher64(), w.Higher64())+x)
}

// Hash function for a byte array.  For convenience, a 64-bit seed is also
// hashed into the result.
func CityHash64WithSeed(s []byte, length uint32, seed uint64) uint64 {
	return CityHash64WithSeeds(s, length, k2, seed)
}

// Hash function for a byte array.  For convenience, two seeds are also
// hashed into the result.
func CityHash64WithSeeds(s []byte, length uint32, seed0, seed1 uint64) uint64 {
	return hashLen16(CityHash64(s, length)-seed0, seed1)
}

// A subroutine for CityHash128().  Returns a decent 128-bit hash for strings
// of any length representable in an int.  Based on City and Murmur.
func cityMurmur(s []byte, length uint32, seed Uint128) Uint128 {
	var a uint64 = seed.Lower64()
	var b uint64 = seed.Higher64()
	var c uint64 = 0
	var d uint64 = 0
	var l int32 = int32(length) - 16
	if l <= 0 { // length <= 16
		c = b*k1 + hashLen0to16(s, length)
		if length >= 8 {
			d = rotate(a+unalignedLoad64(s), 32)
		} else {
			d = rotate(a+c, 32)
		}
	} else { // length > 16
		c = hashLen16(unalignedLoad64(s[length-8:])+k1, a)
		d = hashLen16(b+uint64(length), c+unalignedLoad64(s[length-16:]))
		a += d
		for {
			a ^= shiftMix(unalignedLoad64(s)*k1) * k1
			a *= k1
			b ^= a
			c ^= shiftMix(unalignedLoad64(s[8:])*k1) * k1
			c *= k1
			d ^= c
			s = s[16:]
			l -= 16

			if l <= 0 {
				break
			}
		}
	}
	a = hashLen16(a, c)
	b = hashLen16(d, b)
	return Uint128{a ^ b, hashLen16(b, a)}
}

// Hash function for a byte array.  For convenience, a 128-bit seed is also
// hashed into the result.
func CityHash128WithSeed(s []byte, length uint32, seed Uint128) Uint128 {
	if length < 128 {
		return cityMurmur(s, length, seed)
	}

	var orig_length uint32 = length
	var orig_s []byte = s

	// We expect length >= 128 to be the common case.  Keep 56 bytes of state:
	// v, w, x, y, and z.
	var v, w Uint128
	var x uint64 = seed.Lower64()
	var y uint64 = seed.Higher64()
	var z uint64 = uint64(length) * k1

	v.SetLower64(rotate(y^k1, 49)*k1 + unalignedLoad64(s))
	v.SetHigher64(rotate(v.Lower64(), 42)*k1 + unalignedLoad64(s[8:]))
	w.SetLower64(rotate(y+z, 35)*k1 + x)
	w.SetHigher64(rotate(x+unalignedLoad64(s[88:]), 53) * k1)

	// This is the same inner loop as CityHash64(), manually unrolled.
	for {
		x = rotate(x+y+v.Lower64()+unalignedLoad64(s[16:]), 37) * k1
		y = rotate(y+v.Higher64()+unalignedLoad64(s[48:]), 42) * k1
		x ^= w.Higher64()
		y += v.Lower64()
		z = rotate(z^w.Higher64(), 33)
		v = weakHashLen32WithSeeds_3(s, v.Higher64()*k1, x+w.Lower64())
		w = weakHashLen32WithSeeds_3(s[32:], z+w.Higher64(), y)
		swap(&z, &x)
		s = s[64:]
		x = rotate(x+y+v.Lower64()+unalignedLoad64(s[16:]), 37) * k1
		y = rotate(y+v.Higher64()+unalignedLoad64(s[48:]), 42) * k1
		x ^= w.Higher64()
		y += v.Lower64()
		z = rotate(z^w.Higher64(), 33)
		v = weakHashLen32WithSeeds_3(s, v.Higher64()*k1, x+w.Lower64())
		w = weakHashLen32WithSeeds_3(s[32:], z+w.Higher64(), y)
		swap(&z, &x)
		s = s[64:]
		length -= 128

		if length < 128 {
			break
		}
	}
	y += rotate(w.Lower64(), 37)*k0 + z
	x += rotate(v.Lower64()+z, 49) * k0
	// If 0 < length < 128, hash up to 4 chunks of 32 bytes each from the end of s.
	var tail_done uint32
	for tail_done = 0; tail_done < length; {
		tail_done += 32
		y = rotate(x+y, 42)*k0 + v.Higher64()
		w.SetLower64(w.Lower64() + unalignedLoad64(orig_s[orig_length-tail_done+16:]))
		x = rotate(x, 49)*k0 + w.Lower64()
		w.SetLower64(v.Lower64())
		v = weakHashLen32WithSeeds_3(orig_s[orig_length-tail_done:], v.Lower64(), v.Higher64())
	}
	// At this point our 48 bytes of state should contain more than
	// enough information for a strong 128-bit hash.  We use two
	// different 48-byte-to-8-byte hashes to get a 16-byte final result.
	x = hashLen16(x, v.Lower64())
	y = hashLen16(y, w.Lower64())
	return Uint128{hashLen16(x+v.Higher64(), w.Higher64()) + y,
		hashLen16(x+w.Higher64(), y+v.Higher64())}
}

// Hash function for a byte array.
func CityHash128(s []byte, length uint32) Uint128 {
	if length >= 16 {
		return CityHash128WithSeed(s[16:length], length-16, Uint128{unalignedLoad64(s) ^ k3, unalignedLoad64(s[8:length])})
	} else if length >= 8 {
		return CityHash128WithSeed([]byte(nil), 0, Uint128{unalignedLoad64(s) ^ (uint64(length) * k0), unalignedLoad64(s[length-8:]) ^ k1})
	} else {
		return CityHash128WithSeed(s, length, Uint128{k0, k1})
	}
}
