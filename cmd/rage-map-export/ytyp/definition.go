package ytyp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/tgascoigne/ragekit/cmd/rage-map-export/ymap"
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type DefinitionHeader struct {
	Unk1    uint32
	Unk2    uint32
	Unk3Ptr types.Ptr32
	Unk4    uint32

	Unk4Ptr types.Ptr32
	Unk5    uint32
	Unk6    uint32
	Unk7    uint32

	Unk8Ptr  types.Ptr32
	Unk9     uint32
	Unk10Ptr uint32
	Unk11    uint32

	SectionListPtr types.Ptr32
	Unk13          uint32
	Unk14          uint32
	Unk15          uint32

	Unk16       types.Ptr32
	Unk17       uint32
	UnkCount18  uint16
	UnkCount19  uint16
	NumSections uint16
	Unk21       uint16
}

type Definition struct {
	Header      DefinitionHeader `json:"-"`
	FileName    string
	SectionList []ymap.SectionDef
	StringTable []byte
}

func NewDefinition(filename string) *Definition {
	return &Definition{
		FileName: filename,
	}
}

func (ytyp *Definition) Unpack(res *resource.Container, outpath string) error {
	res.Parse(&ytyp.Header)

	//	fmt.Printf("Header: %#v\n", ytyp.Header)

	/* parse the section table */
	err := res.Detour(ytyp.Header.SectionListPtr, func() error {
		count := ytyp.Header.NumSections
		ytyp.SectionList = make([]ymap.SectionDef, count)
		for i := 0; i < int(count); i++ {
			res.Parse(&ytyp.SectionList[i])
			//			fmt.Printf("ymap.SectionDef %#v\n", ytyp.SectionList[i])

			switch ytyp.SectionList[i].SectionType {

			case ymap.SectionSTRINGS:
				res.Detour(ytyp.SectionList[i].SectionPtr, func() error {
					//					fmt.Printf("String table: %x\n", res.Tell())
					length := int64(ytyp.SectionList[i].SizeBytes)
					ytyp.StringTable = make([]byte, length)
					copy(ytyp.StringTable, res.Data[res.Tell():res.Tell()+length])
					return nil
				})

			default:
				fmt.Printf("Unknown section type: %x\n", ytyp.SectionList[i].SectionType)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(ytyp, "", "\t")
	if err != nil {
		return err
	}

	fmt.Printf("Writing %v\n", outpath)
	err = ioutil.WriteFile(outpath, data, 0744)
	if err != nil {
		return err
	}

	return nil
}
