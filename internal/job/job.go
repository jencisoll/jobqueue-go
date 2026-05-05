package  job

import "time"  

type Status string

const (
	StatusPending 	Status = "pending" //el job  esperando ser  procesado
	StatusProcessing Status = "processing"  //el job se esta  procesando ahora  mismo
	StatusDone 	Status = "done" 	//El job termino exitosamente
	StatusFailed 	Status = "failed "	//el job falló por algún error
)


type Job struct {


ID   string `json:"id"` //identificador único del job
Type string  `json:"type"` //que tipo de trabajo e s 
Status Status `json:"status"` //en que estado está
CreatedAt time.Time `json:"created_at"` //cuando fue creado
UpdatedAt time.Time `json:"updated_at"` //cuándo fue actualizado por última  vez

}



