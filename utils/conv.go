package utils

import (
	"bytes"
	"encoding/binary"
	"math"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const SECTOR_SIZE = 0x800

func GetRequiredSectorsCount(size int64) int64 {
	return (size + SECTOR_SIZE - 1) / SECTOR_SIZE
}

func BytesToString(bs []byte) string {
	n := bytes.IndexByte(bs, 0)
	if n < 0 {
		n = len(bs)
	}

	s, _, err := transform.Bytes(charmap.Windows1252.NewDecoder(), bs[0:n])
	if err != nil {
		panic(err)
	}
	return string(s)
}

func BytesStringLength(bs []byte) int {
	if l := bytes.IndexByte(bs, 0); l == -1 {
		return len(bs)
	} else {
		return l
	}
}

func StringToBytesBuffer(s string, bufSize int, nilTerminate bool) []byte {
	bs, _, err := transform.Bytes(charmap.Windows1252.NewEncoder(), []byte(s))
	if err != nil {
		panic(err)
	}
	if nilTerminate {
		bs = append(bs, 0)
	}
	if len(bs) < bufSize {
		r := make([]byte, bufSize)
		copy(r, bs)
		bs = r
	} else if len(bs) > bufSize {
		panic(bs)
	}
	return bs
}

func StringToBytes(s string, nilTerminate bool) []byte {
	bs, _, err := transform.Bytes(charmap.Windows1252.NewEncoder(), []byte(s))
	if err != nil {
		panic(err)
	}
	if nilTerminate {
		bs = append(bs, 0)
	}
	return bs
}

func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func ReverseBytes(a []byte) []byte {
	r := make([]byte, len(a))
	j := len(r)
	for _, b := range a {
		j--
		r[j] = b
	}
	return r
}

func ReadBytes(out interface{}, raw []byte) {
	if err := binary.Read(bytes.NewReader(raw), binary.LittleEndian, out); err != nil {
		panic(err)
	}
}

func Float32FromFloat16bits(val uint16) float32 {
	sign := (val >> 15) & 1
	exp := int16((val >> 10) & 0x1f)
	frac := val & 0x3ff
	if exp == 0x1f {
		if frac != 0 {
			return float32(math.NaN())
		} else {
			return float32(math.Inf(int(sign)*-2 + 1))
		}
	} else {
		var bits uint64
		bits |= uint64(sign) << 63
		bits |= uint64(frac) << 42
		if exp > 0 {
			bits |= uint64(exp-15+1023) << 52
		}
		return float32(math.Float64frombits(bits))
	}
}
