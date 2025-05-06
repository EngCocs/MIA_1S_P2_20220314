package commands;

import (
	"backend/Herramientas"
	"backend/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

func Mkfs(entrada []string) (string, error){
	var salida strings.Builder
	salida.WriteString("========MKFS========")
	var id string //obligatorio
	paramC := true//parametros correctos
	var pathDico string//guarda la ruta del disco

	for _, parametro := range entrada[1:] {
		tmp := strings.TrimRight(parametro, " ")//quitar espacios en blanco
		valores := strings.Split(tmp, "=")//dividir el parametro en nombre y valor

		if len(valores) != 2 {
			salida.WriteString(fmt.Sprintf("ERROR MKFS, valor desconocido de parametros ", valores[1]))
			break
		}

		if strings.ToLower(valores[0]) == "id" {
			id = strings.ToUpper(valores[1])
		} else if strings.ToLower(valores[0]) == "type" {
			if strings.ToLower(valores[1]) != "full" {
				salida.WriteString(fmt.Sprintf("MKFS Error. Valor de -type desconocido"))
				paramC = false
				break
			}

			//ERROR EN LOS PARAMETROS LEIDO
		} else {
			salida.WriteString(fmt.Sprintf("MKFS Error: Parametro desconocido: ", valores[0]))
			paramC = false
			break //por si en el camino reconoce algo invalido de una vez se sale
		}

	}

	//obtener la particion correspondiente al id
	if id != "" {//si id no esta vacio 
		//BUsca en struck de particiones montadas el id ingresado
		for _, montado := range Structs.Montadas {
			if montado.Id == id {
				pathDico = montado.PathM//cargamos el path de esa particion
			}
		}
		if pathDico == "" {
			salida.WriteString(fmt.Sprintf("ERROR MKFS NO SE ENCONTRA EL ID"))
			paramC = false
		}
	} else {
		salida.WriteString(fmt.Sprintf("ERROR MKFS NO SE INGRESO ID"))
		paramC = false
	}

	if paramC {
		//Abrir el Disco de la particion
		discof, err := Herramientas.OpenFile(pathDico)
		if err != nil {
			return salida.String(), err 
		}

		//Cargar el mbr
		var mbr Structs.MBR
		// leer el mbr
		//ReadObject(file *os.File, data interface{}, position int64)
		if err := Herramientas.ReadObject(discof, &mbr, 0); err != nil {
			return salida.String(), err
		}

		// Close bin discof
		defer discof.Close()//cerrar el disco

		//Buscar particion con el id solicitado
		formatear := true//bandera para saber si se encontro la particion
		for i := 0; i < 4; i++ {
			identificador := Structs.GetId(string(mbr.Partitions[i].Id[:]))// aqui se obtiene el id de la particion
			if identificador == id {
				formatear = false //Si encontro la particion

				//Crear el super bloque que contiene los datos del sistema de archivos. Es similar al mbr en los discos
				var newSuperBloque Structs.Superblock
				//ReadObject(file *os.File, data interface{}, position int64)
				//Herramientas.ReadObject(discof, &newSuperBloque, int64(mbr.Partitions[i].Start))//leer el superbloque de la particion encontrada en el mbr

				//Calcular el numero de inodos que caben en la particion. El numero de bloques es el triple de inodos
				//(formula a partir del tamaño de la particion, esta en el enunciado pag. 10)
				//tamaños fisicos: SuperBloque = 92; Inodo = 124; Bloque = 64
				numerador := int(mbr.Partitions[i].Size) - binary.Size(Structs.Superblock{})//tamaño de la particion menos el tamaño del superbloque
				denominador := 4 + binary.Size(Structs.Inode{}) + 3*binary.Size(Structs.Fileblock{})//4 es el tamaño de un bloque

				n := int32(numerador / denominador) //numero de inodos

				//inicializar atributos generales del superbloque
				newSuperBloque.S_blocks_count = int32(3 * n)      //Total de bloques creados (pueden usarse)
				newSuperBloque.S_free_blocks_count = int32(3 * n) //Numero de bloques libre (Todos estan libres por ahora)

				newSuperBloque.S_inodes_count = n      //Total de inodos creados (pueden usarse)
				newSuperBloque.S_free_inodes_count = n //numero de inodos libres (todos estan libres por ahora)

				newSuperBloque.S_inode_size = int32(binary.Size(Structs.Inode{}))//tamaño de un inodo
				newSuperBloque.S_block_size = int32(binary.Size(Structs.Fileblock{}))//tamaño de un bloque

				//obtener hora de montaje del sistema de archivos
				ahora := time.Now()
				copy(newSuperBloque.S_mtime[:], ahora.Format("02/01/2006 15:04"))//fecha de montaje
				//Si fecha de desmontaje coincide con montaje es porque aun no se monta
				copy(newSuperBloque.S_umtime[:], ahora.Format("02/01/2006 15:04"))//fecha de desmontaje
				newSuperBloque.S_mnt_count += 1 //Se esta montando por primera vez
				newSuperBloque.S_magic = 0xEF53//valor que identifica el sistema de archivos

				crearEXT2(n, mbr.Partitions[i], newSuperBloque, ahora.Format("02/01/2006 15:04"), discof)//crear el sistema de archivos EXT2

				//Fin del formateo
				salida.WriteString(fmt.Sprintf("Particion con id ", id, " formateada correctamente"))

				//Si hubiera una sesion iniciada eliminarla
				break //para que ya no siga recorriendo las demas particiones
			}
		}

		if formatear {
			salida.WriteString(fmt.Sprintf("MKFS Error. No se pudo formatear la particion con id ", id))
			salida.WriteString(fmt.Sprintf("MKFS Error. No existe el id"))
		}
	}
	return salida.String(), nil
}

func crearEXT2(n int32, particion Structs.Partition, newSuperBloque Structs.Superblock, date string, discof *os.File) (string, error) {
	fmt.Println("Superbloque: ", newSuperBloque)
	fmt.Println("Fecha: ", date)
	var salida strings.Builder

	//completar los atributos del super bloque. La estructura de la particion formateada es:
	// | Superbloque | Bitmap Inodos | Bitmap Bloques | Inodos | Bloques | 

	//tipo del sistema de archivos
	newSuperBloque.S_filesystem_type = 2 //2 -> EXT2; 3 -> EXT3
	//Bitmap Inodos inicia donde termina el superbloque fisicamente (y el superbloque esta al inicio de la particion)
	newSuperBloque.S_bm_inode_start = particion.Start + int32(binary.Size(Structs.Superblock{}))
	//Bitmap bloques inicia donde termina el de inodos. Se suma n que es el numero de inodos maximo
	newSuperBloque.S_bm_block_start = newSuperBloque.S_bm_inode_start + n
	//Se crea el primer Inodo. Esta al final de los bloques que son 3 veces el numero de inodos
	newSuperBloque.S_inode_start = newSuperBloque.S_bm_block_start + 3*n
	//Se crea el primer bloque, este esta al final de los inodos fisicos
	newSuperBloque.S_block_start = newSuperBloque.S_inode_start + n*int32(binary.Size(Structs.Inode{}))

	//Se restan 2 bloques y dos inodos. uno para la carpeta raiz y otro para el archivo users.txt
	//lo que se crea al formatear es /users.txt (la carpeta usa un inodo y el archivo otro)
	newSuperBloque.S_free_inodes_count -= 2//se restan 2 inodos
	newSuperBloque.S_free_blocks_count -= 2//se restan 2 bloques

	//primer inodo libre
	//newSuperBloque.S_first_ino = newSuperBloque.S_inode_start + 2*int32(binary.Size(Structs.Inode{})) //multiplico por 2 porque hay 2 inodos creados
	newSuperBloque.S_first_ino = int32(2)//inodo 0 y 1 ocupados
	//primer bloque libre
	//newSuperBloque.S_first_blo = newSuperBloque.S_block_start + 2*int32(binary.Size(Structs.Fileblock{})) //multiplicar por 2 porque hay 2 bloques creados
	newSuperBloque.S_first_blo = int32(2)

	//limpio (formateo) el espacio del bitmap de inodos para evitar inconsistencias
	bmInodeData := make([]byte, n)//n es el numero de inodos y make crea un arreglo de bytes con n elementos
	bmInodeErr := Herramientas.WriteObject(discof, bmInodeData, int64(newSuperBloque.S_bm_inode_start))//escribir el bitmap de inodos en el disco
	if bmInodeErr != nil {
		salida.WriteString(fmt.Sprintf("MKFS Error: ", bmInodeErr))
		return salida.String(), bmInodeErr
	}

	//limpiar (formatear) el espacio del bitmap de bloques para evitar inconsistencias
	bmBlockData := make([]byte, 3*n)
	bmBlockErr := Herramientas.WriteObject(discof, bmBlockData, int64(newSuperBloque.S_bm_block_start))
	if bmBlockErr != nil {
		salida.WriteString(fmt.Sprintf("MKFS Error: ", bmInodeErr))
		return salida.String(), bmBlockErr
	}

	//creo un inodo y lleno el arreglo de bloques con -1
	var newInode Structs.Inode
	for i := 0; i < 15; i++ {
		newInode.I_block[i] = -1//el valor -1 indica que no se esta usando
	}

	//creo todos los inodos del sistema de archivos
	for i := int32(0); i < n; i++ {
		err := Herramientas.WriteObject(discof, newInode, int64(newSuperBloque.S_inode_start+i*int32(binary.Size(Structs.Inode{}))))//escribir el inodo en el disco
		if err != nil {
			salida.WriteString(fmt.Sprintf("MKFS Error: ", err))
			return salida.String(), err
		}
	}

	//Crear todos los bloques de carpeta que se pueden crear
	fileBlocks := make([]Structs.Fileblock, 3*n) //lo puedo trabajar asi porque son instancias de la estructura, el inode llevaban valores
	fileBlocksErr := Herramientas.WriteObject(discof, fileBlocks, int64(newSuperBloque.S_bm_block_start))
	if fileBlocksErr != nil {
		salida.WriteString(fmt.Sprintf("MKFS Error: ", fileBlocksErr))
		return salida.String(), fileBlocksErr
	}

	//Crear el Inode 0
	var Inode0 Structs.Inode
	Inode0.I_uid = 1
	Inode0.I_gid = 1
	Inode0.I_size = 0 //por ser carpeta no tiene tamaño como tal. para saber si existe basarse en I_ui/I_gid
	//unica vez que las 3 fechas son iguales
	copy(Inode0.I_atime[:], date)
	copy(Inode0.I_ctime[:], date)
	copy(Inode0.I_mtime[:], date)
	copy(Inode0.I_type[:], "0") //como es raiz es de tipo carpeta
	copy(Inode0.I_perm[:], "664")//permisos de la carpeta

	for i := int32(0); i < 15; i++ {
		Inode0.I_block[i] = -1
	}

	Inode0.I_block[0] = 0 //apunta al bloque 0

	//Crear el folder con la estructura
	// 	. 		| 0   -> actual (a si mismo)
	// 	..      | 0   -> el padre
	//users.txt | 1
	//			|-1

	var folderBlock0 Structs.Folderblock //Bloque0 -> carpetas
	folderBlock0.B_content[0].B_inodo = 0//apunta a si mismo
	copy(folderBlock0.B_content[0].B_name[:], ".")//nombre de la carpeta
	folderBlock0.B_content[1].B_inodo = 0//apunta al padre
	copy(folderBlock0.B_content[1].B_name[:], "..")//nombre de la carpeta
	folderBlock0.B_content[2].B_inodo = 1//apunta al inodo 1
	copy(folderBlock0.B_content[2].B_name[:], "users.txt")//nombre de la carpeta
	folderBlock0.B_content[3].B_inodo = -1//apunta a nada

	//Inode1 que es el que contiene el archivo (Bloque 0 apunta a este nuevo inodo)
	var Inode1 Structs.Inode
	Inode1.I_uid = 1
	Inode1.I_gid = 1
	Inode1.I_size = int32(binary.Size(Structs.Folderblock{}))
	copy(Inode1.I_atime[:], date)
	copy(Inode1.I_ctime[:], date)
	copy(Inode1.I_mtime[:], date)
	copy(Inode1.I_type[:], "1") //es del archivo
	copy(Inode0.I_perm[:], "664")
	for i := int32(0); i < 15; i++ {
		Inode1.I_block[i] = -1 //estos son los bloques que se usan por el archivo en caso de que sea muy grande (punteros) 
	}
	//Inode1 apunta al bloque1 (en este caso el bloque1 contiene el archivo)
	Inode1.I_block[0] = 1//apunta al bloque 1
	data := "1,G,root\n1,U,root,root,123\n"//datos del archivo users.txt
	var fileBlock1 Structs.Fileblock //Bloque1 -> archivo
	copy(fileBlock1.B_content[:], []byte(data))//escribir los datos en el bloque
	fmt.Println("Creado users.txt con los datos : \n", data)

	//resumen
	//Inodo 0 -> Bloque 0 -> Inodo1 -> bloque1 (archivo)

	//Crear la carpeta raiz /
	//crear el archivo users.txt

	//fmt.Println("Superbloque: ", newSuperBloque)
	fmt.Println("→ Valores calculados del Superbloque:")
fmt.Println("S_inodes_count:", newSuperBloque.S_inodes_count)
fmt.Println("S_blocks_count:", newSuperBloque.S_blocks_count)
fmt.Println("S_bm_inode_start:", newSuperBloque.S_bm_inode_start)
fmt.Println("S_bm_block_start:", newSuperBloque.S_bm_block_start)
fmt.Println("S_inode_start:", newSuperBloque.S_inode_start)
fmt.Println("S_block_start:", newSuperBloque.S_block_start)
fmt.Println("S_first_ino:", newSuperBloque.S_first_ino)
fmt.Println("S_first_blo:", newSuperBloque.S_first_blo)


	// Escribir el superbloque
	Herramientas.WriteObject(discof, newSuperBloque, int64(particion.Start))
	fmt.Println("→ Superbloque antes de cerrar:")
	fmt.Println("S_inodes_count:", newSuperBloque.S_inodes_count)
	fmt.Println("S_free_inodes_count:", newSuperBloque.S_free_inodes_count)
	fmt.Println("S_blocks_count:", newSuperBloque.S_blocks_count)
	fmt.Println("S_free_blocks_count:", newSuperBloque.S_free_blocks_count)
	
	//escribir el bitmap de inodos
	Herramientas.WriteObject(discof, byte(1), int64(newSuperBloque.S_bm_inode_start))
	Herramientas.WriteObject(discof, byte(1), int64(newSuperBloque.S_bm_inode_start+1)) //Se escribieron dos inode
	fmt.Println("→ Bitmap de inodos: marcados como usados los primeros 2")

	//escribir el bitmap de bloques (se usaron dos bloques)
	Herramientas.WriteObject(discof, byte(1), int64(newSuperBloque.S_bm_block_start))
	Herramientas.WriteObject(discof, byte(1), int64(newSuperBloque.S_bm_block_start+1))	
	//escribir inodes
	fmt.Println("→ Bitmap de bloques: marcados como usados los primeros 2")

	//Inode0
	Herramientas.WriteObject(discof, Inode0, int64(newSuperBloque.S_inode_start))
	//Inode1
	Herramientas.WriteObject(discof, Inode1, int64(newSuperBloque.S_inode_start+int32(binary.Size(Structs.Inode{}))))

	//Escribir bloques
	//bloque0
	Herramientas.WriteObject(discof, folderBlock0, int64(newSuperBloque.S_block_start))
	//bloque1
	Herramientas.WriteObject(discof, fileBlock1, int64(newSuperBloque.S_block_start+int32(binary.Size(Structs.Fileblock{}))))
	// Fin crear EXT2
	fmt.Println("Montadas actuales:", Structs.Montadas)
	
	return salida.String(), nil
}
