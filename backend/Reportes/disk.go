package reportes

import (
	"backend/Herramientas"
	"backend/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func GenerarReporteDISK(pathDisco string, pathReporte string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========DISK========\n")
	disco, err := os.Open(pathDisco)
	if err != nil {
		return "",fmt.Errorf("Error al abrir el disco: %v", err)
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		return "",fmt.Errorf("Error al leer el MBR: %v", err)
	}

	totalDisco := float64(mbr.MbrSize)
	ultimaPos := int32(binary.Size(mbr))

	var particionesOrdenadas []Structs.Partition
	for _, p := range mbr.Partitions {
		if p.Size > 0 {
			particionesOrdenadas = append(particionesOrdenadas, p)
		}
	}
	sort.Slice(particionesOrdenadas, func(i, j int) bool {
		return particionesOrdenadas[i].Start < particionesOrdenadas[j].Start
	})
	// Crear el label para la grafica del disk
	// Se usa strings.Builder para construir la cadena de manera eficiente y evitar la concatenacion repetida
	label := strings.Builder{}
	label.WriteString("MBR")
	// Se agrega el tamaño total del disco
	for _, part := range particionesOrdenadas {
		if part.Start > ultimaPos {
			libre := float64(part.Start-ultimaPos) * 100 / totalDisco
			label.WriteString(fmt.Sprintf("| Libre (%.0f%%)", libre))
		}
		// Se agrega la particion actual y su porcentaje
		porcentaje := float64(part.Size) * 100 / totalDisco
		tipo := "Primaria"
		// Se verifica si la particion es extendida
		if string(part.Type[:]) == "E" {
			tipo = "Extendida"
			label.WriteString(fmt.Sprintf("| %s (%.0f%%)", tipo, porcentaje))

			pos := int64(part.Start)
			for {
				var ebr Structs.EBR
				if err := Herramientas.ReadObject(disco, &ebr, pos); err != nil {
					break
				}
				if ebr.Size <= 0 {
					break
				}
				label.WriteString("| EBR")
				porcLogica := float64(ebr.Size) * 100 / totalDisco
				label.WriteString(fmt.Sprintf("| Logica (%.0f%%)", porcLogica))
				if ebr.Next == -1 || ebr.Next == 0 {
					break
				}
				pos = int64(ebr.Next)
			}
		} else {
			label.WriteString(fmt.Sprintf("| %s (%.0f%%)", tipo, porcentaje))
		}
		ultimaPos = part.Start + part.Size
	}
	// Se agrega el espacio libre al final del disco
	if ultimaPos < mbr.MbrSize {
		libreFinal := float64(mbr.MbrSize-ultimaPos) * 100 / totalDisco
		label.WriteString(fmt.Sprintf("| Libre (%.0f%%)", libreFinal))
	}

	dot := strings.Builder{}
	dot.WriteString("digraph DISK {\n")
	dot.WriteString("node [shape=record];\n")
	dot.WriteString(fmt.Sprintf("structDisk [label=\"%s\"]\n", label.String()))
	dot.WriteString("}\n")

	// Crear carpeta si no existe
	dir := filepath.Dir(pathReporte)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return "", fmt.Errorf("Error al crear carpeta del reporte: %v", err)
		}
	}

	dotFile := strings.ReplaceAll(pathReporte, ".jpg", ".dot")
	if err := os.WriteFile(dotFile, []byte(dot.String()), 0644); err != nil {
		return "",fmt.Errorf("Error al escribir archivo dot: %v", err)
	}

	cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", pathReporte)
	if err := cmd.Run(); err != nil {
		return "",fmt.Errorf("Error al generar imagen con dot: %v", err)
	}

	salida.WriteString(fmt.Sprintf("Reporte DISK generado exitosamente en:", pathReporte))
	return salida.String(), nil
}


