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

func GenerarReporteSB(pathDisco string, pathReporte string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========SB========\n")
	disco, err := os.Open(pathDisco)
	if err != nil {
		return "",fmt.Errorf("Error al abrir el disco: %v", err)
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		return "",fmt.Errorf("Error al leer el MBR: %v", err)
	}

	// Tomamos la primera partición válida
	var particion Structs.Partition
	for _, p := range mbr.Partitions {
		if p.Size > 0 && string(p.Type[:]) != "E" {
			particion = p
			break
		}
	}
	if particion.Size == 0 {
		return "",fmt.Errorf("No se encontró partición válida")
	}

	// Obtener el contenido de la tabla con RepSB
	contenidoTabla := Structs.RepSB(particion, disco)

	// Crear el archivo DOT
	dot := strings.Builder{}
	dot.WriteString("digraph sb {\n")
	dot.WriteString("node [shape=plaintext];\n")
	dot.WriteString("sb [label=<\n")
	dot.WriteString("<table border='1' cellborder='1' cellspacing='0'>\n")
	dot.WriteString("<tr><td colspan='2' bgcolor='#004466'><font color='white'><b>Reporte de SUPERBLOQUE</b></font></td></tr>\n")
	dot.WriteString(contenidoTabla)
	dot.WriteString("</table>>];\n")
	dot.WriteString("}\n")

	// Crear carpeta si no existe
	dir := filepath.Dir(pathReporte)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return "",fmt.Errorf("Error al crear carpeta del reporte: %v", err)
		}
	}

	// Guardar el .dot
	dotPath := strings.ReplaceAll(pathReporte, ".jpg", ".dot")
	if err := os.WriteFile(dotPath, []byte(dot.String()), 0644); err != nil {
		return "",fmt.Errorf("Error al escribir el archivo DOT: %v", err)
	}

	// Ejecutar Graphviz
	cmd := exec.Command("dot", "-Tjpg", dotPath, "-o", pathReporte)
	if err := cmd.Run(); err != nil {
		return "",fmt.Errorf("Error al generar imagen con dot: %v", err)
	}

	salida.WriteString(fmt.Sprintf("Reporte SB generado exitosamente:", pathReporte))
	return salida.String(), nil
}
