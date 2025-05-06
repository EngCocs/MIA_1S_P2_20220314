package reportes

import (
	"backend/Herramientas"
	"backend/Structs"
	"backend/permiso"
	"backend/session"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RepFile(pathDisco string, pathFileLS string, pathReporte string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========FILE========\n")
	disco, err := os.Open(pathDisco)
	if err != nil {
		return "",fmt.Errorf("Error al abrir disco: %v", err)
	}
	defer disco.Close()

	// Leer MBR
	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		return "",fmt.Errorf("Error al leer MBR: %v", err)
	}

	// Obtener partición activa
	var particion Structs.Partition
	for _, p := range mbr.Partitions {
		if Structs.GetId(string(p.Id[:])) == session.PartitionID {
			particion = p
			break
		}
	}
	if particion.Size == 0 {
		return "",fmt.Errorf("Partición no encontrada")
	}

	// Leer superbloque
	var super Structs.Superblock
	if err := Herramientas.ReadObject(disco, &super, int64(particion.Start)); err != nil {
		return "",fmt.Errorf("Error al leer superbloque: %v", err)
	}

	// Buscar el inodo
	inodoId := permiso.SearchInode(0, pathFileLS, super, disco)
	if inodoId == 0 {
		return "",fmt.Errorf("No se encontró el archivo: %s", pathFileLS)
	}

	// Leer el inodo
	var inode Structs.Inode
	inodoPos := int64(super.S_inode_start) + int64(inodoId)*int64(binary.Size(Structs.Inode{}))
	if err := Herramientas.ReadObject(disco, &inode, inodoPos); err != nil {
		return "",fmt.Errorf("Error al leer inodo del archivo")
	}

	// Leer bloques directos
	content := ""
	for i := 0; i < 12; i++ {
		if inode.I_block[i] == -1 {
			continue
		}
		blockPos := int64(super.S_block_start) + int64(inode.I_block[i])*int64(binary.Size(Structs.Fileblock{}))
		var block Structs.Fileblock
		if err := Herramientas.ReadObject(disco, &block, blockPos); err != nil {
			continue
		}
		content += string(block.B_content[:])
	}

	// Leer apuntador indirecto simple (ap13)
	if inode.I_block[13] != -1 {
		ptrPos := int64(super.S_block_start) + int64(inode.I_block[13])*int64(binary.Size(Structs.Pointerblock{}))
		var ptrBlock Structs.Pointerblock
		if err := Herramientas.ReadObject(disco, &ptrBlock, ptrPos); err == nil {
			for _, bloque := range ptrBlock.B_pointers {
				if bloque == -1 {
					continue
				}
				blockPos := int64(super.S_block_start) + int64(bloque)*int64(binary.Size(Structs.Fileblock{}))
				var block Structs.Fileblock
				if err := Herramientas.ReadObject(disco, &block, blockPos); err != nil {
					continue
				}
				content += string(block.B_content[:])
			}
		}
	}

	// Limpiar contenido (eliminar bytes nulos y separar por líneas legibles)
contenidoLimpio := ""
bloques := strings.Split(content, "#") // separador para bloques si lo usas al escribir

for _, bloque := range bloques {
	limpio := Structs.GetB_content(bloque)
	if limpio != "" {
		contenidoLimpio += limpio + "\n"
	}
}


	// Crear carpeta destino si no existe
	dir := filepath.Dir(pathReporte)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("Error al crear carpeta del reporte: %v", err)
		}
	}	
	err = os.WriteFile(pathReporte, []byte(contenidoLimpio), 0644)
		
	if err != nil {
		return "",fmt.Errorf("Error al escribir archivo DOT del reporte FILE: %v", err)
	}
	salida.WriteString(fmt.Sprintf("Reporte FILE generado en:", pathReporte))
	return salida.String(), nil
}


