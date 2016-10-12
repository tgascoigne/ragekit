package resource

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/tgascoigne/ragekit/resource/crypto"
	"github.com/tgascoigne/ragekit/resource/types"
	"github.com/tgascoigne/ragekit/util/stack"
)

type EncryptionType uint32

const (
	EncNone   EncryptionType = 0x04E45504F
	EncAES    EncryptionType = 0x0ffffff9
	BlockSize uint32         = 512
)

type PackageHeader struct {
	Magic       uint32
	EntryCount  uint32
	NamesLength uint32
	Encryption  EncryptionType
}

type PackageDirEntry struct {
	NameOffset   uint32
	_            uint32 /* always 0x7FFFFF00 */
	EntriesIndex uint32
	EntriesCount uint32
}

func (p *PackageDirEntry) Name(pkg *Package) string {
	name, err := pkg.readName(p.NameOffset)
	if err != nil {
		panic(err)
	}
	return name
}

func (p *PackageDirEntry) Children(pkg *Package) []PackageNode {
	for i := uint32(0); i < p.EntriesCount; i++ {
		pkg.entriesVisited[p.EntriesIndex+i] = true
	}

	return pkg.entries[p.EntriesIndex : p.EntriesIndex+p.EntriesCount]
}

type PackageBlobEntry struct {
	NameOffset     uint16
	CompressedSize types.Uint24
	Offset         types.Uint24
	Size           uint32
	EncryptFlag    uint32
}

func (p *PackageBlobEntry) Name(pkg *Package) string {
	name, err := pkg.readName(uint32(p.NameOffset))
	if err != nil {
		panic(err)
	}
	return name
}

func (p *PackageBlobEntry) Data(pkg *Package) []byte {
	compressed := true
	size := p.CompressedSize.Uint32()
	if size == 0 {
		size = p.Size
		compressed = false
	}

	offset := p.Offset.Uint32()

	err := pkg.decryptBlocks(offset, size, p.Size, p.Name(pkg), p.EncryptFlag)

	if err != nil {
		panic(err)
	}

	data := pkg.readBlocks(offset, size)
	if compressed {
		data, err = deflate(data, p.Size)
		if err != nil {
			panic(err)
		}
	}
	return data
}

func deflate(compressed []byte, expectedSize uint32) ([]byte, error) {
	deflateReader := flate.NewReader(bytes.NewReader(compressed))
	deflated, err := ioutil.ReadAll(deflateReader)
	if err != nil {
		return nil, err
	}

	if uint32(len(deflated)) != expectedSize {
		return nil, errors.New("deflated size did not match expectation")
	}

	return deflated, nil
}

type PackageResourceEntry struct {
	NameOffset uint16
	SizeRaw    types.Uint24
	Offset     types.Uint24
	SysFlags   uint32
	GfxFlags   uint32
}

func (p *PackageResourceEntry) Name(pkg *Package) string {
	name, err := pkg.readName(uint32(p.NameOffset))
	if err != nil {
		panic(err)
	}
	return name
}

func (p *PackageResourceEntry) Size(pkg *Package) uint32 {
	size := p.SizeRaw.Uint32()

	if size == 0xFFFFFF {
		// size doesnt fit into 24 bits, so it's pushed into some kind of block header
		offset := types.Ptr32(p.Offset.Uint32() * BlockSize)
		pkg.Detour(offset, func() error {
			hdr := make([]byte, 16)
			pkg.Parse(hdr)
			size = (uint32(hdr[7]) << 0) |
				(uint32(hdr[14]) << 8) |
				(uint32(hdr[5]) << 16) |
				(uint32(hdr[2]) << 24)
			return nil
		})
	}

	return size
}

func (p *PackageResourceEntry) Data(pkg *Package) []byte {
	size := p.Size(pkg)
	offset := p.Offset.Uint32() & 0x7FFFFF
	content := pkg.readBlocks(offset, size)
	header := new(ContainerHeader)
	header.Magic = 0x37435352

	header.SysFlags = p.SysFlags
	header.GfxFlags = p.GfxFlags
	header.Version = ((p.GfxFlags & 0xF0000000) >> 28) | ((p.SysFlags & 0xF0000000) >> 24)

	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.LittleEndian, header)
	buffer.Write(content[16:])
	return buffer.Bytes()
}

