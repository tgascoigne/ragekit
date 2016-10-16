package encyclopedia

import "github.com/tgascoigne/ragekit/jenkins"

type Asset struct {
	Hash jenkins.Jenkins32
}

func (a Asset) Label() string {
	return "Asset"
}

func (a Asset) Properties() map[string]interface{} {
	return map[string]interface{}{
		"name": a.Hash.String(),
		"hash": uint32(a.Hash),
	}
}
