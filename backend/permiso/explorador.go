package permiso

import (
	"backend/Herramientas"
	"backend/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"unicode"
)

type Entrada struct {
	Nombre string `json:"nombre"`
	Tipo   string `json:"tipo"` // "carpeta" o "archivo"
	Perm   string `json:"perm"`
	FechaCreacion string `json:"fechaCreacion"`
}

// Lee el contenido del inodo (por ejemplo el inodo 0 para "/")
func ListarContenidoInodo(disco *os.File, super Structs.Superblock, idInodo int32) []Entrada {
	var resultados []Entrada
	pos := int64(super.S_inode_start) + int64(idInodo)*int64(binary.Size(Structs.Inode{}))

	var inodo Structs.Inode
	if err := Herramientas.ReadObject(disco, &inodo, pos); err != nil {
		fmt.Println("Error leyendo inodo raíz:", err)
		return resultados
	}

	for i := 0; i < 12; i++ {
		if inodo.I_block[i] == -1 {
			continue
		}

		var carpeta Structs.Folderblock
		bloquePos := int64(super.S_block_start) + int64(inodo.I_block[i])*int64(binary.Size(Structs.Folderblock{}))
		if err := Herramientas.ReadObject(disco, &carpeta, bloquePos); err != nil {
			continue
		}

		for _, entry := range carpeta.B_content {
			nombre := Structs.GetName(string(entry.B_name[:]))
			if nombre == "" || nombre == "." || nombre == ".." {
				continue
			}
			subInodo := entry.B_inodo
			subPos := int64(super.S_inode_start) + int64(subInodo)*int64(binary.Size(Structs.Inode{}))
			var in Structs.Inode
			if err := Herramientas.ReadObject(disco, &in, subPos); err != nil {
				continue
			}
			tipo := "archivo"
			if in.I_type[0] == '0' {
				tipo = "carpeta"
			}
			//fmt.Println("Fecha de creación:", string(in.I_ctime[:])) // Verifica que la fecha esté correctamente formateada

			resultados = append(resultados, Entrada{
				Nombre: nombre,
				Tipo:   tipo,
				Perm:   string(in.I_perm[:]),
				FechaCreacion: string(in.I_ctime[:]),
			})
		}
	}
	fmt.Println("Contenido listado:", resultados)

	return resultados
}

func SearchPath(path string, disco *os.File, super Structs.Superblock) int32 {
	if path == "/" {
		return 0
	}

	current := int32(0)
	steps := strings.Split(path, "/")[1:]

	for _, step := range steps {
		step = strings.TrimSpace(step)
		step = CleanPath(step)

		found := false

		// Leer inodo actual
		var inodo Structs.Inode
		pos := int64(super.S_inode_start) + int64(current)*int64(binary.Size(Structs.Inode{}))
		if err := Herramientas.ReadObject(disco, &inodo, pos); err != nil {
			return -1
		}

		for i := 0; i < 12; i++ {
			if inodo.I_block[i] == -1 {
				continue
			}

			var folder Structs.Folderblock
			blockPos := int64(super.S_block_start) + int64(inodo.I_block[i])*int64(binary.Size(Structs.Folderblock{}))
			if err := Herramientas.ReadObject(disco, &folder, blockPos); err != nil {
				continue
			}

			for _, entry := range folder.B_content {
				nombre := Structs.GetB_name(string(entry.B_name[:])) 
				fmt.Printf("DEBUG SEARCHPATH - Revisando entrada: '%s'\n", Structs.GetB_name(string(entry.B_name[:])))
				fmt.Printf("Comparando con step: '%s'\n", step)

				if nombre == step {
					current = entry.B_inodo
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			return -1
		}
	}

	return current
}
func CleanPath(p string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) && r != '\u200B' && r != '\uFEFF' {
			return r
		}
		return -1
	}, p)
}

