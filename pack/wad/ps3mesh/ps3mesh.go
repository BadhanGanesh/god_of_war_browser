package ps3mesh

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const (
	PS3_MESH_MAGIC = 0x0003000f
)

type Meta struct {
	Boundaries [6]float32
}

type Object struct {
	StreamsCount uint32
	Metas        []Meta
	Indexes      []uint32
	POS0         []float32 // ID2 4 elements of float32
	BONI         []uint8   // ID4 4 elements of byte
	TEX0         []float32 // ID5 2 elements of uint16 converted to float32
	COL0         []float32 // ID6 3 elements of uint16 converted to float32
	NRM0         []uint8   // ID9 4 elements of byte
}

type Mesh struct {
	Objects []Object
}

func (o *Object) parseStream(header []byte, data []byte) error {
	magic := binary.BigEndian.Uint32(header[0:])
	id := binary.BigEndian.Uint16(header[4:])
	errInvalidIdMagic := fmt.Errorf("Invalid magic 0x%x for stream id 0x%x", magic, id)
	switch id {
	case 2:
		if magic != 0x504F5330 { // POS0
			return errInvalidIdMagic
		}
		o.POS0 = make([]float32, len(data)/4)
		for i := range o.POS0 {
			o.POS0[i] = math.Float32frombits(binary.BigEndian.Uint32(data[i*4:]))
		}
	case 4:
		if magic != 0x424F4E49 { // BONI
			return errInvalidIdMagic
		}
		o.BONI = make([]byte, len(data))
		copy(o.BONI, data)
	case 5:
		if magic != 0x54455830 { // TEX0
			// return errInvalidIdMagic
			return nil
		}
		o.TEX0 = make([]float32, len(data)/2)
		for i := range o.TEX0 {
			o.TEX0[i] = utils.Float32FromFloat16bits(binary.BigEndian.Uint16(data[i*2:]))
		}
	case 6:
		if magic != 0x434F4C30 { // COL0
			// return errInvalidIdMagic
			return nil
		}
		o.COL0 = make([]float32, len(data)/2)
		for i := range o.COL0 {
			o.COL0[i] = utils.Float32FromFloat16bits(binary.BigEndian.Uint16(data[i*2:]))
		}
	case 9:
		if magic != 0x4E524D30 { // NRM0
			return errInvalidIdMagic
		}
		o.NRM0 = make([]byte, len(data))
		copy(o.NRM0, data)
	}
	return nil
}

func (o *Object) parseStreams(b []byte) error {
	o.StreamsCount = binary.BigEndian.Uint32(b[0:])
	for i := uint32(0); i < o.StreamsCount; i++ {
		streamHeader := b[4+i*0x10:]
		streamDataOff := binary.BigEndian.Uint32(streamHeader[8:])
		streamDataSize := binary.BigEndian.Uint32(streamHeader[0xc:])
		streamData := b[streamDataOff : streamDataOff+streamDataSize]
		if err := o.parseStream(streamHeader, streamData); err != nil {
			return fmt.Errorf("Error when parsing stream 0x%x: %v", i, err)
		}
	}
	return nil
}

func (m *Meta) parseMeta(b []byte) error {
	for i := range m.Boundaries {
		m.Boundaries[i] = math.Float32frombits(binary.BigEndian.Uint32(b[i*4:]))
		//utils.LogDump(" OBJECT META ", b[4*6:0x33])
	}
	return nil
}

func (o *Object) Parse(b []byte) error {
	if header_magic := binary.BigEndian.Uint32(b[0:]); header_magic != 0x4d4f444c /* MODL */ {
		return fmt.Errorf("Invalid object header: %x", header_magic)
	}

	streamsSize := binary.BigEndian.Uint32(b[8:])

	if err := o.parseStreams(b[0xc : 0xc+streamsSize]); err != nil {
		return fmt.Errorf("Error when parsing  mdl streams: %v", err)
	}

	indexesCount := binary.BigEndian.Uint32(b[0xc+streamsSize:])
	indexesRawArray := b[0xc+streamsSize+4:]
	o.Indexes = make([]uint32, indexesCount)
	for i := range o.Indexes {
		o.Indexes[i] = binary.BigEndian.Uint32(indexesRawArray[i*4:])
	}

	metasRaws := b[0xc+streamsSize+4+indexesCount*4:]
	o.Metas = make([]Meta, binary.BigEndian.Uint32(metasRaws[4:]))
	for i := range o.Metas {
		if err := o.Metas[i].parseMeta(metasRaws[8+i*0x33 : 8+i*0x33+0x33]); err != nil {
			return fmt.Errorf("Error parsing meta 0x%x: %v", i, err)
		}
	}
	return nil
}

func (m *Mesh) Parse(b []byte) error {
	/*
		+00 uint32 - magic for Server loader
		+04 uint32 - magic "GMDL"
		+08 uint32 -
		+0c uint32 -
		+10 uint32 -
		+14 uint32 - objects count
		+18 uint32 -
		+1c uint32 - offsets to objects [ignore magic for server]
	*/

	m.Objects = make([]Object, binary.BigEndian.Uint32(b[0x14:]))
	for i := range m.Objects {
		objectStart := 4 + binary.BigEndian.Uint32(b[0x1c+i*4:])
		objectEnd := uint32(len(b))
		if i != len(m.Objects)-1 {
			objectEnd = 4 + binary.BigEndian.Uint32(b[0x1c+i*4+4:])
		}

		if err := m.Objects[i].Parse(b[objectStart:objectEnd]); err != nil {
			return fmt.Errorf("Error parsing object #%d: %v", i, err)
		}
	}

	return nil
}

func NewFromData(b []byte) (*Mesh, error) {
	m := &Mesh{}
	if err := m.Parse(b); err != nil {
		return nil, err
	} else {
		return m, nil
	}
}

func (m *Mesh) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return m, nil
}

func init() {
	wad.SetHandler(PS3_MESH_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	})
}
