package crypto

import (
	"bytes"
	"embed"
	"encoding/binary"
	"fmt"
)

//go:embed res/gtav_aes_key.dat
//go:embed res/gtav_ng_key.dat
//go:embed res/gtav_ng_decrypt_tables.dat
//go:embed res/gtav_hash_lut.dat
var resFS embed.FS

const (
	AESKeyFile          = "res/gtav_aes_key.dat"
	NGKeyFile           = "res/gtav_ng_key.dat"
	NGDecryptTablesFile = "res/gtav_ng_decrypt_tables.dat"
	HashLookupFile      = "res/gtav_hash_lut.dat"
)

type Keys struct {
	aesKey         []byte
	ngKeys         [][]byte
	ngDecryptTable [][][]uint32
	hashLookup     []byte
}

func LoadKeys() (Keys, error) {
	var err error
	keys := Keys{}

	// Load AES key
	keys.aesKey, err = resFS.ReadFile(AESKeyFile)
	if err != nil {
		return keys, fmt.Errorf("reading AES key: %w", err)
	}

	// Load NG keys
	ngKeyBytes, err := resFS.ReadFile(NGKeyFile)
	if err != nil {
		return keys, fmt.Errorf("reading NG key: %w", err)
	}

	keys.ngKeys = make([][]byte, 101)
	for i := 0; i < 101; i++ {
		keys.ngKeys[i] = ngKeyBytes[:272]
		ngKeyBytes = ngKeyBytes[272:]
	}

	// Load NG decrypt tables
	ngTableBytes, err := resFS.ReadFile(NGDecryptTablesFile)
	if err != nil {
		return keys, fmt.Errorf("reading NG decrypt tables: %w", err)
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

	// Load hash lookup
	keys.hashLookup, err = resFS.ReadFile(HashLookupFile)
	if err != nil {
		return keys, fmt.Errorf("reading hash lookup: %w", err)
	}

	return keys, nil
}

func (k Keys) NgKeyForFile(filename string, length uint32) []byte {
	hash := CalculateHash(k, filename)
	keyIdx := (hash + (length) + (101 - 40)) % 0x65
	return k.ngKeys[keyIdx]
}
