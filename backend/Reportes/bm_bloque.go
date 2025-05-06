package reportes

import (
	"backend/Herramientas"
	"backend/Structs"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GenerarReporteBmBlock(pathDisco string, pathSalida string) (string, error) {
	var salida strings.Builder
	salida.WriteString("========BM_BLOCK========")
	// Abrir el disco
	disco, err := os.Open(pathDisco)
	if err != nil {
		return "",fmt.Errorf("Error al abrir disco: %v", err)
	}
	defer disco.Close()

	// Leer MBR
	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		return "",fmt.Errorf("Error al leer MBR: %v", err)
	}

	// Buscar partición primaria válida
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
		return "",fmt.Errorf("Error al leer superbloque: %v", err)
	}

	// Leer el bitmap de bloques
	bitmap := make([]byte, sb.S_blocks_count)
	if _, err := disco.ReadAt(bitmap, int64(sb.S_bm_block_start)); err != nil {
		return "",fmt.Errorf("Error al leer bitmap de bloques: %v", err)
	}

	// Crear archivo de salida
	file, err := os.Create(pathSalida)
	if err != nil {
		return "",fmt.Errorf("Error al crear archivo de salida: %v", err)
	}
	defer file.Close()

	// Escribir los bits en líneas de 20
	count := 0
	for _, bit := range bitmap {
		val := "0"
		if bit == 1 {
			val = "1"
		}
		if _, err := file.WriteString(val + " "); err != nil {
			return "", fmt.Errorf("Error al escribir bit: %v", err)
		}
		count++
		if count == 20 {
			count = 0
			if _, err := file.WriteString("\n"); err != nil {
				return "", fmt.Errorf("Error al escribir salto de línea: %v", err)
			}
		}
	}

	// Crear carpeta si no existe
	dir := filepath.Dir(pathSalida)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return "",fmt.Errorf("Error al crear carpeta del reporte: %v", err)
		}
	}

	salida.WriteString(fmt.Sprintf("Reporte bm_block generado exitosamente en:", pathSalida))
	return salida.String(), nil
}
