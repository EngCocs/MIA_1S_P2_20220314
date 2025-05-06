package commands

import (
	"backend/Herramientas"
	
	"backend/Structs"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// mkdisk -Size=3000 -unit=K -path=/home/user/Disco1.mia​
func Mkdisk(parametros []string)(string, error) {
	var salida strings.Builder
	salida.WriteString(fmt.Sprintf("========MKDISK========"))
	//valida entrada de parametros del comando leido
	//PARAMETROS: -size -unit -fit -path
	var size int      //obligatorio
	var path string   //obligatorio
	fit := "F"        //por defecto es ff por eso se inicializa con ese valor (valores para fit: f, w, b pero de entrada se recibe FF, WF o BF)
	unit := 1048576   //1024*1024 Por defecto es M por eso se inicializa con este valor en bytes
	paramC := true    //valida que todos los parametros sean correctos
	sizeInit := false //Para saber si entro el parametro size (obligatorio) false -> no inicializado (esto es por si no viniera en los parametros)
	pathInit := false //para asegurar que el parametro path si existe (obligatorio)

	//_ sería el indice pero se omite y con [1:] indicamos que inicie el indice 1 en lugar del 0
	//esto porque en [0] esta el comando mkdisk que estamos ejecutando
	//recorro parametros del mkdisk asignando sus valores segun sea el caso
	for _, parametro := range parametros[1:] {
		//quito los espacios en blano despues de cada parametro
		tmp2 := strings.TrimRight(parametro, " ")
		//divido cada parametro entre nombre del parametro y su valor # -size=25 -> -size, 25
		tmp := strings.Split(tmp2, "=") 

		//Si falta el valor del parametro actual lo reconoce como error e interrumpe el proceso
		if len(tmp) != 2 {// si no tiene 2 elementos es porque no tiene valor
			salida.WriteString(fmt.Sprintf("MKDISK Error: Valor desconocido del parametro ", tmp[0]))
			paramC = false
			break //para finalizar el ciclo for con el error y no ejecutar lo que haga falta
		}

		//en tmp valido que parametro viene en su primera posicion y que tenga un valor
		//SIZE
		if strings.ToLower(tmp[0]) == "size" {
			sizeInit = true//bandera para saber que si se ingreso el parametro size
			var err error//variable para manejar errores
			size, err = strconv.Atoi(tmp[1]) //se convierte el valor en un entero
			//if err != nil || size <= 0 { //Se manejaria como un solo error
			if err != nil {
				salida.WriteString(fmt.Sprintf("MKDISK Error: -size debe ser un valor numerico. se leyo ", tmp[1]))
				paramC = false
				break
			} else if size <= 0 { //se valida que sea mayor a 0 (positivo)
				salida.WriteString(fmt.Sprintf("MKDISK Error: -size debe ser un valor positivo mayor a cero (0). se leyo ", tmp[1]))
				paramC = false
				break
			}
			//FIT
		} else if strings.ToLower(tmp[0]) == "fit" {
			//Si el ajuste es BF (best fit)
			if strings.ToLower(tmp[1]) == "bf" {
				//asigno el valor del parametro en su respectiva variable
				fit = "B"
				//Si el ajuste es WF (worst fit)
			} else if strings.ToLower(tmp[1]) == "wf" {
				//asigno el valor del parametro en su respectiva variable
				fit = "W"
				//Si el ajuste es ff ya esta definido por lo que si es distinto es un error
			} else if strings.ToLower(tmp[1]) != "ff" {
				salida.WriteString(fmt.Sprintf("MKDISK Error en -fit. Valores aceptados: BF, FF o WF. ingreso: ", tmp[1]))
				paramC = false
				break
			}
			//UNIT
		} else if strings.ToLower(tmp[0]) == "unit" {
			//si la unidad es k
			if strings.ToLower(tmp[1]) == "k" {
				//asigno el valor del parametro en su respectiva variable
				unit = 1024
				//si la unidad no es k ni m es error (si fuera m toma el valor con el que se inicializo unit al inicio del metodo)
			} else if strings.ToLower(tmp[1]) != "m" {
				salida.WriteString(fmt.Sprintf("MKDISK Error en -unit. Valores aceptados: k, m. ingreso: ", tmp[1]))
				paramC = false
				break
			}
			//PATH
		} else if strings.ToLower(tmp[0]) == "path" {
			pathInit = true// bandera para saber que si se ingreso el parametro path
			path = tmp[1]
			//ERROR EN LOS PARAMETROS LEIDOS
		} else {
			salida.WriteString("MKDISK Error: Parametro desconocido: "+ tmp[0])
			paramC = false
			break //por si en el camino reconoce algo invalido de una vez se sale
		}
	}

	if paramC {
		//Verificar que si se haya inicializado el parametro size (es decir que si viniera el parametro)
		if sizeInit && pathInit {
			tam := size * unit // tamaño del disco en bytes
			//carpeta := "./MIA/P1/" //Ruta (carpeta donde se guardara el disco)

			//Nombre disco. Solo por control del disco que se esta creando
			nombreDisco := strings.Split(path, "/") // separar la ruta para obtener el nombre del disco ASI # /home/user/Disco1.mia
			disco := nombreDisco[len(nombreDisco)-1]// obtener el nombre del disco (Disco1.mia)

			// Create file
			err := Herramientas.CrearDisco(path)
			if err != nil {
				salida.WriteString(fmt.Sprintf("MKDISK Error:: ", err))
			}

			// Open bin file
			file, err := Herramientas.OpenFile(path)
			if err != nil {
				return "", err
			}

			// create array of byte(0)
			datos := make([]byte, tam)
			newErr := Herramientas.WriteObject(file, datos, 0)
			if newErr != nil {
				salida.WriteString(fmt.Sprintf("MKDISK Error: ", newErr))
				return "", newErr
			}

			//obtener hora para el id
			ahora := time.Now()
			//obtener los segundos y minutos
			segundos := ahora.Second()
			minutos := ahora.Minute()
			//concatenar los segundos y minutos como una cadena (de 4 digitos)
			cad := fmt.Sprintf("%02d%02d", segundos, minutos)
			//convertir la cadena a numero en un id temporal
			idTmp, err := strconv.Atoi(cad)
			if err != nil {
				salida.WriteString("MKDISK Error: no se convirtio fecha en entero para id")
			}
			//salida.WriteString(fmt.Sprintf("id guardado actual ", idTmp)
			// Create a new instance of MBR
			var newMBR Structs.MBR
			newMBR.MbrSize = int32(tam)
			newMBR.Id = int32(idTmp)
			copy(newMBR.Fit[:], fit)
			copy(newMBR.FechaC[:], ahora.Format("02/01/2006 15:04"))
			// Write object in bin file
			if err := Herramientas.WriteObject(file, newMBR, 0); err != nil {
				return "", err
			}

			// Close bin file
			defer file.Close()

			salida.WriteString("\n Se creo el disco "+ disco+ " de forma exitosa")

			//imprimir el disco creado para validar que todo este correcto
			var TempMBR Structs.MBR
			if err := Herramientas.ReadObject(file, &TempMBR, 0); err != nil {
				return "", err
			}
			Structs.PrintMBR(TempMBR)

			fmt.Printf("\n======End MKDISK======\n")
			return salida.String(), nil

		} else {
			salida.WriteString(fmt.Sprintf("MKDISK Error: Falta parametro -size"))
		}
	}
	return salida.String(), nil

}
