package AdministacionUserAndGrups

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	//"os"
	"backend/Herramientas"
	"backend/Structs"
	"backend/session"
)

// Mkgrp crea un grupo en la partición. Solo root puede ejecutarlo.
func Mkgrp(entrada []string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========MKGRP========")
	// Verificar que haya sesión activa y que el usuario sea root.
	if !session.Active {
		salida.WriteString(fmt.Sprintf("MKGRP Error: No hay sesión activa. Inicia sesión primero."))
		return salida.String(), nil
	}
	if session.CurrentUser != "root" {
		salida.WriteString(fmt.Sprintf("MKGRP Error: Solo el usuario root puede ejecutar este comando."))
		return salida.String(), nil
	}

	// Usar un struct para almacenar la información del grupo a crear.
	var newGroup Structs.Group
	paramCorrectos := true

	// Parseo de parámetros: se espera -name=<nombre_del_grupo>
	for _, parametro := range entrada[1:] {
		tmp := strings.TrimSpace(parametro)
		valores := strings.Split(tmp, "=")
		if len(valores) != 2 {
			salida.WriteString(fmt.Sprintf("MKGRP Error: Parámetro incorrecto:", tmp))
			paramCorrectos = false
			break
		}
		if strings.ToLower(valores[0]) == "name" {
			newGroup.Name = valores[1]
		} else {
			salida.WriteString(fmt.Sprintf("RMGRP Error: Parámetro desconocido:", valores[0]))
			paramCorrectos = false
			break
		}
	}
	if !paramCorrectos || newGroup.Name == "" {
		salida.WriteString(fmt.Sprintf("MKGRP Error: Faltan parámetros obligatorios."))
		return salida.String(), nil
	}

	// Verificar que se tenga asignada la partición de la sesión.
	if session.PartitionID == "" {
		salida.WriteString(fmt.Sprintf("MKGRP Error: No se encontró la partición de la sesión actual."))
		return salida.String(), nil
	}

	// Buscar la ruta del disco de la partición activa.
	var pathDico string
	for _, montado := range Structs.Montadas {
		if montado.Id == session.PartitionID {
			pathDico = montado.PathM
			break
		}
	}
	if pathDico == "" {
		salida.WriteString(fmt.Sprintf("MKGRP Error: No se encontró la partición montada para la sesión."))
		return salida.String(), nil
	}

	// Abrir el disco de la partición.
	disco, err := Herramientas.OpenFile(pathDico)
	if err != nil {
		salida.WriteString(fmt.Sprintf("MKGRP Error al abrir el disco:", err))
		return salida.String(), err
	}
	defer disco.Close()

	// Leer el MBR.
	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		salida.WriteString(fmt.Sprintf("MKGRP Error al leer MBR:", err))
		return salida.String(), err
	}

	// Buscar la partición en el MBR con el id de la sesión y leer su superbloque.
	var super Structs.Superblock
	var particion Structs.Partition
	encontrado := false
	for i := 0; i < 4; i++ {
		if Structs.GetId(string(mbr.Partitions[i].Id[:])) == session.PartitionID {
			particion = mbr.Partitions[i]
			if err := Herramientas.ReadObject(disco, &super, int64(particion.Start)); err != nil {
				salida.WriteString(fmt.Sprintf("MKGRP Error al leer superbloque:", err))
				return salida.String(), err
			}
			encontrado = true
			break
		}
	}
	if !encontrado {
		salida.WriteString(fmt.Sprintf("MKGRP Error: No se pudo encontrar la partición con id", session.PartitionID))
		return salida.String(), nil
	}

	// Leer el contenido actual del archivo users.txt.
	content, err := readUsersFile(disco, &super)
	if err != nil {
		salida.WriteString(fmt.Sprintf("MKGRP Error al leer users.txt:", err))
		return salida.String(), err
	}

	// Verificar que el grupo no exista y determinar el máximo GID actual.
	lines := strings.Split(content, "\n")
	maxGid := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		// Se esperan registros de grupo con la forma: UID, G, groupName.
		if len(parts) >= 3 && parts[1] == "G" {
			gid, err := strconv.Atoi(parts[0])
			if err == nil && gid > maxGid {
				maxGid = gid
			}
			if parts[2] == newGroup.Name {
				salida.WriteString(fmt.Sprintf("MKGRP Error: El grupo", newGroup.Name, "ya existe."))
				return salida.String(), nil
			}
		}
	}
	newGid := maxGid + 1

	// Preparar la nueva línea para el grupo: "newGid, G, newGroup.Name"
	newGroupLine := fmt.Sprintf("%d, G, %s\n", newGid, newGroup.Name)
	newContent := content
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += newGroupLine

	// Actualizar el archivo users.txt en disco.
	var folderBlock Structs.Folderblock
	if err := Herramientas.ReadObject(disco, &folderBlock, int64(super.S_block_start)); err != nil {
		salida.WriteString(fmt.Sprintf("MKGRP Error al leer carpeta raíz:", err))
		return salida.String(), err
	}
	usersInode := int32(-1)
	for i := 0; i < len(folderBlock.B_content); i++ {
		entryName := strings.TrimRight(string(folderBlock.B_content[i].B_name[:]), "\x00")
		if entryName == "users.txt" {
			usersInode = folderBlock.B_content[i].B_inodo
			break
		}
	}
	if usersInode == -1 {
		salida.WriteString(fmt.Sprintf("MKGRP Error: users.txt no encontrado en la carpeta raíz"))
		return salida.String(), nil
	}

	var userInode Structs.Inode
	inodeSize := int32(binary.Size(Structs.Inode{}))
	inodePosition := int64(super.S_inode_start) + int64(usersInode)*int64(inodeSize)
	if err := Herramientas.ReadObject(disco, &userInode, inodePosition); err != nil {
		salida.WriteString(fmt.Sprintf("MKGRP Error al leer inodo de users.txt:", err))
		return salida.String(), err
	}
	userInode.I_size = int32(len(newContent))
	fmt.Println((len(newContent)))
	
	blockIndex := userInode.I_block[0]
	if blockIndex < 0 {// aqui compara si el bloque es menor a 0
		salida.WriteString(fmt.Sprintf("MKGRP Error: bloque inválido para users.txt"))
		return salida.String(), nil
	}
	// Usar la función writeContentToFile para escribir newContent en múltiples bloques.
