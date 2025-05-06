package AdministacionUserAndGrups

import (
	"fmt"
	"os"
	"strings"
	"encoding/binary"

	"backend/Herramientas"
	"backend/Structs"
	"backend/session"
)

// Login usa el struct de login definido en Structs.Login para almacenar los parámetros.
func Login(entrada []string) (string, error){
	var salida strings.Builder
	salida.WriteString("========Chgrp========")
	// Crear una variable del struct Login.
	var loginData Structs.Login
	paramsCorrectos := true

	// Parseo de parámetros: se esperan -user, -pass y -id.
	for _, parametro := range entrada[1:] {
		tmp := strings.TrimSpace(parametro)
		valores := strings.Split(tmp, "=")
		if len(valores) != 2 {
			salida.WriteString(fmt.Sprintf("LOGIN Error: Parámetro incorrecto:", tmp))
			paramsCorrectos = false
			break
		}
		key := strings.ToLower(valores[0])
		value := valores[1]
		switch key {
		case "user":
			loginData.User = value
		case "pass":
			loginData.Pass = value
		case "id":
			loginData.Id = strings.ToUpper(value)
		default:
			salida.WriteString(fmt.Sprintf("LOGIN Error: Parámetro desconocido:", valores[0]))
			paramsCorrectos = false
			break
		}
	}

	if !paramsCorrectos || loginData.User == "" || loginData.Pass == "" || loginData.Id == "" {
		salida.WriteString(fmt.Sprintf("LOGIN Error: Faltan parámetros obligatorios."))
		return salida.String(), nil
	}

	// Verificar si ya hay una sesión activa.
	if session.Active {
		salida.WriteString(fmt.Sprintf("Ya hay un usuario logueado. Cierre la sesión actual para iniciar otra."))
		return salida.String(), nil
	}

	// Buscar la partición montada correspondiente al id.
	var pathDico string
	for _, montado := range Structs.Montadas {
		if montado.Id == loginData.Id {
			pathDico = montado.PathM
			break
		}
	}
	if pathDico == "" {
		salida.WriteString(fmt.Sprintf("LOGIN Error: No se encontró la partición con el id", loginData.Id))
		return salida.String(), nil
	}

	// Abrir el disco de la partición.
	disco, err := Herramientas.OpenFile(pathDico)
	if err != nil {
		salida.WriteString(fmt.Sprintf("LOGIN Error al abrir la partición:", err))
		return salida.String(), err
	}
	defer disco.Close()

	// Leer el MBR para obtener la partición.
	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		salida.WriteString(fmt.Sprintf("LOGIN Error al leer MBR:", err))
		return salida.String(), err
	}

	// Buscar la partición por id y leer su superbloque.
	var super Structs.Superblock
	var particion Structs.Partition
	encontrado := false
	for i := 0; i < 4; i++ {
		if Structs.GetId(string(mbr.Partitions[i].Id[:])) == loginData.Id {
			particion = mbr.Partitions[i]
			if err := Herramientas.ReadObject(disco, &super, int64(particion.Start)); err != nil {
				salida.WriteString(fmt.Sprintf("LOGIN Error al leer superbloque:", err))
				return salida.String(), err
			}
			encontrado = true
			break
		}
	}
	if !encontrado {
		salida.WriteString(fmt.Sprintf("LOGIN Error: No se pudo encontrar la partición con id", loginData.Id))
		return salida.String(), nil
	}

	// Leer el contenido del archivo users.txt.
	content, err := readUsersFile(disco, &super)
	if err != nil {
		salida.WriteString(fmt.Sprintf("LOGIN Error al leer users.txt:", err))
		return salida.String(), err
	}

	// Buscar en el archivo el usuario.
	lines := strings.Split(content, "\n")
	var authSuccess bool = false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		// Registro de usuario: UID, U, grupo, usuario, contraseña.
		if parts[1] == "U" && parts[3] == loginData.User && parts[4] == loginData.Pass {
			authSuccess = true
			break
		}
	}
	if !authSuccess {
		salida.WriteString(fmt.Sprintf("LOGIN Error: Usuario no existe o contraseña incorrecta."))
		return salida.String(), nil
	}

	// Iniciar sesión: actualizar las variables de sesión.
	session.Active = true
	session.CurrentUser = loginData.User
	session.PartitionID = loginData.Id
	salida.WriteString(fmt.Sprintf("Usuario", loginData.User, "logueado correctamente a la partición", loginData.Id))
	return salida.String(), nil
}

