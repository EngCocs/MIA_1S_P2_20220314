package AdministacionUserAndGrups

import (
	"encoding/binary"
	"fmt"
	"strings"

	"backend/Herramientas"
	"backend/Structs"
	"backend/session"
)
// Chgrp cambia el grupo de un usuario en el archivo users.txt
//chgrp -user=user2 -grp=grupo1 ||  chgrp -user=user1 -grp=usuarios2


func Chgrp(entrada []string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========Chgrp========")
	// Verificar que haya sesión activa y que el usuario sea root
	if !session.Active {
		salida.WriteString(fmt.Sprintf("CHGRP Error: No hay sesión activa."))
		return salida.String(), nil
	}
	// Solo el usuario root puede ejecutar este comando
	if session.CurrentUser != "root" {
		salida.WriteString(fmt.Sprintf("CHGRP Error: Solo el usuario root puede ejecutar este comando."))
		return salida.String(), nil
	}

	var cambio Structs.ChangeGRP
	paramCorrectos := true

	for _, param := range entrada[1:] {
		tmp := strings.TrimSpace(param)
		valores := strings.Split(tmp, "=")
		if len(valores) != 2 {
			salida.WriteString(fmt.Sprintf("CHGRP Error: Parámetro incorrecto:", param))
			paramCorrectos = false
			break
		}
		key := strings.ToLower(valores[0])
		value := strings.ReplaceAll(strings.TrimSpace(valores[1]), "\u200B", "")
		switch key {
		case "user":
			cambio.User = value
		case "grp":
			cambio.Grp = value
		default:
			salida.WriteString(fmt.Sprintf("CHGRP Error: Parámetro desconocido:", key))
			paramCorrectos = false
		}
	}

	if !paramCorrectos || cambio.User == "" || cambio.Grp == "" {
		salida.WriteString(fmt.Sprintf("CHGRP Error: Faltan parámetros obligatorios."))
		return salida.String(), nil
	}

	// Obtener path del disco
	var pathDico string
	for _, montado := range Structs.Montadas {
		if montado.Id == session.PartitionID {
			pathDico = montado.PathM
			break
		}
	}
	if pathDico == "" {
		salida.WriteString(fmt.Sprintf("CHGRP Error: No se encontró la partición montada."))
		return salida.String(), nil
	}

	disco, err := Herramientas.OpenFile(pathDico)
	if err != nil {
		salida.WriteString(fmt.Sprintf("CHGRP Error al abrir el disco:", err))
		return salida.String(), err
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		salida.WriteString(fmt.Sprintf("CHGRP Error al leer el MBR:", err))
		return salida.String(), err
	}

	var super Structs.Superblock
	var particion Structs.Partition
	encontrado := false
	for i := 0; i < 4; i++ {
		if Structs.GetId(string(mbr.Partitions[i].Id[:])) == session.PartitionID {
			particion = mbr.Partitions[i]
			if err := Herramientas.ReadObject(disco, &super, int64(particion.Start)); err != nil {
				salida.WriteString(fmt.Sprintf("CHGRP Error al leer superbloque:", err))
				return salida.String(), err
			}
			encontrado = true
			break
		}
	}
	if !encontrado {
		salida.WriteString(fmt.Sprintf("CHGRP Error: No se encontró la partición."))
		return salida.String(), nil
	}

	// Leer el contenido actual del archivo users.txt.
	content, err := readUsersFile(disco, &super)
	if err != nil {
		salida.WriteString(fmt.Sprintf("CHGRP Error al leer users.txt:", err))
		return salida.String(), err
	}

	lineas := strings.Split(content, "\n")
	grupoExiste := false
	usuarioEncontrado := false
    // Verificar que el grupo exista 
	for _, linea := range lineas {
		partes := strings.Split(strings.TrimSpace(linea), ",")
		for j := range partes {
			partes[j] = strings.TrimSpace(partes[j])
		}
		if len(partes) >= 3 && partes[1] == "G" && partes[2] == cambio.Grp && partes[0] != "0" {
			grupoExiste = true
		}
	}

	if !grupoExiste {
		salida.WriteString(fmt.Sprintf("CHGRP Error: El grupo", cambio.Grp, "no existe o está eliminado."))
		return salida.String(), nil
	}
	//use el for para recorrer las lineas y buscar el usuario
	//y cambiar el grupo
	for i, linea := range lineas {
		partes := strings.Split(strings.TrimSpace(linea), ",")
		for j := range partes {
			partes[j] = strings.TrimSpace(partes[j])
		}
		
		if len(partes) >= 5 && partes[1] == "U" && partes[3] == cambio.User {
			if partes[0] == "0" {
				salida.WriteString(fmt.Sprintf("CHGRP Error: El usuario", cambio.User, "ya está eliminado."))
				return salida.String(), nil
			}
			// Cambiar el grupo pasamos el grupo a la posicion 2 antes era 3
			partes[2] = cambio.Grp // Modificamos el grupo
			lineas[i] = strings.Join(partes, ", ")
			usuarioEncontrado = true
			break
		}
	}

	if !usuarioEncontrado {
		salida.WriteString(fmt.Sprintf("CHGRP Error: El usuario", cambio.User, "no se encontró."))
		return salida.String(), nil
	}

	// Reconstruir el nuevo contenido
	newContent := ""
	for _, linea := range lineas {
		if strings.TrimSpace(linea) != "" {
			newContent += linea + "\n"
		}
	}

	//leemos el inodo de la carpeta raiz
	var folderBlock Structs.Folderblock
	if err := Herramientas.ReadObject(disco, &folderBlock, int64(super.S_block_start)); err != nil {
		salida.WriteString(fmt.Sprintf("CHGRP Error al leer carpeta raíz:", err))
		return salida.String(), err
	}
	// Buscar el inodo de users.txt
	usersInode := int32(-1)
	for _, content := range folderBlock.B_content {
		entryName := strings.TrimRight(string(content.B_name[:]), "\x00")
		if entryName == "users.txt" {
			usersInode = content.B_inodo
			break
		}
	}
	if usersInode == -1 {
		salida.WriteString(fmt.Sprintf("CHGRP Error: users.txt no encontrado en la raíz."))
		return salida.String(), nil
	}
	// Leer el inodo del archivo
	var userInode Structs.Inode
	inodeSize := int32(binary.Size(Structs.Inode{}))
	inodePosition := int64(super.S_inode_start) + int64(usersInode)*int64(inodeSize)
	if err := Herramientas.ReadObject(disco, &userInode, inodePosition); err != nil {
		salida.WriteString(fmt.Sprintf("CHGRP Error al leer inodo de users.txt:", err))
		return salida.String(), err
	}

	// Escribir el nuevo contenido
	if err := session.WriteContentToFile(disco, &super, &userInode, newContent); err != nil {
		salida.WriteString(fmt.Sprintf("CHGRP Error al escribir users.txt:", err))
		return salida.String(), err
	}

	if err := Herramientas.WriteObject(disco, userInode, inodePosition); err != nil {
		salida.WriteString(fmt.Sprintf("CHGRP Error al actualizar inodo de users.txt:", err))
		return salida.String(), err
	}

	salida.WriteString(fmt.Sprintf("Usuario", cambio.User, "cambiado al grupo", cambio.Grp, "correctamente."))
	return salida.String(), nil
}
