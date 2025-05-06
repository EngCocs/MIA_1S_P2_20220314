package administacioncarparch

import (
	"backend/Herramientas"
	"backend/Structs"
	"backend/permiso"
	"backend/session"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	
)

func Mkfile(entrada []string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========MKFILE========")
	var file Structs.File
	file.Size = 0 // Default

	if !session.Active {
		salida.WriteString(fmt.Sprintf("MKFILE Error: No hay sesión activa."))
		return salida.String(), nil
	}

	// Procesar parámetros
	for _, parametro := range entrada[1:] {
		tmp := strings.TrimSpace(parametro)
		tmp = cleanPath(tmp)

		if strings.ToLower(tmp) == "r" {
			file.R = true
			continue
		}

		valores := strings.Split(tmp, "=")
		if len(valores) != 2 {
			salida.WriteString(fmt.Sprintf("MKFILE Error: Parámetro incorrecto:", tmp))
			return salida.String(), nil
		}
		key := strings.ToLower(valores[0])
		valor := strings.ReplaceAll(valores[1], "\"", "")

		switch key {
		case "path":
			file.Path = cleanPath(valor)//se limpia el path
		case "size":
			size, err := PositiveInt(valor)//se convierte el valor a entero
			if err != nil {
				salida.WriteString(fmt.Sprintf("MKFILE Error: Tamaño inválido."))
				return salida.String(), err
			}
			file.Size = size
		case "cont":
			file.Cont = cleanPath(valor)
		default:
			salida.WriteString(fmt.Sprintf("MKFILE Error: Parámetro desconocido:", key))
			return salida.String(), nil
		}
	}

	if file.Path == "" {
		salida.WriteString(fmt.Sprintf("MKFILE Error: Falta parámetro -path."))
		return salida.String(), nil
	}

	// Leer ruta de disco
	var discoPath string
	for _, m := range Structs.Montadas {
		if m.Id == session.PartitionID {
			discoPath = m.PathM
			break
		}
	}
	if discoPath == "" {
		salida.WriteString(fmt.Sprintf("MKFILE Error: Partición no montada."))
		return salida.String(), nil
	}

	disco, err := Herramientas.OpenFile(discoPath)
	if err != nil {
		salida.WriteString(fmt.Sprintf("MKFILE Error al abrir disco:", err))
		return salida.String(), err
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		salida.WriteString(fmt.Sprintf("MKFILE Error al leer MBR:", err))
		return salida.String(), err
	}
	//aqui se busca la particion que se va a utilizar
	var super Structs.Superblock
	var partStart int64
	for _, part := range mbr.Partitions {
		if Structs.GetId(string(part.Id[:])) == session.PartitionID {
			partStart = int64(part.Start)
			Herramientas.ReadObject(disco, &super, partStart)
			break
		}
	}

	// Separar carpeta padre y nombre del archivo asi 
dirPath, fileName := cleanPaths(file.Path)
	if fileName == "" {
		salida.WriteString(fmt.Sprintf("MKFILE Error: El path debe incluir un nombre de archivo."))
		return salida.String(), nil
	}
	

	// Aqui validamos o creamos la ruta padre
	// Aqui validamos o creamos la ruta padre
	steps := strings.Split(dirPath, "/")[1:]
	inodoActual := int32(0)
	
	for i, carpeta := range steps {
		id := permiso.BuscarEnCarpeta(carpeta, inodoActual, super, disco)
		if id >= 0 {
			inodoActual = id
			continue
		} else if id == -2 {
			salida.WriteString(fmt.Sprintf("MKFILE Error: '%s' ya existe pero no es una carpeta.\n", carpeta))
			return salida.String(), nil
		} else if !file.R {
			salida.WriteString(fmt.Sprintf("MKFILE Error: Carpeta '%s' no existe. Usa -r para crearla.\n", carpeta))
			return salida.String(), nil
		} else {
			// Crear las carpetas faltantes desde aquí
			for j := i; j < len(steps); j++ {
				inodoActual = permiso.CreateFolder(inodoActual, steps[j], &super, disco, partStart)
			}
			// Leer nuevamente el superbloque actualizado tras crear carpetas
			// if err := Herramientas.ReadObject(disco, &super, partStart); err != nil {
			// 	salida.WriteString("MKFILE Error al recargar superbloque después de mkdir.")
			// 	return salida.String(), err
			// }

			break
		}
	}
	
	




	// Crear archivo (reserva del inodo)

fmt.Printf("⚠️ PREVIO - Mkfile va a usar inodo %d\n", super.S_first_ino)

nInodo := super.S_first_ino
fmt.Printf("⚠️ PREVIO - Mkfile va a usar inodo %d\n", super.S_first_ino)
// Reservar en bitmap
disco.WriteAt([]byte{1}, int64(super.S_bm_inode_start)+int64(nInodo))

super.S_first_ino++
super.S_free_inodes_count--

// Guardar superbloque de inmediato para evitar sobrescritura futura
if err := Herramientas.WriteObject(disco, super, partStart); err != nil {
	return salida.String(), err
}


// Inicializar inodo
nuevoInodo := Structs.Inode{}
nuevoInodo.I_type[0] = '1'
copy(nuevoInodo.I_perm[:], "664")
nuevoInodo.I_uid = 1
nuevoInodo.I_gid = 1
fecha := time.Now().Format("02/01/2006 15:04")
copy(nuevoInodo.I_ctime[:], fecha)
copy(nuevoInodo.I_mtime[:], fecha)
copy(nuevoInodo.I_atime[:], fecha)
fmt.Printf("DEBUG inico → inodoActual: %d, creando archivo '%s' con inodo: %d, tipo: %v\n", inodoActual, fileName, nInodo, nuevoInodo.I_type[0])

for i := 0; i < 15; i++ {
	nuevoInodo.I_block[i] = -1
}

// Escribir inodo en disco antes de asignar contenido (previene lecturas basura)
posInodo := int64(super.S_inode_start) + int64(nInodo)*int64(binary.Size(Structs.Inode{}))
fmt.Printf("📁 Mkfile: usando inodo %d para archivo '%s'\n", nInodo, fileName)


// Obtener contenido
var contenido string
if file.Cont != "" {
	data, err := os.ReadFile(file.Cont)
	if err != nil {
		salida.WriteString(fmt.Sprintf("MKFILE Error: Archivo fuente -cont no encontrado: %v", err))
		return salida.String(), err
	}
	if file.Size > 0 && int32(len(data)) > file.Size {
		data = data[:file.Size] // solo recorta si se pasa
	}
	
	nuevoInodo.I_size = int32(len(data))

	contenido = string(data)
} else {
	contenido = generarContenido(file.Size)
	nuevoInodo.I_size = int32(len(contenido))
}

fmt.Printf("DEBUG ultimo desopues de todo el contenido a escribir para '%s': [%v]\n", fileName, []byte(contenido))




// Escribir contenido al archivo (actualiza bloques del inodo)
if err := session.WriteContentToFile(disco, &super, &nuevoInodo, contenido); err != nil {
	salida.WriteString(fmt.Sprintf("MKFILE Error al escribir archivo: %v", err))
	return salida.String(), err
}
fmt.Printf("DEBUG medio → inodoActual: %d, creando archivo '%s' con inodo: %d, tipo: %v\n", inodoActual, fileName, nInodo, nuevoInodo.I_type[0])
fmt.Printf("Escribiendo inodo %d con tipo: %v, tamaño: %d\n", nInodo, nuevoInodo.I_type[0], nuevoInodo.I_size)

// Reescribir el inodo ya con bloques y tamaño actualizado


fmt.Printf("DEBUG MKFILE -> Escribiendo inodo en pos: %d con bloques: %v\n", posInodo, nuevoInodo.I_block)

if err := Herramientas.WriteObject(disco, nuevoInodo, posInodo); err != nil {
	salida.WriteString(fmt.Sprintf("MKFILE Error al guardar inodo actualizado: %v", err))
	return salida.String(), err
}

fmt.Printf("→ Insertando '%s' en inodo padre %d con nuevo inodo %d\n", fileName, inodoActual, nInodo)

// Insertar entrada en carpeta padre
permiso.AgregarEntradaACarpeta(disco, &super, inodoActual, fileName, nInodo)

// Guardar superbloque actualizado
// savePos := int64(super.S_inode_start) - int64(binary.Size(Structs.Superblock{}))
// Herramientas.WriteObject(disco, super, savePos)
// Guardar superbloque actualizado donde realmente está
if err := Herramientas.WriteObject(disco, super, partStart); err != nil {
    return salida.String(), err
}


fmt.Printf("DEBUG ultimo → inodoActual: %d, creando archivo '%s' con inodo: %d, tipo: %v\n", inodoActual, fileName, nInodo, nuevoInodo.I_type[0])

salida.WriteString(fmt.Sprintf("MKFILE: Archivo '%s' creado correctamente.\n", fileName))
return salida.String(), nil

}

// Limpia el path de caracteres especiales
func cleanPaths(fullPath string) (string, string) {
	fullPath = strings.TrimRight(fullPath, "/")
	parts := strings.Split(fullPath, "/")
	if len(parts) <= 1 {
		return "/", parts[0]
	}
	return strings.Join(parts[:len(parts)-1], "/"), parts[len(parts)-1]
}
//genera el contenido del archivo
func generarContenido(size int32) string {
	var builder strings.Builder
	numeros := "0123456789"//Esto lo uso como una plantilla de contenido
	for int32(builder.Len()) < size {
		builder.WriteString(numeros)
	}
	return builder.String()[:size]
}
//Esta funcion  recibe un string y lo convierte a un entero de 32 bits
func PositiveInt(s string) (int32, error) {
	valor, err := strconv.Atoi(s)
	if err != nil || valor < 0 {
		return 0, fmt.Errorf("número inválido")
	}
	return int32(valor), nil
}


