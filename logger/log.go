package logger

import "log"

func init(){

}

func Info(str string){
	log.Println(str)
}

func Warn(err error){
	log.Println(err.Error())
}

func Fatal(str string){
	log.Println(str)
}