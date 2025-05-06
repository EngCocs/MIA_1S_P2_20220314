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

func GenerarReporteInode(pathDisco string, pathReporte string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========REPORTE DE INODOS========\n")
	disco, err := os.Open(pathDisco)
	if err != nil {
		return "",fmt.Errorf("Error al abrir el disco: %v", err)
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		return "",fmt.Errorf("Error al leer el MBR: %v", err)
	}

	var particion Structs.Partition
	for _, p := range mbr.Partitions {
		if p.Size > 0 && string(p.Type[:]) != "E" {
			particion = p
			break
		}
	}
	if particion.Size == 0 {
		return "",fmt.Errorf("No se encontró una partición válida para generar el reporte de inodos")
	}

	var sb Structs.Superblock
	if err := Herramientas.ReadObject(disco, &sb, int64(particion.Start)); err != nil {
		return "",fmt.Errorf("Error al leer el Superbloque: %v", err)
	}
	fmt.Println("→ Verificación previa al REP")
fmt.Println("S_inodes_count:", sb.S_inodes_count)
fmt.Println("S_blocks_count:", sb.S_blocks_count)
fmt.Println("S_free_inodes_count:", sb.S_free_inodes_count)
fmt.Println("S_free_blocks_count:", sb.S_free_blocks_count)

	// VERIFICACIÓN DEL BITMAP DE INODOS Y BLOQUES
	fmt.Println("======= Bitmap de Inodos (1 = usado) =======")
	for i := int32(0); i < sb.S_inodes_count; i++ {
		byteVal := make([]byte, 1)
		if _, err := disco.ReadAt(byteVal, int64(sb.S_bm_inode_start+i)); err != nil {
			continue
		}
		//fmt.Printf("Inodo[%d]: %d\n", i, byteVal[0])
	}
	fmt.Println("======= Bitmap de Bloques (1 = usado) =======")
	for i := int32(0); i < sb.S_blocks_count; i++ {
		byteVal := make([]byte, 1)
		if _, err := disco.ReadAt(byteVal, int64(sb.S_bm_block_start+i)); err != nil {
			continue
		}
		//fmt.Printf("Bloque[%d]: %d\n", i, byteVal[0])
	}

	dot := strings.Builder{}
	dot.WriteString("digraph Inodos {\n")
	dot.WriteString("node [shape=plaintext];\n")

	prevInodo := ""
	for i := int32(0); i < sb.S_inodes_count; i++ {
		byteVal := make([]byte, 1)
		if _, err := disco.ReadAt(byteVal, int64(sb.S_bm_inode_start+i)); err != nil {
			continue
		}
		//fmt.Printf("Bitmap inode[%d]: %d\n", i, byteVal[0])

		if byteVal[0] != '1' && byteVal[0] != 1 {
			continue
		}

		offset := int64(sb.S_inode_start) + int64(i*sb.S_inode_size)
		var inodo Structs.Inode
		if err := Herramientas.ReadObject(disco, &inodo, offset); err != nil {
			continue
		}

		nombreInodo := fmt.Sprintf("Inodo%d", i+1)
		dot.WriteString(fmt.Sprintf("%s [label=<\n", nombreInodo))
		dot.WriteString("<table border='1' cellborder='1' cellspacing='0'>\n")
		dot.WriteString(fmt.Sprintf("<tr><td colspan='2'><b>Inodo %d</b></td></tr>\n", i+1))
		dot.WriteString(fmt.Sprintf("<tr><td><b>i_uid</b></td><td>%d</td></tr>\n", inodo.I_uid))
		dot.WriteString(fmt.Sprintf("<tr><td><b>i_gid</b></td><td>%d</td></tr>\n", inodo.I_gid))
		dot.WriteString(fmt.Sprintf("<tr><td><b>i_size</b></td><td>%d</td></tr>\n", inodo.I_size))
		dot.WriteString(fmt.Sprintf("<tr><td><b>i_atime</b></td><td>%s</td></tr>\n", strings.TrimRight(string(inodo.I_atime[:]), "\x00")))
		dot.WriteString(fmt.Sprintf("<tr><td><b>i_ctime</b></td><td>%s</td></tr>\n", strings.TrimRight(string(inodo.I_ctime[:]), "\x00")))
		dot.WriteString(fmt.Sprintf("<tr><td><b>i_mtime</b></td><td>%s</td></tr>\n", strings.TrimRight(string(inodo.I_mtime[:]), "\x00")))

		for j, block := range inodo.I_block {
			if block != -1 {
				dot.WriteString(fmt.Sprintf("<tr><td><b>i_block_%d</b></td><td>%d</td></tr>\n", j+1, block))
			}
		}
		dot.WriteString(fmt.Sprintf("<tr><td><b>i_perm</b></td><td>%s</td></tr>\n", safeString(inodo.I_perm[:])))

		dot.WriteString("</table>>];\n")

		if prevInodo != "" {
			dot.WriteString(fmt.Sprintf("%s -> %s;\n", prevInodo, nombreInodo))
		}
		prevInodo = nombreInodo
	}

	dot.WriteString("}\n")

	// Crear carpeta si no existe
	dir := filepath.Dir(pathReporte)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return "",fmt.Errorf("Error al crear carpeta del reporte: %v", err)
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
	fmt.Printf("Total de inodos: %d\n", sb.S_inodes_count)
fmt.Printf("Total de bloques: %d\n", sb.S_blocks_count)
fmt.Printf("Inodos usados: %d\n", sb.S_inodes_count - sb.S_free_inodes_count)
fmt.Printf("Bloques usados: %d\n", sb.S_blocks_count - sb.S_free_blocks_count)


salida.WriteString(fmt.Sprintf("Reporte de Inodos generado exitosamente en:", pathReporte))
	return salida.String(), nil
}

func safeString(b []byte) string {
	s := strings.TrimRight(string(b), "\x00")
	if len(s) == 0 {
		return "-"
	}
	return s
}


