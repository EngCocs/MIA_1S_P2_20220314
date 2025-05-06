package session

import (
	"backend/Herramientas"
	"encoding/binary"
	"backend/Structs"
	"fmt"
	"os"
)

// WriteContentToFile escribe contenido en múltiples bloques (directos e indirectos) de un inodo
func WriteContentToFile(disco *os.File, super *Structs.Superblock, inode *Structs.Inode, content string) error {
	blockSize := int(binary.Size(Structs.Fileblock{}))
	contentBytes := []byte(content)
	totalLen := len(contentBytes)
	numBlocksNeeded := (totalLen + blockSize - 1) / blockSize

	maxDirect := 12
	maxSimple := 16
	maxDouble := 16 * 16
	maxTriple := 16 * 16 * 16
	totalMax := maxDirect + maxSimple + maxDouble + maxTriple

	if numBlocksNeeded > totalMax {
		return fmt.Errorf("el archivo es demasiado grande (límite: %d bloques)", totalMax)
	}

	for i := 0; i < numBlocksNeeded; i++ {
		start := i * blockSize
		end := start + blockSize
		if end > totalLen {
			end = totalLen
		}
		chunk := contentBytes[start:end]

		var blockIndex int32 = -1
		switch {
		case i < maxDirect:
			if inode.I_block[i] == -1 {
				b, err := ObtenerBloqueLibre(disco, super)
				if err != nil {
					return err
				}
				inode.I_block[i] = b
			}
			blockIndex = inode.I_block[i]

		case i < maxDirect+maxSimple:
			idx := i - maxDirect
			if inode.I_block[12] == -1 {
				b, err := ObtenerBloqueLibre(disco, super)
				if err != nil {
					return err
				}
				inode.I_block[12] = b
				pb := Structs.Pointerblock{}
				for j := range pb.B_pointers {
					pb.B_pointers[j] = -1
				}
				Herramientas.WriteObject(disco, pb, int64(super.S_block_start)+int64(b)*int64(blockSize))
			}
			ptrPos := int64(super.S_block_start) + int64(inode.I_block[12])*int64(blockSize)
			pb := Structs.Pointerblock{}
			Herramientas.ReadObject(disco, &pb, ptrPos)
			if pb.B_pointers[idx] == -1 {
				b, err := ObtenerBloqueLibre(disco, super)
				if err != nil {
					return err
				}
				pb.B_pointers[idx] = b
				Herramientas.WriteObject(disco, pb, ptrPos)
			}
			blockIndex = pb.B_pointers[idx]

		case i < maxDirect+maxSimple+maxDouble:
			idx := i - maxDirect - maxSimple
			outer := idx / 16
			inner := idx % 16
			if inode.I_block[13] == -1 {
				b, err := ObtenerBloqueLibre(disco, super)
				if err != nil {
					return err
				}
				inode.I_block[13] = b
				outerPB := Structs.Pointerblock{}
				for j := range outerPB.B_pointers {
					outerPB.B_pointers[j] = -1
				}
				Herramientas.WriteObject(disco, outerPB, int64(super.S_block_start)+int64(b)*int64(blockSize))
			}
			outerPos := int64(super.S_block_start) + int64(inode.I_block[13])*int64(blockSize)
			outerPB := Structs.Pointerblock{}
			Herramientas.ReadObject(disco, &outerPB, outerPos)
			if outerPB.B_pointers[outer] == -1 {
				b, err := ObtenerBloqueLibre(disco, super)
				if err != nil {
					return err
				}
				outerPB.B_pointers[outer] = b
				innerPB := Structs.Pointerblock{}
				for j := range innerPB.B_pointers {
					innerPB.B_pointers[j] = -1
				}
				Herramientas.WriteObject(disco, innerPB, int64(super.S_block_start)+int64(b)*int64(blockSize))
				Herramientas.WriteObject(disco, outerPB, outerPos)
			}
			innerPos := int64(super.S_block_start) + int64(outerPB.B_pointers[outer])*int64(blockSize)
			innerPB := Structs.Pointerblock{}
			Herramientas.ReadObject(disco, &innerPB, innerPos)
			if innerPB.B_pointers[inner] == -1 {
				b, err := ObtenerBloqueLibre(disco, super)
				if err != nil {
					return err
				}
				innerPB.B_pointers[inner] = b
				Herramientas.WriteObject(disco, innerPB, innerPos)
			}
			blockIndex = innerPB.B_pointers[inner]

		default:
			idx := i - maxDirect - maxSimple - maxDouble
			lvl1 := idx / (16 * 16)
			lvl2 := (idx / 16) % 16
			lvl3 := idx % 16
			if inode.I_block[14] == -1 {
				b, err := ObtenerBloqueLibre(disco, super)
				if err != nil {
					return err
				}
				inode.I_block[14] = b
				p1 := Structs.Pointerblock{}
				for j := range p1.B_pointers {
					p1.B_pointers[j] = -1
				}
				Herramientas.WriteObject(disco, p1, int64(super.S_block_start)+int64(b)*int64(blockSize))
			}
			lvl1Pos := int64(super.S_block_start) + int64(inode.I_block[14])*int64(blockSize)
			p1 := Structs.Pointerblock{}
			Herramientas.ReadObject(disco, &p1, lvl1Pos)
			if p1.B_pointers[lvl1] == -1 {
				b, err := ObtenerBloqueLibre(disco, super)
				if err != nil {
					return err
				}
				p1.B_pointers[lvl1] = b
				p2 := Structs.Pointerblock{}
				for j := range p2.B_pointers {
					p2.B_pointers[j] = -1
				}
				Herramientas.WriteObject(disco, p2, int64(super.S_block_start)+int64(b)*int64(blockSize))
				Herramientas.WriteObject(disco, p1, lvl1Pos)
			}
			lvl2Pos := int64(super.S_block_start) + int64(p1.B_pointers[lvl1])*int64(blockSize)
			p2 := Structs.Pointerblock{}
			Herramientas.ReadObject(disco, &p2, lvl2Pos)
			if p2.B_pointers[lvl2] == -1 {
				b, err := ObtenerBloqueLibre(disco, super)
				if err != nil {
					return err
				}
				p2.B_pointers[lvl2] = b
				p3 := Structs.Pointerblock{}
				for j := range p3.B_pointers {
					p3.B_pointers[j] = -1
				}
				Herramientas.WriteObject(disco, p3, int64(super.S_block_start)+int64(b)*int64(blockSize))
				Herramientas.WriteObject(disco, p2, lvl2Pos)
			}
			lvl3Pos := int64(super.S_block_start) + int64(p2.B_pointers[lvl2])*int64(blockSize)
			p3 := Structs.Pointerblock{}
			Herramientas.ReadObject(disco, &p3, lvl3Pos)
			if p3.B_pointers[lvl3] == -1 {
				b, err := ObtenerBloqueLibre(disco, super)
				if err != nil {
					return err
				}
				p3.B_pointers[lvl3] = b
				Herramientas.WriteObject(disco, p3, lvl3Pos)
			}
			blockIndex = p3.B_pointers[lvl3]
		}

		blockPos := int64(super.S_block_start) + int64(blockIndex)*int64(blockSize)
		var fileBlock Structs.Fileblock
		for i := range fileBlock.B_content {
			fileBlock.B_content[i] = 0
		}
		copy(fileBlock.B_content[:], chunk)

		if err := Herramientas.WriteObject(disco, fileBlock, blockPos); err != nil {
			return err
		}
	}

	inode.I_size = int32(totalLen)

	

	return nil
}

func ObtenerBloqueLibre(disco *os.File, super *Structs.Superblock) (int32, error) {
	bitmap := make([]byte, super.S_blocks_count)
	if err := Herramientas.ReadObject(disco, &bitmap, int64(super.S_bm_block_start)); err != nil {
		return -1, err
	}

	for i := int32(0); i < super.S_blocks_count; i++ {
		if bitmap[i] == 0 {
			bitmap[i] = 1
			super.S_free_blocks_count--

			//  Guardar bitmap actualizado
			if err := Herramientas.WriteObject(disco, bitmap, int64(super.S_bm_block_start)); err != nil {
				return -1, err
			}

			//  Guardar superbloque actualizado (posicion real)
			posSuper := int64(super.S_inode_start) - int64(binary.Size(*super))
			if err := Herramientas.WriteObject(disco, *super, posSuper); err != nil {
				return -1, err
			}

			fmt.Printf("DEBUG BLOQUE LIBRE ASIGNADO: %d\n", i)
			return i, nil
		}
	}

	return -1, fmt.Errorf("no hay bloques libres")
}


