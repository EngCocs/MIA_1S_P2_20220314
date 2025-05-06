package Analizador

import (
	commands "backend/commands"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Analyzer procesa una lista de comandos y devuelve resultados y errores
func Analyzer(entradas []string) ([]string, []string) {
	var results []string
	var errors []string

	for i, entrada := range entradas {
		// Limpia la entrada y evita procesar líneas vacías o comentarios
		cleanedentrada := strings.TrimSpace(entrada) // Elimina espacios en blanco
		if cleanedentrada == "" || strings.HasPrefix(cleanedentrada, "#") {
			results = append(results, cleanedentrada) // Mostrar los comentarios
			continue
		}

		// Tokeniza la entrada
		tokens := strings.Fields(cleanedentrada) // Divide la cadena en palabras
		if len(tokens) == 0 {
			errors = append(errors, fmt.Sprintf("Comando %d: No se proporcionó ningún comando", i))
			continue
		}

		// Convertimos el comando a minúsculas para evitar problemas con mayúsculas
		command := strings.ToLower(tokens[0])
		args := tokens[1:] // Argumentos del comando

		// Manejamos los comandos
		var msg string
		var err error

		switch command {
		case "mkdisk":
			msg, err = commands.Mkdisk(args)
		
		case "clear":
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				errors = append(errors, fmt.Sprintf("Comando %d: Error al limpiar la pantalla: %s", i, err))
			}
		default:
			err = fmt.Errorf("comando desconocido: %s", command)
		}

		// Manejo de errores
		if err != nil {
			errors = append(errors, err.Error())
		} else {
			results = append(results, msg)
		}
	}

	return results, errors
}
