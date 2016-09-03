package item

import "encoding/json"

type Sections map[SectionType][]Section

func (s Sections) MarshalJSON() ([]byte, error) {
	// The key of maps aren't serialized using MarshalJSON, so we need to convert it to a map[String]
	m := make(map[string][]Section)
	for k, v := range s {
		m[k.String()] = v
	}
	return json.Marshal(m)
}

func (s Sections) Add(typ SectionType, section Section) {
	if _, ok := s[typ]; !ok {
		s[typ] = make([]Section, 0)
	}

	s[typ] = append(s[typ], section)
}
