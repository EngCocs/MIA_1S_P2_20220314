package commands


import (
	"fmt"
	"strings"

	"backend/Structs"
)

// Mounted muestra todas las particiones montadas en memoria.
// No recibe parámetros y lista los IDs de cada partición montada.
func Mounted(entrada []string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========MOUNTED========")
	// Verifica si hay alguna partición montada.
	if len(Structs.Montadas) == 0 {
		salida.WriteString(fmt.Sprintf("No hay particiones montadas en el sistema"))
		return salida.String(), nil
	}

	// Recorre el slice Montadas y extrae los IDs.
	var ids []string
	for _, montada := range Structs.Montadas {
		ids = append(ids, montada.Id)
	}

	// Une los IDs con comas y los muestra.
	resultado := strings.Join(ids, ", ")
	salida.WriteString(fmt.Sprintf("Particiones montadas:", resultado))
	return salida.String(), nil
}