package resource

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/Microsoft/go-winio/wim/lzx"
	"github.com/tgascoigne/ragekit/resource/crypto"
	"github.com/tgascoigne/ragekit/resource/types"
	"github.com/tgascoigne/ragekit/util/stack"
)

const (
	resMagic1 = 0x52534337
	resMagic2 = 0x52534307
	baseSize  = 0x2000
	stringMax = 256
)

const (
	CryptoKeyEnv = "RAGEKIT_KEY_DIR"
)

var ErrInvalidResource error = errors.New("invalid resource")
var ErrInvalidString error = errors.New("invalid string")

type ContainerHeader struct {
	Magic    uint32
	Version  uint32
	SysFlags uint32
	GfxFlags uint32
}

func (c ContainerHeader) Type() uint8 {
	return uint8((c.Version >> 24) & 0xFF)
}

const (
	ResourceBN       = 0x2b // flate
	ResourceMap      = 0x02 // flate
	ResourceScript   = 0x0a // ng + flate
	ResourceTexture  = 0x0d // flate
	ResourceDrawable = 0xa5 // flate
	ResourceFrag     = 0xa2 // flate
	ResourceClip     = 0x2e // flate
	ResourceED       = 0x19 // flate
	ResourceVR       = 0x01
)

type Container struct {
	Header    ContainerHeader
	SysOffset int64
	GfxOffset int64

	jumpStack stack.Stack
	position  int64
	size      int64
	Data      []byte
}

func (res *Container) Decrypt(ctx *crypto.Context) error {
	plaintext, err := ctx.DecryptAES(res.Data[res.position:])
	if err != nil {
		return err
	}

	res.Data = append(res.Data[:res.position], plaintext...)
	res.size = int64(len(res.Data))
	return nil
}

func (res *Container) DecryptNG(ctx *crypto.Context, filename string, filesize uint32) error {
	// We need to remove the header (0x10) from the file size, as it's not encrypted
	plaintext, err := ctx.DecryptNG(res.Data[res.position:], filename, filesize)
	if err != nil {
		return err
	}

	res.Data = append(res.Data[:res.position], plaintext...)
	res.size = int64(len(res.Data))
	return nil
}

func (res *Container) Deflate() error {
	deflateReader := flate.NewReader(bytes.NewReader(res.Data[res.position:]))
	deflated, err := ioutil.ReadAll(deflateReader)
	if err != nil {
		return err
	}
	res.Data = append(res.Data[:res.position], deflated...)
	res.size = int64(len(res.Data))
	return nil
}

func (res *Container) DecompressLZX() error {
	lzxReader, err := lzx.NewReader(bytes.NewReader(res.Data[res.position:]), int(res.size-res.position))
	if err != nil {
		return err
	}
	decompressed, err := ioutil.ReadAll(lzxReader)
	if err != nil {
		return err
	}
	res.Data = append(res.Data[:res.position], decompressed...)
	res.size = int64(len(res.Data))
	return nil
}

func (res *Container) Unpack(data []byte, filename string, filesize uint32) error {
	reader := bytes.NewReader(data)

	res.Data = data
	res.size = int64(len(data))

	header := &res.Header
	if err := binary.Read(reader, binary.BigEndian, header); err != nil {
		return err
	}

	if header.Magic != resMagic1 && header.Magic != resMagic2 {
		return ErrInvalidResource
	}

	res.SysOffset = 0x10
	res.GfxOffset = res.SysOffset + int64(getPartitionSize(header.SysFlags))

	res.jumpStack.Allocate(0xF)
	res.Seek(0x50000000, 0) // seek to the start of the system partition
	res.Data = data
	res.size = int64(len(data))

	keyDir := os.Getenv(CryptoKeyEnv)
	if keyDir == "" {
		keyDir = "."
	}
	keys, err := crypto.LoadKeysFromDir(keyDir)
	if err != nil {
		panic(err)
	}

	ctx := crypto.NewContext(keys)

	if res.Header.Type() == ResourceScript {
		err = res.DecryptNG(ctx, filename, filesize)
		//err = res.Decrypt(ctx)
		if err != nil {
			fmt.Printf("Decrypt failed: %v\n", err)
		}
	}

	err = res.Deflate()
	if err != nil {
		fmt.Printf("Deflate failed: %v\n", err)
	}

	return nil
}

