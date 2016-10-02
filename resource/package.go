package resource

import (
	"github.com/tgascoigne/ragekit/resource/crypto"
	"github.com/tgascoigne/ragekit/resource/types"
	"github.com/tgascoigne/ragekit/util/stack"
)

type Package struct {
	Header PackageHeader

	filename string
	filesize uint32

	entriesPtr types.Ptr32
	namesPtr   types.Ptr32

	entries        []PackageNode
	entriesVisited []bool

	cryptoCtx *crypto.Context
	jumpStack stack.Stack
	position  int64
	size      int64
	Data      []byte
}

type PackageNode interface {
	Name(pkg *Package) string
}

type PackageDirectory interface {
	PackageNode
	Children(pkg *Package) []PackageNode
}

type PackageFile interface {
	PackageNode
	Data(pkg *Package) []byte
}

func (pkg *Package) UnvisitedEntries() []PackageNode {
	unvisited := make([]PackageNode, 0)
	for i, v := range pkg.entriesVisited {
		if !v {
			unvisited = append(unvisited, pkg.entries[i])
		}
	}
	return unvisited
}

func (pkg *Package) Unpack(data []byte, filename string, filesize uint32) error {
	err := pkg.unpackHeader(data, filename, filesize)
	if err != nil {
		return err
	}

	ctx, err := pkg.cryptoContext()
	if err != nil {
		return err
	}

	pkg.cryptoCtx = ctx

	err = pkg.decryptTOC(ctx)
	if err != nil {
		return err
	}

	pkg.entries = make([]PackageNode, pkg.Header.EntryCount)
	pkg.entriesVisited = make([]bool, pkg.Header.EntryCount)

	for i := uint32(0); i < pkg.Header.EntryCount; i++ {
		entry, err := pkg.parseEntry()
		if err != nil {
			return err
		}

		pkg.entries[i] = entry
	}

	return nil
}

func (pkg *Package) Root() PackageDirectory {
	pkg.entriesVisited[0] = true
	return pkg.entries[0].(PackageDirectory)
}
