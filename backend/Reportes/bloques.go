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

func GenerarReporteBlock(pathDisco string, pathReporte string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========REPORTE DE BLOQUES========\n")
	disco, err := os.Open(pathDisco)
	if err != nil {
		return "",fmt.Errorf("Error al abrir el disco: %v", err)
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		return  "",fmt.Errorf("Error al leer MBR: %v", err)
	}

	// Buscar la primera partición válida (no extendida)
	var particion Structs.Partition
	for _, p := range mbr.Partitions {
		if p.Size > 0 && string(p.Type[:]) != "E" {
			particion = p
			break
		}
	}
	if particion.Size == 0 {
		return  "",fmt.Errorf("No se encontró una partición válida")
	}

	// Leer el superbloque
	var sb Structs.Superblock
	if err := Herramientas.ReadObject(disco, &sb, int64(particion.Start)); err != nil {
		return  "",fmt.Errorf("Error al leer el superbloque: %v", err)
	}

	dot := strings.Builder{}
	dot.WriteString("digraph Bloques {\n")
	dot.WriteString("rankdir=LR;\n")
	dot.WriteString("node [shape=plaintext];\n")

	blockSize := int64(64) // Tamaño base

	// En la parte del bucle que recorre los bloques:
for i := int32(0); i < sb.S_blocks_count; i++ {
	byteVal := make([]byte, 1)
	if _, err := disco.ReadAt(byteVal, int64(sb.S_bm_block_start+i)); err != nil {
		continue
	}
	if byteVal[0] != 1 {
		continue
	}

	blockPos := int64(sb.S_block_start) + int64(i)*blockSize

	// ───── BLOQUE DE CARPETA ─────
	var fb Structs.Folderblock
	if err := Herramientas.ReadObject(disco, &fb, blockPos); err == nil {
		if fb.B_content[0].B_inodo != -1 || fb.B_content[1].B_inodo != -1 {
			dot.WriteString(fmt.Sprintf("block%d [label=<\n", i))
			dot.WriteString("<table border='1' cellborder='1' cellspacing='0'>\n")
			dot.WriteString("<tr><td colspan='2' bgcolor='#CCE5FF'><b>Bloque Carpeta</b></td></tr>\n")
			for _, content := range fb.B_content {
				name := strings.TrimRight(string(content.B_name[:]), "\x00")
				if name != "" && content.B_inodo != -1 {
					dot.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>\n", name, content.B_inodo))
				}
			}
			dot.WriteString("</table>>];\n")
			continue
		}
	}

	// ───── BLOQUE DE ARCHIVO ─────
	var archivo Structs.Fileblock
	if err := Herramientas.ReadObject(disco, &archivo, blockPos); err == nil {
		text := strings.TrimRight(string(archivo.B_content[:]), "\x00")
		if text != "" {
			dot.WriteString(fmt.Sprintf("block%d [label=<\n", i))
			dot.WriteString("<table border='1' cellborder='1' cellspacing='0'>\n")
			dot.WriteString("<tr><td bgcolor='#D4EDDA'><b>Bloque Archivo</b></td></tr>\n")
			dot.WriteString(fmt.Sprintf("<tr><td>%s</td></tr>\n", text))
			dot.WriteString("</table>>];\n")
			continue
		}
	}

	// ───── BLOQUE DE APUNTADORES ─────
	var pb Structs.Pointerblock
	if err := Herramientas.ReadObject(disco, &pb, blockPos); err == nil {
		dot.WriteString(fmt.Sprintf("block%d [label=<\n", i))
		dot.WriteString("<table border='1' cellborder='1' cellspacing='0'>\n")
		dot.WriteString("<tr><td colspan='4' bgcolor='#F8D7DA'><b>Bloque Apuntadores</b></td></tr>\n")
		for j := 0; j < len(pb.B_pointers); j += 4 {
			dot.WriteString("<tr>")
			for k := 0; k < 4 && j+k < len(pb.B_pointers); k++ {
				ptr := pb.B_pointers[j+k]
				if ptr != -1 {
					dot.WriteString(fmt.Sprintf("<td>%d</td>", ptr))
					dot.WriteString(fmt.Sprintf("</tr>\nblock%d -> block%d;\n", i, ptr)) // Agregar conexión visual
				} else {
					dot.WriteString("<td>-1</td>")
				}
			}
			dot.WriteString("</tr>\n")
		}
		dot.WriteString("</table>>];\n")
		// Conexiones visuales
	for _, ptr := range pb.B_pointers {
		if ptr != -1 {
			dot.WriteString(fmt.Sprintf("block%d -> block%d;\n", i, ptr))
		}
	}
	continue
	}
}


	dot.WriteString("}\n")

	// Crear carpeta si no existe
	dir := filepath.Dir(pathReporte)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return  "",fmt.Errorf("Error al crear carpeta del reporte: %v", err)
		}
	}

	dotFile := strings.ReplaceAll(pathReporte, ".jpg", ".dot")
	if err := os.WriteFile(dotFile, []byte(dot.String()), 0644); err != nil {
		return  "", fmt.Errorf("Error al escribir .dot: %v", err)
	}

	cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", pathReporte)
	if err := cmd.Run(); err != nil {
		return  "",fmt.Errorf("Error al generar imagen con dot: %v", err)
	}

	salida.WriteString(fmt.Sprintf("Reporte de Bloques generado exitosamente en:", pathReporte))
	return salida.String(), nil
}

