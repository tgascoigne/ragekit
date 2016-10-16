package encyclopedia

import (
	"fmt"

	"github.com/tgascoigne/ragekit/jenkins"
	"github.com/tgascoigne/ragekit/resource/item"
)

type PlacementRecord struct {
	Type  item.SectionType
	Entry item.SectionEntry
}

func (r PlacementRecord) Label() string {
	return fmt.Sprintf("%s", r.Type)
}

func (r PlacementRecord) Properties() map[string]interface{} {
	props := make(map[string]interface{})
	for fieldHash, value := range r.Entry {
		name := jenkins.Jenkins32(fieldHash).AsPropertyName()
		props[name] = value
	}

	return props
}
