package AdministacionUserAndGrups

import (
     "fmt"
     "backend/session"
        "strings"
)

// sessionActive y currentUser son variables globales compartidas
// Si aún no se han declarado en este paquete, declara sólo una vez en un archivo y úsalas en todos los comandos.


// Logout cierra la sesión activa por lo tanto no se le pone parametros
func Logout(entrada []string) (string, error) {
    var salida strings.Builder
    salida.WriteString("========LOGOUT========")
    // Verifica que no se envíen parámetros, ya que no son necesarios
    if len(entrada) > 1 {
        salida.WriteString(fmt.Sprintf("LOGOUT Error: Este comando no recibe parámetros."))
        return salida.String(), nil
    }
    
    // Verificar si hay una sesión activa
    if !session.Active {
        salida.WriteString(fmt.Sprintf("LOGOUT Error: No hay una sesión activa. Inicia sesión para poder cerrar la sesión."))
        return salida.String(), nil
    }
    
    // Cerrar sesión
    salida.WriteString(fmt.Sprintf("Sesión cerrada correctamente para el usuario", session.CurrentUser))
    session.Active = false
    session.CurrentUser = ""
    return salida.String(), nil
}


