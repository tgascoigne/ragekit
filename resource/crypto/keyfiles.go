package crypto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

const (
	AESKeyFile          = "gtav_aes_key.dat"
	NGKeyFile           = "gtav_ng_key.dat"
	NGDecryptTablesFile = "gtav_ng_decrypt_tables.dat"
	HashLookupFile      = "gtav_hash_lut.dat"
)

type Keys struct {
	aesKey         []byte
	ngKeys         [][]byte
	ngDecryptTable [][][]uint32
	hashLookup     []byte
}

func LoadKeysFromDir(dir string) (Keys, error) {
	var err error
	keys := Keys{}

	keys.aesKey, err = ioutil.ReadFile(fmt.Sprintf("%s/%s", dir, AESKeyFile))
	if err != nil {
		return keys, err
	}

	var ngKeyBytes []byte
	ngKeyBytes, err = ioutil.ReadFile(fmt.Sprintf("%s/%s", dir, NGKeyFile))
	if err != nil {
		return keys, err
	}

	keys.ngKeys = make([][]byte, 101)
	for i := 0; i < 101; i++ {
		keys.ngKeys[i] = ngKeyBytes[:272]
		ngKeyBytes = ngKeyBytes[272:]
	}

	var ngTableBytes []byte
	ngTableBytes, err = ioutil.ReadFile(fmt.Sprintf("%s/%s", dir, NGDecryptTablesFile))
	if err != nil {
		return keys, err
	}

	tableReader := bytes.NewReader(ngTableBytes)

	keys.ngDecryptTable = make([][][]uint32, 17)
	for i := 0; i < 17; i++ {
		keys.ngDecryptTable[i] = make([][]uint32, 16)
		for j := 0; j < 16; j++ {
			keys.ngDecryptTable[i][j] = make([]uint32, 256)
			binary.Read(tableReader, binary.LittleEndian, keys.ngDecryptTable[i][j])
		}
	}

	keys.hashLookup, err = ioutil.ReadFile(fmt.Sprintf("%s/%s", dir, HashLookupFile))
	if err != nil {
		return keys, err
	}

	return keys, nil
}

func (k Keys) NgKeyForFile(filename string, length uint32) []byte {
	hash := CalculateHash(k, filename)
	keyIdx := (hash + (length) + (101 - 40)) % 0x65
	return k.ngKeys[keyIdx]
}
