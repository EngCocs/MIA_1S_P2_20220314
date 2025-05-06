package reportes

import (
	"backend/Herramientas"
	"backend/Structs"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GenerarReporteMBR(pathDisco string, pathReporte string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========MBR========\n")
	disco, err := os.Open(pathDisco)
	if err != nil {
		return "",fmt.Errorf("Error al abrir el disco: %v", err)
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		return "",fmt.Errorf("Error al leer el MBR: %v", err)
	}

	dot := strings.Builder{}
	dot.WriteString("digraph MBR {\n")
	dot.WriteString("node [shape=plaintext]\n")
	dot.WriteString("ReporteMBR [label=<\n")
	dot.WriteString("<table border='1' cellborder='1' cellspacing='0' style='rounded,filled' bgcolor='#e0e0e0'>\n")
	dot.WriteString("<tr><td colspan='2'><b>REPORTE DE MBR</b></td></tr>\n")
	dot.WriteString(fmt.Sprintf("<tr><td><b>mbr_tamano</b></td><td>%d</td></tr>\n", mbr.MbrSize))
	dot.WriteString(fmt.Sprintf("<tr><td><b>mbr_fecha_creacion</b></td><td>%s</td></tr>\n", string(mbr.FechaC[:])))
	dot.WriteString(fmt.Sprintf("<tr><td><b>mbr_disk_signature</b></td><td>%d</td></tr>\n", mbr.Id))

	for _, part := range mbr.Partitions {
		if part.Size > 0 {
			dot.WriteString("<tr><td colspan='2' bgcolor='#e3f2fd'><b>Partition</b></td></tr>\n")
			//dot.WriteString("<tr><td colspan='2' bgcolor='#e3f2fd'><b>Partition</b></td></tr>\n")

			// Mostrar part_status como 1 o 0
			status := "0"
			if part.Status[0] == 1 || part.Status[0] == '1' || part.Status[0] == 'A' {
				status = "1"
			}

			dot.WriteString(fmt.Sprintf("<tr><td><b>part_status</b></td><td>%s</td></tr>\n", status))
			dot.WriteString(fmt.Sprintf("<tr><td><b>part_type</b></td><td>%s</td></tr>\n", string(part.Type[:])))
			dot.WriteString(fmt.Sprintf("<tr><td><b>part_fit</b></td><td>%s</td></tr>\n", string(part.Fit[:])))
			dot.WriteString(fmt.Sprintf("<tr><td><b>part_start</b></td><td>%d</td></tr>\n", part.Start))
			dot.WriteString(fmt.Sprintf("<tr><td><b>part_size</b></td><td>%d</td></tr>\n", part.Size))
			dot.WriteString(fmt.Sprintf("<tr><td><b>part_name</b></td><td>%s</td></tr>\n", Structs.GetName(string(part.Name[:]))))

			// Si es extendida, buscar particiones lógicas
			if string(part.Type[:]) == "e" || string(part.Type[:]) == "E" {
				dot.WriteString("<tr><td colspan='2'><b>Particiones Lógicas</b></td></tr>\n")

				pos := int64(part.Start)
				for {
					var ebr Structs.EBR
					if err := Herramientas.ReadObject(disco, &ebr, pos); err != nil {
						break
					}

					// Mostrar solo si tiene contenido válido
					if ebr.Size > 0 {
						statusEBR := "0"
						if ebr.Status[0] == 1 || ebr.Status[0] == '1' || ebr.Status[0] == 'A' {
							statusEBR = "1"
						}

						dot.WriteString("<tr><td colspan='2' bgcolor='#fdd'><b>Partición Lógica</b></td></tr>\n")
						dot.WriteString(fmt.Sprintf("<tr><td><b>part_status</b></td><td>%s</td></tr>\n", statusEBR))
						dot.WriteString(fmt.Sprintf("<tr><td><b>part_next</b></td><td>%d</td></tr>\n", ebr.Next))
						dot.WriteString(fmt.Sprintf("<tr><td><b>part_fit</b></td><td>%s</td></tr>\n", string(ebr.Fit[:])))
						dot.WriteString(fmt.Sprintf("<tr><td><b>part_start</b></td><td>%d</td></tr>\n", ebr.Start))
						dot.WriteString(fmt.Sprintf("<tr><td><b>part_size</b></td><td>%d</td></tr>\n", ebr.Size))
						dot.WriteString(fmt.Sprintf("<tr><td><b>part_name</b></td><td>%s</td></tr>\n", Structs.GetName(string(ebr.Name[:]))))
					}

					if ebr.Next == -1 || ebr.Next == 0 {
						break
					}
					pos = int64(ebr.Next)
				}
			}
		}
	}

	dot.WriteString("</table>>];\n}\n")

	// Crear carpeta si no existe
	dir := filepath.Dir(pathReporte)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return "",fmt.Errorf("Error al crear carpeta del reporte: %v", err)
		}
	}

	// Generar archivo .dot
	dotFile := strings.ReplaceAll(pathReporte, ".jpg", ".dot")
	if err := os.WriteFile(dotFile, []byte(dot.String()), 0644); err != nil {
		return "",fmt.Errorf("error al escribir archivo dot: %v", err)
	}

	// Generar imagen .jpg
	cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", pathReporte)
	if err := cmd.Run(); err != nil {
		return "",fmt.Errorf("error al generar imagen con dot: %v", err)
	}

	salida.WriteString(fmt.Sprintf("Reporte MBR generado exitosamente en:", pathReporte))
	return salida.String(), nil
}



