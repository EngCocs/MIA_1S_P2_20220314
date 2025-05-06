package reportes

import (
	"backend/Herramientas"
	"backend/Structs"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GenerarReporteBMInode(pathDisco string, pathReporte string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========BM_INODE========\n")
	// Abrimos el disco
	disco, err := os.Open(pathDisco)
	if err != nil {
		return "",fmt.Errorf("Error al abrir el disco: %v", err)
	}
	defer disco.Close()

	// Leer MBR
	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		return "",fmt.Errorf("Error al leer MBR: %v", err)
	}

	// Buscar la primera partición válida no extendida
	var particion Structs.Partition
	for _, p := range mbr.Partitions {
		if p.Size > 0 && string(p.Type[:]) != "E" {
			particion = p
			break
		}
	}
	if particion.Size == 0 {
		return "",fmt.Errorf("No se encontró una partición válida")
	}

	// Leer Superbloque
	var sb Structs.Superblock
	if err := Herramientas.ReadObject(disco, &sb, int64(particion.Start)); err != nil {
		return "",fmt.Errorf("Error al leer el superbloque: %v", err)
	}

	// Construir salida
	var builder strings.Builder
	for i := int32(0); i < sb.S_inodes_count; i++ {
		byteVal := make([]byte, 1)
		if _, err := disco.ReadAt(byteVal, int64(sb.S_bm_inode_start+i)); err != nil {
			break
		}
		if byteVal[0] == 1 {
			builder.WriteString("1")
		} else {
			builder.WriteString("0")
		}
		if (i+1)%20 == 0 {
			builder.WriteString("\n")
		} else {
			builder.WriteString(" ")
		}
	}

	// Crear carpeta si no existe
	dir := filepath.Dir(pathReporte)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return "",fmt.Errorf("Error al crear carpeta del reporte: %v", err)
		}
	}
	// Guardar archivo .txt
	if err := os.WriteFile(pathReporte, []byte(builder.String()), 0644); err != nil {
		return "",fmt.Errorf("Error al escribir el reporte bm_inode: %v", err)
	}

	salida.WriteString(fmt.Sprintf("Reporte bm_inode generado exitosamente en:", pathReporte))
	return salida.String(), nil
}