// La función readUsersFile se mantiene sin cambios, realizando el Trim de bytes nulos.
func readUsersFile(disco *os.File, super *Structs.Superblock) (string, error) {
	var folderBlock Structs.Folderblock
	if err := Herramientas.ReadObject(disco, &folderBlock, int64(super.S_block_start)); err != nil {
		return "", err
	}

	var usersInode int32 = -1
	for i := 0; i < len(folderBlock.B_content); i++ {
		entryName := Structs.GetName(string(folderBlock.B_content[i].B_name[:]))
		if entryName == "users.txt" {
			usersInode = folderBlock.B_content[i].B_inodo
			break
		}
	}
	if usersInode == -1 {
		return "", fmt.Errorf("users.txt no encontrado en la carpeta raíz")
	}

	// Leer el inodo del archivo
	var userInode Structs.Inode
	inodeSize := int32(binary.Size(Structs.Inode{}))
	inodePosition := int64(super.S_inode_start) + int64(usersInode)*int64(inodeSize)
	if err := Herramientas.ReadObject(disco, &userInode, inodePosition); err != nil {
		return "", err
	}

	blockSize := int32(binary.Size(Structs.Fileblock{}))
	content := ""

	// Leer bloques directos
	for i := 0; i < 12; i++ {
		block := userInode.I_block[i]
		if block == -1 {
			continue
		}
		var fileBlock Structs.Fileblock
		blockPos := int64(super.S_block_start) + int64(block)*int64(blockSize)
		if err := Herramientas.ReadObject(disco, &fileBlock, blockPos); err != nil {
			return "", err
		}
		content += string(fileBlock.B_content[:])
	}

	// Leer bloque indirecto simple (I_block[12])
	if userInode.I_block[12] != -1 {
		var pb Structs.Pointerblock
		ptrPos := int64(super.S_block_start) + int64(userInode.I_block[12])*int64(blockSize)
		if err := Herramientas.ReadObject(disco, &pb, ptrPos); err != nil {
			return "", err
		}

		for i := 0; i < 16; i++ {
			if pb.B_pointers[i] == -1 {
				continue
			}
			var fileBlock Structs.Fileblock
			blockPos := int64(super.S_block_start) + int64(pb.B_pointers[i])*int64(blockSize)
			if err := Herramientas.ReadObject(disco, &fileBlock, blockPos); err != nil {
				return "", err
			}
			content += string(fileBlock.B_content[:])
		}
	}

	// Leer bloque indirecto doble (I_block[13])
	if userInode.I_block[13] != -1 {
		var outerPB Structs.Pointerblock
		outerPos := int64(super.S_block_start) + int64(userInode.I_block[13])*int64(blockSize)
		if err := Herramientas.ReadObject(disco, &outerPB, outerPos); err != nil {
			return "", err
		}

		for i := 0; i < 16; i++ {
			if outerPB.B_pointers[i] == -1 {
				continue
			}
			var innerPB Structs.Pointerblock
			innerPos := int64(super.S_block_start) + int64(outerPB.B_pointers[i])*int64(blockSize)
			if err := Herramientas.ReadObject(disco, &innerPB, innerPos); err != nil {
				return "", err
			}

			for j := 0; j < 16; j++ {
				if innerPB.B_pointers[j] == -1 {
					continue
				}
				var fileBlock Structs.Fileblock
				blockPos := int64(super.S_block_start) + int64(innerPB.B_pointers[j])*int64(blockSize)
				if err := Herramientas.ReadObject(disco, &fileBlock, blockPos); err != nil {
					return "", err
				}
				content += string(fileBlock.B_content[:])
			}
		}
	}

	// Eliminar bytes nulos (\x00) si quedaron al final
	content = strings.TrimRight(content, "\x00")
	return content, nil
}



