package AdministacionUserAndGrups

import (
	"encoding/binary"
	"fmt"
	"strings"

	"backend/Herramientas"
	"backend/Structs"
	"backend/session"
)

// Rmgrp elimina un grupo en el archivo users.txt.
// Solo lo puede ejecutar el usuario root.
func Rmgrp(entrada []string) (string, error) {	
	var salida strings.Builder
	salida.WriteString("========RMGRP========")
	// Verificar que haya sesión activa y que el usuario sea root.
	if !session.Active {
		salida.WriteString(fmt.Sprintf("RMGRP Error: No hay sesión activa. Inicia sesión primero."))
		return salida.String(), nil
	}
	if session.CurrentUser != "root" {
		salida.WriteString(fmt.Sprintf("RMGRP Error: Solo el usuario root puede ejecutar este comando."))
		return salida.String(), nil
	}

	// Usar un struct para almacenar la información del grupo a eliminar.
	var groupToDelete Structs.Group
	paramCorrectos := true
	// Se espera el parámetro -name=<nombre_del_grupo>
	for _, parametro := range entrada[1:] {
		tmp := strings.TrimSpace(parametro)
    valores := strings.Split(tmp, "=")
    if len(valores) != 2 {
        salida.WriteString(fmt.Sprintf("RMGRP Error: Parámetro incorrecto:", tmp))
        paramCorrectos = false
        break
    }
    if strings.ToLower(valores[0]) == "name" {
        // Aseguramos eliminar espacios y caracteres extra
        groupToDelete.Name = strings.ReplaceAll(strings.TrimSpace(valores[1]), "\u200B", "")
    } else {
        salida.WriteString(fmt.Sprintf("RMGRP Error: Parámetro desconocido:", valores[0]))
        paramCorrectos = false
        break
    }
	}
	if !paramCorrectos || groupToDelete.Name == "" {
		salida.WriteString(fmt.Sprintf("RMGRP Error: Faltan parámetros obligatorios."))
		return salida.String(), nil
	}

	// Verificar que se tenga asignada la partición de la sesión.
	if session.PartitionID == "" {
		salida.WriteString(fmt.Sprintf("RMGRP Error: No se encontró la partición de la sesión actual."))
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
		salida.WriteString(fmt.Sprintf("RMGRP Error: No se encontró la partición montada para la sesión."))
		return salida.String(), nil
	}

	// Abrir el disco de la partición.
	disco, err := Herramientas.OpenFile(pathDico)
	if err != nil {
		salida.WriteString(fmt.Sprintf("RMGRP Error al abrir el disco:", err))
		return salida.String(), err
	}
	defer disco.Close()

	// Leer el MBR.
	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		salida.WriteString(fmt.Sprintf("RMGRP Error al leer MBR:", err))
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
				salida.WriteString(fmt.Sprintf("RMGRP Error al leer superbloque:", err))
				return salida.String(), err
			}
			encontrado = true
			break
		}
	}
	if !encontrado {
		salida.WriteString(fmt.Sprintf("RMGRP Error: No se pudo encontrar la partición con id", session.PartitionID))
		return salida.String(), nil
	}

	// Leer el contenido actual del archivo users.txt.
	content, err := readUsersFile(disco, &super)
	if err != nil {
		salida.WriteString(fmt.Sprintf("RMGRP Error al leer users.txt:", err))
		return salida.String(), err
	}

	// Buscar el grupo a eliminar en el contenido.
	lineas := strings.Split(content, "\n")
	encontradoGrupo := false
	for i, linea := range lineas {
		trimmed := strings.TrimSpace(linea)
		if trimmed == "" {
			continue
		}
		parts := strings.Split(trimmed, ",")
		for j := range parts {
			parts[j] = strings.TrimSpace(parts[j])
		}
		//fmt.Printf("Comparando [%s] con [%s]\n", parts[2], groupToDelete.Name)
		// Se espera que el registro de grupo tenga la forma: UID, G, groupName.
		if len(parts) >= 3 && parts[1] == "G" && parts[2] == groupToDelete.Name {
			if parts[0] == "0" {
				salida.WriteString(fmt.Sprintf("RMGRP Error: El grupo", groupToDelete.Name, "ya está eliminado."))
				return salida.String(), nil
			}
			// Marcar el grupo como eliminado: establecer UID a "0".
			parts[0] = "0"
			// Reconstruir la línea.
			newLine := strings.Join(parts, ", ")
			lineas[i] = newLine
			encontradoGrupo = true
			fmt.Printf("Comparando [%s] con [%s]\n", parts[2], groupToDelete.Name)
			break
		}
	}
	if !encontradoGrupo {
		salida.WriteString(fmt.Sprintf("RMGRP Error: El grupo", groupToDelete.Name, "no se encontró en la partición."))
		return salida.String(), nil
	}

	// Reconstruir el contenido actualizado.
	newContent := ""
	for _, linea := range lineas {
		if strings.TrimSpace(linea) != "" {
			newContent += linea + "\n"
		}
	}

	// Actualizar el archivo users.txt en disco.
	var folderBlock Structs.Folderblock
	if err := Herramientas.ReadObject(disco, &folderBlock, int64(super.S_block_start)); err != nil {
		salida.WriteString(fmt.Sprintf("RMGRP Error al leer carpeta raíz:", err))
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
		salida.WriteString(fmt.Sprintf("RMGRP Error: users.txt no encontrado en la carpeta raíz"))
		return salida.String(), nil
	}

	var userInode Structs.Inode
	inodeSize := int32(binary.Size(Structs.Inode{}))
	inodePosition := int64(super.S_inode_start) + int64(usersInode)*int64(inodeSize)
	if err := Herramientas.ReadObject(disco, &userInode, inodePosition); err != nil {
		salida.WriteString(fmt.Sprintf("RMGRP Error al leer inodo de users.txt:", err))
		return salida.String(), err
	}
	
	// Escribir el nuevo contenido completo en users.txt
	if err := session.WriteContentToFile(disco, &super, &userInode, newContent); err != nil {
    	salida.WriteString(fmt.Sprintf("RMGRP Error al escribir el nuevo contenido en users.txt:", err))
    	return salida.String(), err
	}	
	fmt.Println(len(newContent))
	// Guardar el inodo actualizado con nuevo I_size
	if err := Herramientas.WriteObject(disco, userInode, inodePosition); err != nil {
    	salida.WriteString(fmt.Sprintf("RMGRP Error al actualizar el inodo de users.txt:", err))
   	 return salida.String(), err
	}

	// Mostrar el listado de grupos actualizado.
	salida.WriteString(fmt.Sprintf("Grupos actualizados:"))
	//for para recorrer el listado de grupos actuales
	for _, linea := range lineas {
		if strings.TrimSpace(linea) != "" {
			salida.WriteString(fmt.Sprintf(linea))
		}
	}
	fmt.Printf("%q\n", lineas)
	salida.WriteString(fmt.Sprintf("Grupo", groupToDelete.Name, "eliminado correctamente."))
	return salida.String(), nil
}


