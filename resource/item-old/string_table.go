package item

import "encoding/json"

type StringTable []byte

func (table StringTable) MarshalJSON() ([]byte, error) {
	s := make(map[int]string)
	i := 0
	for i < len(table) {
		start := i
		for table[i] != 0 && i < len(table) {
			i++
		}
		str := string(table[start:i])
		i++
		s[start] = str
	}
	return json.Marshal(s)
}
