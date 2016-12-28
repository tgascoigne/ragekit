package script

import (
	"bufio"
	"compress/flate"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Native64 uint64

func (n Native64) unmangle(codesize uint32, index int) Native64 {
	rotateN := (uint64(codesize) + uint64(index)) % 64
	return (n << rotateN) | (n >> uint64(64-rotateN))
}

type Native32 uint32

type nativeTable map[string]nativeCategory

type nativeCategory map[string]NativeSpec

type NativeSpec struct {
	Jhash  string `json:"jhash"`
	Name   string `json:"name"`
	Params []struct {
		Name       string `json:"name"`
		Type       Type   `json:"-"`
		TypeString string `json:"type"`
	} `json:"params"`
	Results       Type   `json:"-"`
	ResultsString string `json:"results"`
}

type NativeDB struct {
	jsonTable nativeTable
	table     map[Native64]NativeSpec
}

func LoadNatives(path string) (*NativeDB, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	table := new(NativeDB)
	err = json.Unmarshal(data, &table.jsonTable)
	if err != nil {
		return nil, err
	}

	table.table = make(map[Native64]NativeSpec)

	for _, category := range table.jsonTable {
		for hashStr, entry := range category {
			hash, err := strconv.ParseUint(hashStr[2:], 16, 64)
			if err != nil {
				return nil, err
			}

			entry.Results = GetType(entry.ResultsString)
			for i := range entry.Params {
				entry.Params[i].Type = GetType(entry.Params[i].TypeString)
			}

			table.table[Native64(hash)] = entry
		}
	}

	return table, nil
}

func (t *NativeDB) LoadTranslations(path string) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}

	flateReader := flate.NewReader(file)
	scanner := bufio.NewScanner(flateReader)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")

		newHash, err := strconv.ParseUint(parts[0], 16, 64)
		if err != nil {
			return err
		}

		oldHash, err := strconv.ParseUint(parts[1], 16, 64)
		if err != nil {
			return err
		}

		// copy spec at oldHash into newHash
		if entry, ok := t.table[Native64(oldHash)]; ok {
			t.table[Native64(newHash)] = entry
		} else {
			//fmt.Printf("Missing source native for translation %x -> %x\n", oldHash, newHash)
		}
	}

	return nil
}

func (t *NativeDB) LookupNative(hash Native64) *NativeSpec {
	if entry, ok := t.table[hash]; ok {
		return &entry
	}

	return nil
}
