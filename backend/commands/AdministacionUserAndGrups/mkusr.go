package AdministacionUserAndGrups

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"backend/Herramientas"
	"backend/Structs"
	"backend/session"
)

// Mkusr crea un usuario en la partición. Solo lo puede ejecutar el usuario root.
// Se utiliza el struct CreateUser para almacenar los parámetros.
func Mkusr(entrada []string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========MKUSR========")
	// Verificar que haya sesión activa y que el usuario sea root.
	if !session.Active {
		salida.WriteString(fmt.Sprintf("MKUSR Error: No hay sesión activa. Inicia sesión primero."))
		return salida.String(), nil
	}
	if session.CurrentUser != "root" {
		salida.WriteString(fmt.Sprintf("MKUSR Error: Solo el usuario root puede ejecutar este comando."))
		return salida.String(), nil
	}

	// Declarar variable del tipo CreateUser
	var newUser Structs.CreateUser
	paramCorrectos := true

	// Parsear parámetros: se esperan -user, -pass y -grp.
	for _, parametro := range entrada[1:] {
		tmp := strings.TrimSpace(parametro)
		valores := strings.Split(tmp, "=")
		if len(valores) != 2 {
			salida.WriteString(fmt.Sprintf("MKUSR Error: Parámetro incorrecto:", tmp))
			paramCorrectos = false
			break
		}

		//Usamos switch para para que sea vea mas ordenado 
		key := strings.ToLower(valores[0])
		value := strings.TrimSpace(valores[1])
		switch key {
		case "user":
			newUser.User = value
		case "pass":
			newUser.Pass = value
		case "grp":
			newUser.Grp = value
		default:
			salida.WriteString(fmt.Sprintf("MKUSR Error: Parámetro desconocido:", valores[0]))
			paramCorrectos = false
		}
	}
	if !paramCorrectos || newUser.User == "" || newUser.Pass == "" || newUser.Grp == "" {
		salida.WriteString(fmt.Sprintf("MKUSR Error: Faltan parámetros obligatorios."))
		return salida.String(), nil
	}

	// --Validar que ninguno exceda 10 caracteres--
	if len(newUser.User) >= 12 {
		salida.WriteString(fmt.Sprintf("MKUSR Error: El nombre del usuario excede 10 caracteres."))
		return salida.String(), nil
	}
	if len(newUser.Pass) >= 10 {
		salida.WriteString(fmt.Sprintf("MKUSR Error: La contraseña excede 10 caracteres."))
		return salida.String(), nil
	}
	if len(newUser.Grp) >= 10 {
		salida.WriteString(fmt.Sprintf("MKUSR Error: El nombre del grupo excede 10 caracteres."))
		return salida.String(), nil
	}

	// Verificar que se tenga asignada la partición de la sesión.
	if session.PartitionID == "" {
		salida.WriteString(fmt.Sprintf("MKUSR Error: No se encontró la partición de la sesión actual."))
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
		salida.WriteString(fmt.Sprintf("MKUSR Error: No se encontró la partición montada para la sesión."))
		return salida.String(), nil
	}

	// Abrir el disco de la partición.
	disco, err := Herramientas.OpenFile(pathDico)
	if err != nil {
		salida.WriteString(fmt.Sprintf("MKUSR Error al abrir el disco:", err))
		return salida.String(), nil
	}
	defer disco.Close()

	// Leer el MBR.
	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		salida.WriteString(fmt.Sprintf("MKUSR Error al leer MBR:", err))
		return salida.String(), err
	}

	// Buscar la partición con session.PartitionID y leer su superbloque.
	var super Structs.Superblock
	var particion Structs.Partition
	encontrado := false
	for i := 0; i < 4; i++ {
		if Structs.GetId(string(mbr.Partitions[i].Id[:])) == session.PartitionID {
			particion = mbr.Partitions[i]
			if err := Herramientas.ReadObject(disco, &super, int64(particion.Start)); err != nil {
				salida.WriteString(fmt.Sprintf("MKUSR Error al leer superbloque:", err))
				return salida.String(), err
			}
			encontrado = true
			break
		}
	}
	if !encontrado {
		salida.WriteString(fmt.Sprintf("MKUSR Error: No se pudo encontrar la partición con id", session.PartitionID))
		return salida.String(), nil
	}

	// Leer el contenido actual del archivo users.txt.
	content, err := readUsersFile(disco, &super)
	if err != nil {
		salida.WriteString(fmt.Sprintf("MKUSR Error al leer users.txt:", err))
		return salida.String(), err
	}

	// Verificar que el grupo indicado exista.
	lines := strings.Split(content, "\n")
	groupExists := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.Split(trimmed, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		// Se espera que el registro de grupo tenga la forma: UID, G, groupName.
		if len(parts) >= 3 && parts[1] == "G" && parts[2] == newUser.Grp {
			groupExists = true
			break
		}
	}
	if !groupExists {
		salida.WriteString(fmt.Sprintf("MKUSR Error: El grupo", newUser.Grp, "no existe en la partición."))
		return salida.String(), nil
	}

	// Verificar que el usuario no exista ya.
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.Split(trimmed, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) >= 5 && parts[1] == "U" && parts[3] == newUser.User {
			salida.WriteString(fmt.Sprintf("MKUSR Error: El usuario", newUser.User, "ya existe."))
			return salida.String(), nil
		}
	}

	// Determinar el nuevo UID: buscar el máximo UID entre registros de usuario y sumar 1
	maxUID := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.Split(trimmed, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) >= 5 && parts[1] == "U" {
			uid, err := strconv.Atoi(parts[0])
			if err == nil && uid > maxUID {
				maxUID = uid
			}
		}
	}
	newUID := maxUID + 1 // Si solo existe el usuario root, newUID será 2

	// Preparar la nueva línea para el usuario.
	// Formato: "UID, U, groupName, userName, password\n"
	newUserLine := fmt.Sprintf("%d,U,%s,%s,%s\n", newUID, newUser.Grp, newUser.User, newUser.Pass)
	newContent := content
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += newUserLine

	// Actualizar el archivo users.txt en disco.
	//  Leer el folderBlock (directorio raíz) para ubicar el inodo de users.txt.
	var folderBlock Structs.Folderblock
	if err := Herramientas.ReadObject(disco, &folderBlock, int64(super.S_block_start)); err != nil {
		salida.WriteString(fmt.Sprintf("MKUSR Error al leer carpeta raíz:", err))
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
		salida.WriteString(fmt.Sprintf("MKUSR Error: users.txt no encontrado en la carpeta raíz"))
		return salida.String(), nil
	}

	//  Leer el inodo de users.txt.
	var userInode Structs.Inode
	inodeSize := int32(binary.Size(Structs.Inode{}))
	inodePosition := int64(super.S_inode_start) + int64(usersInode)*int64(inodeSize)
	if err := Herramientas.ReadObject(disco, &userInode, inodePosition); err != nil {
		salida.WriteString(fmt.Sprintf("MKUSR Error al leer inodo de users.txt:", err))
		return salida.String(), err
	}
	//  Actualizar el tamaño del archivo en el inodo.
	// Escribir el nuevo contenido completo en users.txt
	if err := session.WriteContentToFile(disco, &super, &userInode, newContent); err != nil {
		salida.WriteString(fmt.Sprintf("MKUSR Error al escribir el nuevo contenido en users.txt:", err))
		return salida.String(), err
	}
	userInode.I_size = int32(len(newContent))

	// Guardar el inodo actualizado con nuevo I_size
	if err := Herramientas.WriteObject(disco, userInode, inodePosition); err != nil {
		salida.WriteString(fmt.Sprintf("MKUSR Error al actualizar el inodo de users.txt:", err))
		return salida.String(), err
	}
	
	salida.WriteString(fmt.Sprintf("Usuario", newUser.User, "creado correctamente con UID", newUID))
	return salida.String(), nil
}
