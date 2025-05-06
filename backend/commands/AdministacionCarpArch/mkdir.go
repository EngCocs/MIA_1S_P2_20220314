package administacioncarparch

import (
	"backend/Herramientas"
	"backend/Structs"
	"backend/permiso"
	"backend/session"
	"unicode"
	//"encoding/binary"
	"fmt"
	//"os"
	"strings"
)

func Mkdir(entrada []string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========MKDIR========")
	fmt.Println("MKDIR")

	var path string
	var crearRecursivo bool = false

	if !session.Active {
		salida.WriteString(fmt.Sprintf("MKDIR Error: No hay sesión activa."))
		return salida.String(), nil
	}

	// Procesar parámetros
	for _, parametro := range entrada[1:] {
		tmp := strings.TrimSpace(parametro)
		valores := strings.Split(tmp, "=")

		// Revisar si es el flag -p sin valor
		if strings.ToLower(tmp) == "p" {
			crearRecursivo = true
			continue
		}

		if len(valores) != 2 {
			salida.WriteString(fmt.Sprintf("MKDIR Error: Parámetro incorrecto:", tmp))
			return salida.String(), nil
		}

		clave := strings.ToLower(valores[0])
		valor := strings.ReplaceAll(valores[1], "\"", "")
		if clave == "path" {
			path = cleanPath(valor) // valor ya viene limpio y seguro

		} else {
			salida.WriteString(fmt.Sprintf("MKDIR Error: Parámetro desconocido:", clave))
			return salida.String(), nil
		}
	}

	if path == "" {
		salida.WriteString(fmt.Sprintf("MKDIR Error: El parámetro -path es obligatorio."))
		return salida.String(), nil
	}

	// Obtener ruta del disco
	var pathDisco string
	for _, m := range Structs.Montadas {
		if m.Id == session.PartitionID {
			pathDisco = m.PathM
			break
		}
	}
	if pathDisco == "" {
		salida.WriteString(fmt.Sprintf("MKDIR Error: No se encontró la partición montada."))
		return salida.String(), nil
	}

	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		salida.WriteString(fmt.Sprintf("MKDIR Error al abrir el disco:", err))
		return salida.String(), err
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		salida.WriteString(fmt.Sprintf("MKDIR Error al leer MBR:", err))
		return salida.String(), err
	}

	var super Structs.Superblock
	var particion Structs.Partition
	encontrado := false
	for i := 0; i < 4; i++ {
		if Structs.GetId(string(mbr.Partitions[i].Id[:])) == session.PartitionID {
			particion = mbr.Partitions[i]
			if err := Herramientas.ReadObject(disco, &super, int64(particion.Start)); err != nil {
				salida.WriteString(fmt.Sprintf("MKDIR Error al leer superbloque:", err))
				return salida.String(), err
			}
			encontrado = true
			break
		}
	}
	if !encontrado {
		salida.WriteString(fmt.Sprintf("MKDIR Error: No se encontró la partición de la sesión actual."))
		return salida.String(), nil
	}

	// Validar existencia de carpetas padre y crear si es necesario
	steps := strings.Split(path, "/")[1:]
	idActual := int32(0)

	for i, dir := range steps {
		idEncontrado := permiso.SearchInode(idActual, "/"+dir, super, disco)
		if idEncontrado != idActual {
			idActual = idEncontrado
		} else {
			if !crearRecursivo {
				fmt.Printf("MKDIR Error: La carpeta padre '%s' no existe. Usa -p para crearla.\n", dir)
				return salida.String(), nil
			}
			for j := i; j < len(steps); j++ {
				idActual = permiso.CreateFolder(idActual, steps[j], &super, disco,int64(particion.Start))
				fmt.Printf("MKDIR: Creando carpeta '%s'\n", steps[j])

			}
			break
		}
	}
	
	permiso.VerificarContenidoCarpeta(disco, super, idActual)

	salida.WriteString(fmt.Sprintf("MKDIR: Ruta creada correctamente."))
	return salida.String(), nil
}

//eata funcion limpia el path de caracteres invisibles y no imprimibles :)
func cleanPath(p string) string {
	// Remueve caracteres invisibles y no imprimibles como \u200B, \uFEFF, etc.
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) && r != '\u200B' && r != '\uFEFF' {
			return r
		}
		return -1
	}, p)
}

