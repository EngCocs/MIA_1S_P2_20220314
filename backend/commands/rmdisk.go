package commands

import (
	"fmt"
	"os"
	"strings"
)


func Rmdisk(entrada []string) (string, error) {
	var salida strings.Builder
	salida.WriteString(fmt.Sprintf("========RMDISK========"))
	var path string //Path del Disco
	paramC := true


	for _, parametro := range entrada[1:] {
		tmp := strings.TrimRight(parametro, " ")
		valores := strings.Split(tmp, "=")

		if len(valores) != 2 {
			salida.WriteString(fmt.Sprintf("ERROR RMDISK, valor desconocido de parametros ", tmp))
			return salida.String(), nil
		}

		//******************* PATH *************
		if strings.ToLower(valores[0]) == "path" {
			path = strings.ReplaceAll(valores[1], "\"", "")
			_, err := os.Stat(path)
			if os.IsNotExist(err) {
				salida.WriteString(fmt.Sprintf("ERROR RMDISK: El disco no existe"))
				paramC = false
				break // Terminar el bucle porque encontramos un nombre único
			}
			
		
	   }
	}
	if paramC {
		if path != "" { // si el path no esta vacio
			salida.WriteString(fmt.Sprintf("Disco :", path, " Eliminado"))
			err := os.Remove(path)//eliminar el disco
			
			if err != nil {
				salida.WriteString(fmt.Sprintf("ERROR RMDISK: No se pudo eliminar el disco"))
			}
		} else {
			salida.WriteString(fmt.Sprintf("ERROR RMDISK: No se encontro el parametro PATH"))
		}

	}
	return salida.String(), nil
}