func (res *Container) Parse(dest interface{}) {
	var err error
	switch dest.(type) {
	case *string:
		/* find NULL */
		var i int64
		buf := make([]byte, stringMax)
		err = ErrInvalidString
		for i = 0; i < stringMax; i++ {
			if res.Data[res.position+i] == 0 {
				err = nil
				break
			}
		}
		if err != nil {
			panic(err)
		}

		str := dest.(*string)
		copy(buf, res.Data[res.position:res.position+i])
		*str = string(buf[:i])
		res.position += i
	default:
		err = binary.Read(res, nativeEndian, dest)

		if err != nil && err != io.EOF {
			panic(err)
		}
	}
}

func (res *Container) ParseBigEndian(dest interface{}) {
	var err error
	switch dest.(type) {
	case *string:
		/* find NULL */
		var i int64
		buf := make([]byte, stringMax)
		err = ErrInvalidString
		for i = 0; i < stringMax; i++ {
			if res.Data[res.position+i] == 0 {
				err = nil
				break
			}
		}
		if err != nil {
			panic(err)
		}

		str := dest.(*string)
		copy(buf, res.Data[res.position:res.position+i])
		*str = string(buf[:i])
		res.position += i
	default:
		err = binary.Read(res, binary.BigEndian, dest)

		if err != nil && err != io.EOF {
			panic(err)
		}
	}
}

func (res *Container) Detour(addr types.Ptr32, callback func() error) error {
	if err := res.Jump(addr); err != nil {
		return err
	}

	defer res.Return()

	return callback()
}

func (res *Container) Peek(addr types.Ptr32, data interface{}) error {
	if err := res.Jump(addr); err != nil {
		return err
	}

	if err := binary.Read(res, binary.BigEndian, data); err != nil {
		return err
	}

	if err := res.Return(); err != nil {
		return err
	}
	return nil
}

func (res *Container) PeekElem(addr types.Ptr32, element int, data interface{}) error {
	return res.Peek(addr+types.Ptr32(element*intDataSize(data)), data)
}

/* Container Util functions */
func (res *Container) Read(p []byte) (int, error) {
	var read int64
	toRead := int64(len(p))

	if res.position+toRead >= res.size {
		return 0, io.EOF
	}

	for read < toRead && res.position+read < res.size {
		p[read] = res.Data[res.position]
		res.position++
		read++
	}
	return int(read), nil
}

func (res *Container) Skip(offset int64) (int64, error) {
	return res.Seek(offset, 1)
}

func (res *Container) Seek(offset int64, whence int) (int64, error) {
	var err error
	var new_offset int64
	if whence == 0 {
		partition := (offset >> 24) & 0xFF
		part_offset := int64(offset & 0xFFFFFF)
		switch partition {
		case 0x50:
			new_offset = res.SysOffset + part_offset
		case 0x60:
			new_offset = res.GfxOffset + part_offset
		default:
			new_offset = offset
		}
		//		fmt.Printf("seeking to %x\n", new_offset)
	} else if whence == 1 {
		new_offset = res.position + offset
	} else if whence == 2 {
		new_offset = res.size - offset - 1
	}

	if new_offset < 0 || new_offset >= res.size {
		err = io.EOF
	}

	res.position = new_offset
	return res.position, err
}

func (res *Container) Tell() int64 {
	return res.position
}

func (res *Container) Jump(offset types.Ptr32) error {
	position := res.Tell()
	res.jumpStack.Push(&stack.Item{position})
	_, err := res.Seek(int64(offset), 0)
	return err
}

func (res *Container) Return() error {
	position := res.jumpStack.Pop()
	_, err := res.Seek(position.Value.(int64), 0)
	//	fmt.Printf("returning to %x\n", position.Value)
	return err
}

func getPartitionSize(flags uint32) uint32 {
	var base uint32 = baseSize << (flags & 0xF)
	var size uint32 = 0
	size += ((flags >> 17) & 0x7F)
	size += (((flags >> 11) & 0x3F) << 1)
	size += (((flags >> 7) & 0xF) << 2)
	size += (((flags >> 5) & 0x3) << 3)
	size += (((flags >> 4) & 0x1) << 4)
	size *= base
	for i := uint(0); i < 4; i++ {
		if (flags >> (24 + i) & 1) == 1 {
			size += (base >> (1 + i))
		}
	}
	return size
}
