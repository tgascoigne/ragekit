package script

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
)

type Native64 uint64
type Native32 uint32

type nativeTable map[string]nativeCategory

type nativeCategory map[string]NativeEntry

type NativeEntry struct {
	Jhash  string `json:"jhash"`
	Name   string `json:"name"`
	Params []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"params"`
	Results string `json:"results"`
}

type HashTable struct {
	jsonTable nativeTable
	table     map[Native64]NativeEntry
}

func LoadNatives(path string) (*HashTable, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	table := new(HashTable)
	err = json.Unmarshal(data, &table.jsonTable)
	if err != nil {
		return nil, err
	}

	table.table = make(map[Native64]NativeEntry)

	for _, category := range table.jsonTable {
		for hashStr, entry := range category {
			hash, err := strconv.ParseUint(hashStr[2:], 16, 64)
			if err != nil {
				return nil, err
			}

			table.table[Native64(hash)] = entry
		}
	}

	return table, nil
}

func (t *HashTable) LookupNative(hash Native64) *NativeEntry {
	if entry, ok := t.table[hash]; ok {
		return &entry
	}

	return nil
}
