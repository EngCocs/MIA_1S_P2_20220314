package AdministacionUserAndGrups

import (
	"encoding/binary"
	"fmt"
	"strings"

	"backend/Herramientas"
	"backend/Structs"
	"backend/session"
)

func Rmusr(entrada []string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========RMUSR========")
	if !session.Active {
		salida.WriteString(fmt.Sprintf("RMUSR Error: No hay sesión activa."))
		return salida.String(), nil
	}
	if session.CurrentUser != "root" {
		salida.WriteString(fmt.Sprintf("RMUSR Error: Solo el usuario root puede ejecutar este comando."))
		return salida.String(), nil
	}

	var userNameDelete string
	paramCorrectos := true

	for _, param := range entrada[1:] {
		tmp := strings.TrimSpace(param)
		valores := strings.Split(tmp, "=")
		if len(valores) != 2 {
			salida.WriteString(fmt.Sprintf("RMUSR Error: Parámetro incorrecto:", param))
			paramCorrectos = false
			break
		}
		if strings.ToLower(valores[0]) == "user" {
			userNameDelete = strings.ReplaceAll(strings.TrimSpace(valores[1]), "\u200B", "")
		} else {
			salida.WriteString(fmt.Sprintf("RMUSR Error: Parámetro desconocido:", valores[0]))
			paramCorrectos = false
			break
		}
	}
	if !paramCorrectos || userNameDelete == "" {
		salida.WriteString(fmt.Sprintf("RMUSR Error: Falta el parámetro obligatorio -user."))
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
		salida.WriteString(fmt.Sprintf("RMUSR Error: No se encontró la partición montada para la sesión."))
		return salida.String(), nil
	}

	disco, err := Herramientas.OpenFile(pathDico)
	if err != nil {
		salida.WriteString(fmt.Sprintf("RMUSR Error al abrir el disco:", err))
		return salida.String(), err
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		salida.WriteString(fmt.Sprintf("RMUSR Error al leer el MBR:", err))
		return salida.String(), err
	}

	var super Structs.Superblock
	var particion Structs.Partition
	encontrado := false
	for i := 0; i < 4; i++ {
		if Structs.GetId(string(mbr.Partitions[i].Id[:])) == session.PartitionID {
			particion = mbr.Partitions[i]
			if err := Herramientas.ReadObject(disco, &super, int64(particion.Start)); err != nil {
				salida.WriteString(fmt.Sprintf("RMUSR Error al leer superbloque:", err))
				return salida.String(), err
			}
			encontrado = true
			break
		}
	}
	if !encontrado {
		salida.WriteString(fmt.Sprintf("RMUSR Error: Partición no encontrada."))
		return salida.String(), nil
	}

	// Leer el contenido del archivo users.txt
	content, err := readUsersFile(disco, &super)
	if err != nil {
		salida.WriteString(fmt.Sprintf("RMUSR Error al leer users.txt:", err))
		return salida.String(), err
	}

	// Buscar el usuario a eliminar
	lineas := strings.Split(content, "\n")
	UsuarioEncontrado := false
	for i, linea := range lineas {
		trimmed := strings.TrimSpace(linea)
		if trimmed == "" {
			continue
		}
		parts := strings.Split(trimmed, ",")
		for j := range parts {
			parts[j] = strings.TrimSpace(parts[j])
		}
		if len(parts) >= 5 && parts[1] == "U" && parts[3] == userNameDelete {
			if parts[0] == "0" {
				salida.WriteString(fmt.Sprintf("RMUSR Error: El usuario", userNameDelete, "ya está eliminado."))
				return salida.String(), nil
			}
			parts[0] = "0" // Marcar como eliminado
			lineas[i] = strings.Join(parts, ", ")
			UsuarioEncontrado = true
			break
		}
	}
	if !UsuarioEncontrado {
		salida.WriteString(fmt.Sprintf("RMUSR Error: El usuario", userNameDelete, "no se encontró."))
		return salida.String(), nil
	}

	// Reconstruir el nuevo contenido
	newContent := ""
	for _, linea := range lineas {
		if strings.TrimSpace(linea) != "" {//le quita los espacios en blanco
			newContent += linea + "\n"//le agregamos un salto de linea
		}
	}

	// Buscar inodo de users.txt
	var folderBlock Structs.Folderblock
	if err := Herramientas.ReadObject(disco, &folderBlock, int64(super.S_block_start)); err != nil {
		salida.WriteString(fmt.Sprintf("RMUSR Error al leer carpeta raíz:", err))
		return salida.String(), err
	}
	usersInode := int32(-1)//apunta a -1 dando a entender que no se encontro
	for _, content := range folderBlock.B_content {
		EntraName := strings.TrimRight(string(content.B_name[:]), "\x00")//aqui asignamos la entrada
		if EntraName == "users.txt" {
			usersInode = content.B_inodo
			break
		}
	}
	if usersInode == -1 {
		salida.WriteString(fmt.Sprintf("RMUSR Error: users.txt no encontrado en la raíz."))
		return salida.String(), nil
	}

	// Leer el inodo
	var userInode Structs.Inode
	inodeSize := int32(binary.Size(Structs.Inode{}))
	inodePosition := int64(super.S_inode_start) + int64(usersInode)*int64(inodeSize)
	if err := Herramientas.ReadObject(disco, &userInode, inodePosition); err != nil {
		salida.WriteString(fmt.Sprintf("RMUSR Error al leer inodo de users.txt:", err))
		return salida.String(), err
	}

	// Escribir el nuevo contenido
	if err := session.WriteContentToFile(disco, &super, &userInode, newContent); err != nil {
		salida.WriteString(fmt.Sprintf("RMUSR Error al escribir users.txt:", err))
		return salida.String(), err
	}

	// Actualizar el inodo
	if err := Herramientas.WriteObject(disco, userInode, inodePosition); err != nil {
		salida.WriteString(fmt.Sprintf("RMUSR Error al actualizar inodo:", err))
		return salida.String(), err
	}
	for _, linea := range lineas {
		if strings.TrimSpace(linea) != "" {
			salida.WriteString(fmt.Sprintf(linea))
		}
	}
	salida.WriteString(fmt.Sprintf("Usuario", userNameDelete, "eliminado correctamente."))
	return salida.String(), nil
}
