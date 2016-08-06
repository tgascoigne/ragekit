package crypto

/* An implementation of the 'NG' cipher
   Ported to Go from Neodymium146's fantastic C# version (thanks!) */

import (
	"crypto/cipher"
	"encoding/binary"
)

const (
	NGBlockSize = 16
)

type NGCipher struct {
	ngKey []uint32
	table [][][]uint32
}

func NewNGCipher(ngKeyBytes []byte, decryptTable [][][]uint32) cipher.Block {
	ngKey := make([]uint32, len(ngKeyBytes)/4)
	for i := 0; i < len(ngKey); i++ {
		ngKey[i] = binary.LittleEndian.Uint32(ngKeyBytes[i*4 : (i+1)*4])
	}

	return &NGCipher{
		ngKey: ngKey,
		table: decryptTable,
	}
}

func (c *NGCipher) BlockSize() int {
	return NGBlockSize
}

func (c *NGCipher) Encrypt(dst []byte, src []byte) {
	panic("not implemented")
}

func (c *NGCipher) Decrypt(dst []byte, src []byte) {
	// prepare key...
	subkeys := make([][]uint32, 17)
	for i := 0; i < 17; i++ {
		subkeys[i] = make([]uint32, 4)
		subkeys[i][0] = c.ngKey[4*i+0]
		subkeys[i][1] = c.ngKey[4*i+1]
		subkeys[i][2] = c.ngKey[4*i+2]
		subkeys[i][3] = c.ngKey[4*i+3]
	}

	buffer := make([]byte, c.BlockSize())
	copy(buffer, src)

	buffer = c.decryptRoundA(buffer, subkeys[0], c.table[0])
	buffer = c.decryptRoundA(buffer, subkeys[1], c.table[1])
	for k := 2; k <= 15; k++ {
		buffer = c.decryptRoundB(buffer, subkeys[k], c.table[k])
	}
	buffer = c.decryptRoundA(buffer, subkeys[16], c.table[16])

	copy(dst, buffer)
}

// round 1,2,16
func (c *NGCipher) decryptRoundA(data []byte, key []uint32, table [][]uint32) []byte {
	x1 :=
		table[0][data[0]] ^
			table[1][data[1]] ^
			table[2][data[2]] ^
			table[3][data[3]] ^
			key[0]
	x2 :=
		table[4][data[4]] ^
			table[5][data[5]] ^
			table[6][data[6]] ^
			table[7][data[7]] ^
			key[1]
	x3 :=
		table[8][data[8]] ^
			table[9][data[9]] ^
			table[10][data[10]] ^
			table[11][data[11]] ^
			key[2]
	x4 :=
		table[12][data[12]] ^
			table[13][data[13]] ^
			table[14][data[14]] ^
			table[15][data[15]] ^
			key[3]

	result := make([]byte, 16)
	binary.LittleEndian.PutUint32(result[0:], x1)
	binary.LittleEndian.PutUint32(result[4:], x2)
	binary.LittleEndian.PutUint32(result[8:], x3)
	binary.LittleEndian.PutUint32(result[12:], x4)
	return result
}

// round 3-15
func (c *NGCipher) decryptRoundB(data []byte, key []uint32, table [][]uint32) []byte {
	x1 :=
		table[0][data[0]] ^
			table[7][data[7]] ^
			table[10][data[10]] ^
			table[13][data[13]] ^
			key[0]
	x2 :=
		table[1][data[1]] ^
			table[4][data[4]] ^
			table[11][data[11]] ^
			table[14][data[14]] ^
			key[1]
	x3 :=
		table[2][data[2]] ^
			table[5][data[5]] ^
			table[8][data[8]] ^
			table[15][data[15]] ^
			key[2]
	x4 :=
		table[3][data[3]] ^
			table[6][data[6]] ^
			table[9][data[9]] ^
			table[12][data[12]] ^
			key[3]

	result := make([]byte, 16)
	result[0] = byte((x1 >> 0) & 0xFF)
	result[1] = byte((x1 >> 8) & 0xFF)
	result[2] = byte((x1 >> 16) & 0xFF)
	result[3] = byte((x1 >> 24) & 0xFF)
	result[4] = byte((x2 >> 0) & 0xFF)
	result[5] = byte((x2 >> 8) & 0xFF)
	result[6] = byte((x2 >> 16) & 0xFF)
	result[7] = byte((x2 >> 24) & 0xFF)
	result[8] = byte((x3 >> 0) & 0xFF)
	result[9] = byte((x3 >> 8) & 0xFF)
	result[10] = byte((x3 >> 16) & 0xFF)
	result[11] = byte((x3 >> 24) & 0xFF)
	result[12] = byte((x4 >> 0) & 0xFF)
	result[13] = byte((x4 >> 8) & 0xFF)
	result[14] = byte((x4 >> 16) & 0xFF)
	result[15] = byte((x4 >> 24) & 0xFF)
	return result
}