if err := session.WriteContentToFile(disco, &super, &userInode, newContent); err != nil {
    salida.WriteString(fmt.Sprintf("MKGRP Error al escribir el nuevo contenido en users.txt:", err))
    return salida.String(), err
}
if err := Herramientas.WriteObject(disco, userInode, inodePosition); err != nil {
    salida.WriteString(fmt.Sprintf("MKGRP Error al actualizar el inodo de users.txt:", err))
    return salida.String(), err
}

salida.WriteString(fmt.Sprintf("Grupo", newGroup.Name, "creado correctamente con ID", newGid))
return salida.String(), nil
}

// writeContentToFile escribe el contenido en el disco usando múltiples bloques si es necesario.
// Se asume que solo se usan los bloques directos (por ejemplo, los primeros 12 de I_block).
// Esta función actualiza el tamaño del inodo y asigna nuevos bloques (simulando la obtención
// de bloques libres, aquí se usa super.S_first_blo como contador simple)
// func writeContentToFile(disco *os.File, super *Structs.Superblock, inode *Structs.Inode, content string) error {
//     blockSize := int(binary.Size(Structs.Fileblock{}))
//     contentBytes := []byte(content)
//     totalLen := len(contentBytes)
//     numBlocksNeeded := (totalLen + blockSize - 1) / blockSize

//     maxDirect := 12
//     maxSimple := 16
//     maxDouble := 16 * 16
//     maxTriple := 16 * 16 * 16
//     totalMax := maxDirect + maxSimple + maxDouble + maxTriple

