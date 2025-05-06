package Structs

/*
Almacena la informacion de los Discos montados:
Se asigna una letra a cada disco montado y
va sumando 1 cada vez que se monta otra particion en dicho disco
*/
var Pmontaje []Mount //Lista de discos montados

type Mount struct {
	MPath  string //Path del Disco
	Letter byte   //Letra que se le asigna
	Cont   int    //COntador del numero de particion montada
}

// Para llevar el conteo de los discos
func AddPathM(path string, L byte, cont int) {
	Pmontaje = append(Pmontaje, Mount{MPath: path, Letter: L, Cont: cont})// aqui se agrega el disco montado
}

// ==============================================================================

// Almacena la informacion de cada Id junto a su Path
var Montadas []mountAlready//Lista de particiones montadas

type mountAlready struct {
	Id    string //Id de la particion
	PathM string //Path del disco al que pertenece la particion
}

// Ingresar particion montada
func AddMontadas(id string, path string) {
	Montadas = append(Montadas, mountAlready{Id: id, PathM: path})
}