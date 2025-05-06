package Structs

type Login struct {
    User  string
    Pass string
    Id string
}

type Group struct {
    
    Name  string
}


type CreateUser struct {
    User string
    Pass string
    Grp string//Grupo
}

type ChangeGRP struct {
    User string
    Grp string
}




