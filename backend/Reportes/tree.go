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

func GenerarReporteTree(pathDisco string, pathReporte string) error {
	disco, err := os.Open(pathDisco)
	if err != nil {
		return fmt.Errorf("Error al abrir disco: %v", err)
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		return fmt.Errorf("Error al leer MBR: %v", err)
	}

	var particion Structs.Partition
	for _, p := range mbr.Partitions {
		if p.Size > 0 && string(p.Type[:]) != "E" {
			particion = p
			break
		}
	}
	if particion.Size == 0 {
		return fmt.Errorf("No se encontró partición válida")
	}

	var sb Structs.Superblock
	if err := Herramientas.ReadObject(disco, &sb, int64(particion.Start)); err != nil {
		return fmt.Errorf("Error al leer superbloque: %v", err)
	}

	dot := strings.Builder{}
	dot.WriteString("digraph Tree {\n")
	dot.WriteString("rankdir=LR;\n")
	dot.WriteString("node [shape=plaintext];\n")

	visitados := make(map[int32]bool)
	err = recorrerInodo(&dot, disco, sb, 0, 0, visitados)
	if err != nil {
		return err
	}

	// Crear carpeta si no existe
	dir := filepath.Dir(pathReporte)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("Error al crear carpeta del reporte: %v", err)
		}
	}

	dot.WriteString("}\n")
	dotFile := strings.ReplaceAll(pathReporte, ".jpg", ".dot")
	if err := os.WriteFile(dotFile, []byte(dot.String()), 0644); err != nil {
		return fmt.Errorf("Error al guardar .dot: %v", err)
	}

	cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", pathReporte)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error al generar imagen: %v", err)
	}

	fmt.Println("Reporte TREE generado exitosamente:", pathReporte)
	return nil
}

func recorrerInodo(dot *strings.Builder, disco *os.File, sb Structs.Superblock, numInodo int32, nivel int, visitados map[int32]bool) error {
	if visitados[numInodo] {
		return nil
	}
	visitados[numInodo] = true

	inodoPos := int64(sb.S_inode_start) + int64(numInodo)*int64(sb.S_inode_size)
	var inodo Structs.Inode
	if err := Herramientas.ReadObject(disco, &inodo, inodoPos); err != nil {
		return fmt.Errorf("Error al leer inodo %d: %v", numInodo, err)
	}

	tipoTexto := "Archivo"
	if inodo.I_type[0] == '0' {
		tipoTexto = "Carpeta"
	}

	dot.WriteString(fmt.Sprintf("inodo%d [label=<\n", numInodo))
	dot.WriteString("<table border='1' cellborder='1' cellspacing='0'>\n")
	dot.WriteString(fmt.Sprintf("<tr><td colspan='2'><b>Inodo %d (%s)</b></td></tr>\n", numInodo, tipoTexto))
	dot.WriteString(fmt.Sprintf("<tr><td>i_type</td><td>%d</td></tr>\n", inodo.I_type[0]-48))
	dot.WriteString("<tr><td>op0</td><td>apuntador directo</td></tr>\n")
	dot.WriteString(fmt.Sprintf("<tr><td>ap0</td><td>%d</td></tr>\n", inodo.I_block[0]))
	dot.WriteString("<tr><td>op1</td><td>apuntador indirecto</td></tr>\n")
	dot.WriteString(fmt.Sprintf("<tr><td>ap13</td><td>%d</td></tr>\n", inodo.I_block[13]))
	dot.WriteString("<tr><td>op2</td><td>doble indirecto</td></tr>\n")
	dot.WriteString(fmt.Sprintf("<tr><td>ap14</td><td>%d</td></tr>\n", inodo.I_block[14]))
	perm := strings.Trim(string(inodo.I_perm[:]), "\x00")
	dot.WriteString(fmt.Sprintf("<tr><td>i_perm</td><td>%s</td></tr>\n", perm))
	dot.WriteString("</table>>];\n")

	if inodo.I_type[0] == '0' {
		for _, bloque := range inodo.I_block[:12] {
			if bloque == -1 {
				continue
			}
			blockPos := int64(sb.S_block_start) + int64(bloque)*int64(sb.S_block_size)
			var fb Structs.Folderblock
			if err := Herramientas.ReadObject(disco, &fb, blockPos); err != nil {
				continue
			}
			dot.WriteString(fmt.Sprintf("block%d [label=<\n", bloque))
			dot.WriteString("<table border='1' cellborder='1' cellspacing='0'>\n")
			dot.WriteString(fmt.Sprintf("<tr><td colspan='2'><b>Bloque Carpeta %d</b></td></tr>\n", bloque))
			for _, content := range fb.B_content {
				name := Structs.GetB_name(string(content.B_name[:]))
				if name != "" && content.B_inodo != -1 {
					dot.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>\n", name, content.B_inodo))
				}
			}
			dot.WriteString("</table>>];\n")
			dot.WriteString(fmt.Sprintf("inodo%d -> block%d;\n", numInodo, bloque))

			for _, content := range fb.B_content {
				name := Structs.GetB_name(string(content.B_name[:]))
				if name != "" && name != "." && name != ".." && content.B_inodo != -1 {
					dot.WriteString(fmt.Sprintf("block%d -> inodo%d;\n", bloque, content.B_inodo))
					recorrerInodo(dot, disco, sb, content.B_inodo, nivel+1, visitados)
				}
			}
		}
	} else {
		for _, bloque := range inodo.I_block[:12] {
			if bloque == -1 {
				continue
			}
			blockPos := int64(sb.S_block_start) + int64(bloque)*int64(sb.S_block_size)
			var fileBlock Structs.Fileblock
			if err := Herramientas.ReadObject(disco, &fileBlock, blockPos); err != nil {
				continue
			}
			content := Structs.GetB_content(string(fileBlock.B_content[:]))
			dot.WriteString(fmt.Sprintf("block%d [label=<\n", bloque))
			dot.WriteString("<table border='1' cellborder='1' cellspacing='0'>\n")
			dot.WriteString(fmt.Sprintf("<tr><td><b>Bloque Archivo %d</b></td></tr>\n", bloque))
			dot.WriteString(fmt.Sprintf("<tr><td>%s</td></tr>\n", content))
			dot.WriteString("</table>>];\n")
			dot.WriteString(fmt.Sprintf("inodo%d -> block%d;\n", numInodo, bloque))
		}
	}

	if inodo.I_block[13] != -1 {
		indirectPos := int64(sb.S_block_start) + int64(inodo.I_block[13])*int64(sb.S_block_size)
		var ptrBlock Structs.Pointerblock
		if err := Herramientas.ReadObject(disco, &ptrBlock, indirectPos); err == nil {
			dot.WriteString(fmt.Sprintf("indirect%d [label=<\n", inodo.I_block[13]))
			dot.WriteString("<table border='1' cellborder='1' cellspacing='0'>\n")
			dot.WriteString(fmt.Sprintf("<tr><td colspan='1'><b>Bloque Indirecto %d</b></td></tr>\n", inodo.I_block[13]))
			for i := 0; i < 2; i++ {
				val := ptrBlock.B_pointers[i]
				dot.WriteString(fmt.Sprintf("<tr><td>%d</td></tr>\n", val))
			}
			dot.WriteString("</table>>];\n")
			dot.WriteString(fmt.Sprintf("inodo%d -> indirect%d;\n", numInodo, inodo.I_block[13]))
		}
	}
	return nil
}



