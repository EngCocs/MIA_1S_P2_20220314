package reportes

import (
	"backend/Herramientas"
	"backend/Structs"
	"backend/permiso"
	"backend/session"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RepLS(pathDisco string, pathDir string, pathReporte string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========LS========\n")
	disco, err := os.Open(pathDisco)
	if err != nil {
		return "",fmt.Errorf("Error al abrir disco: %v", err)
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		return "",fmt.Errorf("Error al leer MBR: %v", err)
	}

	var particion Structs.Partition
	for _, p := range mbr.Partitions {
		if Structs.GetId(string(p.Id[:])) == session.PartitionID {
			particion = p
			break
		}
	}
	if particion.Size == 0 {
		return  "",fmt.Errorf("Partición no encontrada")
	}

	var super Structs.Superblock
	if err := Herramientas.ReadObject(disco, &super, int64(particion.Start)); err != nil {
		return "",fmt.Errorf("Error al leer superbloque: %v", err)
	}

	inodoID := permiso.SearchInode(0, pathDir, super, disco)
	if inodoID == -1 {
		return "",fmt.Errorf("Directorio no encontrado: %s", pathDir)
	}

	var carpeta Structs.Inode
	inodoPos := int64(super.S_inode_start) + int64(inodoID)*int64(binary.Size(Structs.Inode{}))
	if err := Herramientas.ReadObject(disco, &carpeta, inodoPos); err != nil {
		return "",fmt.Errorf("Error al leer inodo de carpeta")
	}

	report := "digraph ls {\n"
	report += "node [shape=plaintext];\n"
	report += "ls_table [label=<\n"
	report += "<table border='1' cellborder='1' cellspacing='0'>\n"
	report += "<tr><td bgcolor='#E0F7FA'><b>Permisos</b></td><td><b>Owner</b></td><td><b>Grupo</b></td><td><b>Size (en Bytes)</b></td><td><b>Fecha</b></td><td><b>Hora</b></td><td><b>Tipo</b></td><td><b>Name</b></td></tr>\n"

	for i := 0; i < 12; i++ {
		//fmt.Printf("  [%d] = %d\n", i, carpeta.I_block[i])
		if carpeta.I_block[i] == -1 {
			continue
		}

		blockPos := int64(super.S_block_start) + int64(carpeta.I_block[i])*int64(binary.Size(Structs.Folderblock{}))
		var folder Structs.Folderblock
		if err := Herramientas.ReadObject(disco, &folder, blockPos); err != nil {
			continue
		}

		for _, entry := range folder.B_content {
			name := Structs.GetB_name(string(entry.B_name[:]))
			if name == "-" || name == "." || name == ".." {
				continue
			}
			//fmt.Println("-> Entrada encontrada:", name)
			var objInode Structs.Inode
			inodoTargetPos := int64(super.S_inode_start) + int64(entry.B_inodo)*int64(binary.Size(Structs.Inode{}))
			if err := Herramientas.ReadObject(disco, &objInode, inodoTargetPos); err != nil {
				continue
			}

			perms := string(objInode.I_perm[:])
			if perms == "" {
				perms = "---"
			}
			owner := fmt.Sprintf("User%d", objInode.I_uid)
			group := fmt.Sprintf("Grupo%d", objInode.I_gid)
			size := objInode.I_size

			fecha := string(objInode.I_mtime[:])
			if len(fecha) > 10 {
				fecha = fecha[:10]
			}
			hora := string(objInode.I_mtime[:])
			if len(hora) >= 16 {
				hora = hora[11:16]
			}

			tipo := "Archivo"
			if objInode.I_type[0] == 0 {
				tipo = "Carpeta"
			}

			report += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n", perms, owner, group, size, fecha, hora, tipo, name)
		}
	}

	report += "</table>>];\n}\n"

	// Crear carpeta de destino si no existe
	dir := filepath.Dir(pathReporte)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return "", fmt.Errorf("error al crear carpeta destino: %v", err)
		}
	}

	
	dotFile := strings.ReplaceAll(pathReporte, ".jpg", ".dot")
	if err := os.WriteFile(dotFile, []byte(report), 0644); err != nil {
		return "",fmt.Errorf("error al escribir archivo dot: %v", err)
	}

	cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", pathReporte)
	if err := cmd.Run(); err != nil {
		return "",fmt.Errorf("error al generar imagen con dot: %v", err)
	}


	
	salida.WriteString(fmt.Sprintf("Reporte LS generado exitosamente en:", pathReporte))
	return salida.String(), nil
}