//     if numBlocksNeeded > totalMax {
//          fmt.Printf("El archivo es demasiado grande (limite: %d bloques)\n", totalMax)
//     }
//     // Recorrer los bloques necesarios para escribir el contenido
//     for i := 0; i < numBlocksNeeded; i++ {
//         start := i * blockSize
//         end := start + blockSize
//         if end > totalLen {
//             end = totalLen
//         }
//         chunk := contentBytes[start:end]//chunk es un arreglo de bytes

//         var blockIndex int32 = -1
//         // Determinar el índice del bloque a usar
//         switch {
//         case i < maxDirect:
//             // Bloques directos
//             if inode.I_block[i] == -1 {
//                 b, err := obtenerBloqueLibre(disco, super)
//                 if err != nil { return err }
//                 inode.I_block[i] = b
//             }
//             blockIndex = inode.I_block[i]
//             //aqui se escribe el contenido en el bloque encontrado
//         case i < maxDirect+maxSimple:
//             // Indirecto simple
//             idx := i - maxDirect
//             if inode.I_block[12] == -1 {
//                 b, err := obtenerBloqueLibre(disco, super)
//                 if err != nil { return err }
//                 inode.I_block[12] = b
//                 pb := Structs.Pointerblock{}
//                 for j := range pb.B_pointers { pb.B_pointers[j] = -1 }
//                 Herramientas.WriteObject(disco, pb, int64(super.S_block_start)+int64(b)*int64(blockSize))
//             }
//             //aqui asignamos el valor de la posicion
//             ptrPos := int64(super.S_block_start) + int64(inode.I_block[12])*int64(blockSize)
//             pb := Structs.Pointerblock{}
//             Herramientas.ReadObject(disco, &pb, ptrPos)

//             if pb.B_pointers[idx] == -1 {
//                 b, err := obtenerBloqueLibre(disco, super)
//                 if err != nil { return err }
//                 pb.B_pointers[idx] = b
//                 Herramientas.WriteObject(disco, pb, ptrPos)
//             }

//             blockIndex = pb.B_pointers[idx]

//         case i < maxDirect+maxSimple+maxDouble:
//             // Indirecto doble
//             idx := i - maxDirect - maxSimple
//             outer := idx / 16
//             inner := idx % 16

//             if inode.I_block[13] == -1 {
//                 b, err := obtenerBloqueLibre(disco, super)
//                 if err != nil { return err }
//                 inode.I_block[13] = b
//                 outerPB := Structs.Pointerblock{}
//                 for j := range outerPB.B_pointers { outerPB.B_pointers[j] = -1 }
//                 Herramientas.WriteObject(disco, outerPB, int64(super.S_block_start)+int64(b)*int64(blockSize))
//             }

//             outerPos := int64(super.S_block_start) + int64(inode.I_block[13])*int64(blockSize)
//             outerPB := Structs.Pointerblock{}
//             Herramientas.ReadObject(disco, &outerPB, outerPos)

//             if outerPB.B_pointers[outer] == -1 {
//                 b, err := obtenerBloqueLibre(disco, super)
//                 if err != nil { return err }
//                 outerPB.B_pointers[outer] = b
//                 innerPB := Structs.Pointerblock{}
//                 for j := range innerPB.B_pointers { innerPB.B_pointers[j] = -1 }
//                 Herramientas.WriteObject(disco, innerPB, int64(super.S_block_start)+int64(b)*int64(blockSize))
//                 Herramientas.WriteObject(disco, outerPB, outerPos)
//             }

//             innerPos := int64(super.S_block_start) + int64(outerPB.B_pointers[outer])*int64(blockSize)
//             innerPB := Structs.Pointerblock{}
//             Herramientas.ReadObject(disco, &innerPB, innerPos)

//             if innerPB.B_pointers[inner] == -1 {
//                 b, err := obtenerBloqueLibre(disco, super)
//                 if err != nil { return err }
//                 innerPB.B_pointers[inner] = b
//                 Herramientas.WriteObject(disco, innerPB, innerPos)
//             }

//             blockIndex = innerPB.B_pointers[inner]

