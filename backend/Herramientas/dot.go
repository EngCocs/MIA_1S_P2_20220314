package Herramientas

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// EscribirDotYGenerarImagen guarda un archivo .dot y genera la imagen final (.jpg, .png, etc.)
func EscribirDotYGenerarImagen(dotContent string, outputPath string) error {
	// Crear carpetas necesarias
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creando directorio: %v", err)
	}

	// Crear archivo temporal .dot
	tempDot := outputPath + ".dot"
	file, err := os.Create(tempDot)
	if err != nil {
		return fmt.Errorf("error creando archivo .dot: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(dotContent); err != nil {
		return fmt.Errorf("error escribiendo contenido .dot: %v", err)
	}

	// Determinar el tipo de imagen a generar (por extensión)
	ext := filepath.Ext(outputPath)
	format := ext[1:] // quitar el punto (.)
	cmd := exec.Command("dot", "-T"+format, tempDot, "-o", outputPath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error ejecutando dot: %v", err)
	}

	fmt.Println("Reporte generado exitosamente en:", outputPath)
	return nil
}