func (pkg *Package) unpackHeader(data []byte, filename string, filesize uint32) error {
	reader := bytes.NewReader(data)

	pkg.jumpStack.Allocate(0xF)
	pkg.Data = data
	pkg.size = int64(len(data))
	pkg.filename = filename
	pkg.filesize = filesize

	header := &pkg.Header
	if err := binary.Read(reader, nativeEndian, header); err != nil {
		return err
	}

	if header.Magic == 0xF00 {
		return ErrInvalidResource
	}

	pkg.Seek(0x10, 0) // seek past the header

	return nil
}

func (pkg *Package) readName(offset uint32) (result string, err error) {
	ptr := pkg.namesPtr + types.Ptr32(offset)

	err = pkg.Detour(ptr, func() error {
		pkg.Parse(&result)
		return nil
	})

	return result, err
}

func (pkg *Package) decryptBlocks(index, count, uncompressedLength uint32, filename string, encryptFlags uint32) error {
	blockOffset := types.Ptr32(BlockSize * index)

	return pkg.Detour(blockOffset, func() error {
		if encryptFlags != 1 {
			return nil
		}

		if pkg.Header.Encryption == EncAES {
			err := pkg.Decrypt(pkg.cryptoCtx, count)
			if err != nil {
				return err
			}
		} else {
			err := pkg.DecryptNG(pkg.cryptoCtx, filename, count, uncompressedLength)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (pkg *Package) readBlocks(index, count uint32) []byte {
	blockOffset := types.Ptr32(BlockSize * index)
	blocks := make([]byte, count)

	err := pkg.Detour(blockOffset, func() error {
		pkg.Parse(blocks)
		return nil
	})

	if err != nil {
		panic(err)
	}

	return blocks
}

func (pkg *Package) parseEntry() (PackageNode, error) {
	var entryType uint32
	err := pkg.Detour(types.Ptr32(pkg.position+4), func() error {
		pkg.Parse(&entryType)
		return nil
	})

	if err != nil {
		return nil, err
	}

	switch {
	case entryType == 0x7FFFFF00:
		// directory entry

		dirEntry := new(PackageDirEntry)
		pkg.Parse(dirEntry)

		return dirEntry, nil

	case (entryType & 0x80000000) == 0:
		// blob entry

		blobEntry := new(PackageBlobEntry)
		pkg.Parse(blobEntry)

		return blobEntry, nil

	default:
		// resource entry

		resourceEntry := new(PackageResourceEntry)
		pkg.Parse(resourceEntry)

		return resourceEntry, nil
	}
	panic("not reachable")
}

func (pkg *Package) cryptoContext() (*crypto.Context, error) {
	keyDir := os.Getenv(CryptoKeyEnv)
	if keyDir == "" {
		keyDir = "."
	}
	keys, err := crypto.LoadKeysFromDir(keyDir)
	if err != nil {
		return nil, err
	}

	return crypto.NewContext(keys), nil
}

func (pkg *Package) decryptTOC(ctx *crypto.Context) error {
	entrySize := uint32(binary.Size(new(PackageDirEntry)))
	entriesTotalBytes := entrySize * pkg.Header.EntryCount

	pkg.entriesPtr = types.Ptr32(pkg.position)
	pkg.namesPtr = types.Ptr32(pkg.position) + types.Ptr32(entriesTotalBytes)
	//	pkg.blocksPtr = pkg.namesPtr + types.Ptr32(pkg.Header.NamesLength)
	//	fmt.Printf("block ptr is %v\n", pkg.blocksPtr)

	if pkg.Header.Encryption == EncNone {
		// nothing to do
	} else if pkg.Header.Encryption == EncAES {
		// AES
		err := pkg.Detour(pkg.entriesPtr, func() error {
			return pkg.Decrypt(ctx, entriesTotalBytes)
		})

		if err != nil {
			return err
		}

		err = pkg.Detour(pkg.namesPtr, func() error {
			return pkg.Decrypt(ctx, pkg.Header.NamesLength)
		})

		if err != nil {
			return err
		}

	} else {
		// NG
		err := pkg.Detour(pkg.entriesPtr, func() error {
			return pkg.DecryptPackageNG(ctx, entriesTotalBytes)
		})

		if err != nil {
			return err
		}

		err = pkg.Detour(pkg.namesPtr, func() error {
			return pkg.DecryptPackageNG(ctx, pkg.Header.NamesLength)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (pkg *Package) Decrypt(ctx *crypto.Context, length uint32) error {
	start, end := uint32(pkg.position), uint32(pkg.position)+length
	plaintext, err := ctx.DecryptAES(pkg.Data[start:end])
	if err != nil {
		return err
	}

	copy(pkg.Data[start:end], plaintext)
	return nil
}

func (pkg *Package) DecryptPackageNG(ctx *crypto.Context, length uint32) error {
	start, end := uint32(pkg.position), uint32(pkg.position)+length
	plaintext, err := ctx.DecryptNG(pkg.Data[start:end], pkg.filename, pkg.filesize)
	if err != nil {
		return err
	}

	copy(pkg.Data[start:end], plaintext)
	return nil
}

func (pkg *Package) DecryptNG(ctx *crypto.Context, filename string, length uint32, uncompressedLength uint32) error {
	start, end := uint32(pkg.position), uint32(pkg.position)+length
	plaintext, err := ctx.DecryptNG(pkg.Data[start:end], filename, uncompressedLength)
	if err != nil {
		return err
	}

	copy(pkg.Data[start:end], plaintext)
	return nil
}

func (pkg *Package) Parse(dest interface{}) {
	var err error
	switch dest.(type) {
	case *string:
		/* find NULL */
		var i int64
		buf := make([]byte, stringMax)
		err = ErrInvalidString
		for i = 0; i < stringMax; i++ {
			if pkg.Data[pkg.position+i] == 0 {
				err = nil
				break
			}
		}
		if err != nil {
			panic(err)
		}

		str := dest.(*string)
		copy(buf, pkg.Data[pkg.position:pkg.position+i])
		*str = string(buf[:i])
		pkg.position += i
	default:
		err = binary.Read(pkg, nativeEndian, dest)

		if err != nil {
			panic(err)
		}
	}
}

func (pkg *Package) Detour(addr types.Ptr32, callback func() error) error {
	if err := pkg.Jump(addr); err != nil {
		return err
	}

	defer pkg.Return()

	return callback()
}

func (pkg *Package) Read(p []byte) (int, error) {
	var read int64
	toRead := int64(len(p))

	if pkg.position+toRead > pkg.size {
		return 0, io.EOF
	}

	for read < toRead && pkg.position+read < pkg.size {
		p[read] = pkg.Data[pkg.position]
		pkg.position++
		read++
	}
	return int(read), nil
}

func (pkg *Package) Skip(offset int64) (int64, error) {
	return pkg.Seek(offset, 1)
}

func (pkg *Package) Seek(offset int64, whence int) (int64, error) {
	var err error
	var new_offset int64
	if whence == 0 {
		new_offset = offset
		//		fmt.Printf("seeking to %x\n", new_offset)
	} else if whence == 1 {
		new_offset = pkg.position + offset
	} else if whence == 2 {
		new_offset = pkg.size - offset - 1
	}

	if new_offset < 0 || new_offset >= pkg.size {
		err = io.EOF
	}

	pkg.position = new_offset
	return pkg.position, err
}

func (pkg *Package) Tell() int64 {
	return pkg.position
}

func (pkg *Package) Jump(offset types.Ptr32) error {
	position := pkg.Tell()
	pkg.jumpStack.Push(&stack.Item{position})
	_, err := pkg.Seek(int64(offset), 0)
	return err
}

func (pkg *Package) Return() error {
	position := pkg.jumpStack.Pop()
	_, err := pkg.Seek(position.Value.(int64), 0)
	//	fmt.Printf("returning to %x\n", position.Value)
	return err
}
