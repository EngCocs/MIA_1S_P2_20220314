package administacioncarparch

import (
	"backend/Herramientas"
	"backend/Structs"
	"backend/permiso"
	"backend/session"
	"encoding/binary"
	"fmt"
	"strings"
)

func Cat(entrada []string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========CAT========\n")
	if !session.Active {
		salida.WriteString("CAT Error: No hay sesión activa.\n")
		return salida.String(), nil
	}

	// Obtener path del disco
	var pathDisco string
	for _, m := range Structs.Montadas {
		if m.Id == session.PartitionID {
			pathDisco = m.PathM
			break
		}
	}
	if pathDisco == "" {
		salida.WriteString("CAT Error: No se encontró la partición montada.\n")
		return salida.String(), nil
	}
	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		salida.WriteString(fmt.Sprintf("CAT Error al abrir el disco: %v\n", err))
		return salida.String(), err
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		salida.WriteString(fmt.Sprintf("CAT Error al leer MBR: %v\n", err))
		return salida.String(), err
	}

	var super Structs.Superblock
	for i := 0; i < 4; i++ {
		if Structs.GetId(string(mbr.Partitions[i].Id[:])) == session.PartitionID {
			Herramientas.ReadObject(disco, &super, int64(mbr.Partitions[i].Start))
			break
		}
	}

	// Recorrer archivos
	for _, param := range entrada[1:] {
		if !strings.HasPrefix(param, "-file")  {
			salida.WriteString(fmt.Sprintf("CAT Error: Parámetro inválido: %s\n", param))
			continue
		}
		partes := strings.SplitN(param, "=", 2)
		if len(partes) != 2 {
			salida.WriteString(fmt.Sprintf("CAT Error: Formato incorrecto para parámetro: %s\n", param))
			continue
		}
		path := strings.ReplaceAll(partes[1], "\"", "")
		path = strings.TrimSpace(path)

		// Buscar el inodo del archivo
		idInodo := permiso.SearchInode(0, path, super, disco)
		if idInodo == -1 {
			salida.WriteString(fmt.Sprintf("CAT Error: No se encontró el archivo %s\n", path))
			continue
		}

		// Leer inodo
		var inode Structs.Inode
		inodoPos := int64(super.S_inode_start) + int64(idInodo)*int64(binary.Size(Structs.Inode{}))
		if err := Herramientas.ReadObject(disco, &inode, inodoPos); err != nil {
			salida.WriteString(fmt.Sprintf("CAT Error: No se pudo leer inodo de %s\n", path))
			continue
		}
		fmt.Printf("DEBUG CAT -> Inodo leido en pos %d: bloques: %v\n", inodoPos, inode.I_block)

		fmt.Printf("Tipo en %s: %v - como string: '%s'\n", path, inode.I_type[:], strings.Trim(string(inode.I_type[:]), "\x00"))
		fmt.Printf("DEBUG CAT -> Inodo encontrado: %d en path: %s\n", idInodo, path)
		// Verificar que sea un archivo
		if string(inode.I_type[0]) != "1" {
			salida.WriteString(fmt.Sprintf("CAT Error: %s no es un archivo.\n", path))
			continue
		}

		// Verificar permisos de lectura (simulado)
		if !strings.Contains(string(inode.I_perm[:]), "6") && session.CurrentUser != "root" {
			salida.WriteString(fmt.Sprintf("CAT Error: No tiene permisos de lectura en %s\n", path))
			continue
		}

		salida.WriteString(fmt.Sprintf("Contenido de %s:\n", path))
		var leidos int32 = 0
blockSize := int32(binary.Size(Structs.Fileblock{}))
totalSize := inode.I_size

for i := 0; i < 12 && leidos < totalSize; i++ {
	if inode.I_block[i] == -1 {
		continue
	}

	blockPos := int64(super.S_block_start) + int64(inode.I_block[i])*int64(blockSize)
	var block Structs.Fileblock
	if err := Herramientas.ReadObject(disco, &block, blockPos); err != nil {
		salida.WriteString(fmt.Sprintf("Error al leer bloque: %v\n", err))
		continue
	}

	toRead := blockSize
	if totalSize-leidos < blockSize {
		toRead = totalSize - leidos
	}
	fmt.Printf("DEBUG CAT -> Leyendo bloque directo #%d (i_block[%d] = %d), pos: %d, leidos: %d, leer: %d\n",
		i, i, inode.I_block[i], blockPos, leidos, toRead)
	fmt.Printf("DEBUG CAT -> Bytes del bloque: %v\n", block.B_content[:toRead])
	

	// Limpieza: recortar hasta el primer byte nulo si existe (opcional si quieres aún más limpio)
	for j := int32(0); j < toRead; j++ {
		if block.B_content[j] == 0 {
			break // deja de leer al primer \x00 si es relleno
		}
		salida.WriteByte(block.B_content[j])
	}

	leidos += toRead
}


		salida.WriteString("-----------------------------\n")
	}
	return salida.String(), nil
}