//         default:
//             // Indirecto triple
//             idx := i - maxDirect - maxSimple - maxDouble
//             lvl1 := idx / (16 * 16)
//             lvl2 := (idx / 16) % 16
//             lvl3 := idx % 16

//             if inode.I_block[14] == -1 {
//                 b, err := obtenerBloqueLibre(disco, super)
//                 if err != nil { return err }
//                 inode.I_block[14] = b
//                 p1 := Structs.Pointerblock{}
//                 for j := range p1.B_pointers { p1.B_pointers[j] = -1 }
//                 Herramientas.WriteObject(disco, p1, int64(super.S_block_start)+int64(b)*int64(blockSize))
//             }

//             lvl1Pos := int64(super.S_block_start) + int64(inode.I_block[14])*int64(blockSize)
//             p1 := Structs.Pointerblock{}
//             Herramientas.ReadObject(disco, &p1, lvl1Pos)

//             if p1.B_pointers[lvl1] == -1 {
//                 b, err := obtenerBloqueLibre(disco, super)
//                 if err != nil { return err }
//                 p1.B_pointers[lvl1] = b
//                 p2 := Structs.Pointerblock{}
//                 for j := range p2.B_pointers { p2.B_pointers[j] = -1 }
//                 Herramientas.WriteObject(disco, p2, int64(super.S_block_start)+int64(b)*int64(blockSize))
//                 Herramientas.WriteObject(disco, p1, lvl1Pos)
//             }

//             lvl2Pos := int64(super.S_block_start) + int64(p1.B_pointers[lvl1])*int64(blockSize)
//             p2 := Structs.Pointerblock{}
//             Herramientas.ReadObject(disco, &p2, lvl2Pos)

//             if p2.B_pointers[lvl2] == -1 {
//                 b, err := obtenerBloqueLibre(disco, super)
//                 if err != nil { return err }
//                 p2.B_pointers[lvl2] = b
//                 p3 := Structs.Pointerblock{}
//                 for j := range p3.B_pointers { p3.B_pointers[j] = -1 }
//                 Herramientas.WriteObject(disco, p3, int64(super.S_block_start)+int64(b)*int64(blockSize))
//                 Herramientas.WriteObject(disco, p2, lvl2Pos)
//             }

//             lvl3Pos := int64(super.S_block_start) + int64(p2.B_pointers[lvl2])*int64(blockSize)
//             p3 := Structs.Pointerblock{}
//             Herramientas.ReadObject(disco, &p3, lvl3Pos)

//             if p3.B_pointers[lvl3] == -1 {
//                 b, err := obtenerBloqueLibre(disco, super)
//                 if err != nil { return err }
//                 p3.B_pointers[lvl3] = b
//                 Herramientas.WriteObject(disco, p3, lvl3Pos)
//             }

//             blockIndex = p3.B_pointers[lvl3]
//         }

//         // Escribir contenido en el bloque encontrado
//         blockPos := int64(super.S_block_start) + int64(blockIndex)*int64(blockSize)
//         var fileBlock Structs.Fileblock
//         copy(fileBlock.B_content[:], chunk)
//         if err := Herramientas.WriteObject(disco, fileBlock, blockPos); err != nil {
//             return err
//         }
//     }

//     inode.I_size = int32(totalLen)
//     return nil
// }



// func obtenerBloqueLibre(disco *os.File, super *Structs.Superblock) (int32, error) {
//     bitmap := make([]byte, super.S_blocks_count)
//     if err := Herramientas.ReadObject(disco, &bitmap, int64(super.S_bm_block_start)); err != nil {
//         return -1, err
//     }

//     for i := int32(0); i < super.S_blocks_count; i++ {
//         if bitmap[i] == 0 {
//             bitmap[i] = 1
//             super.S_free_blocks_count--
//             if err := Herramientas.WriteObject(disco, bitmap, int64(super.S_bm_block_start)); err != nil {
//                 return -1, err
//             }
//             return i, nil
//         }
//     }
//     return -1, fmt.Errorf("no hay bloques libres")
// }

