package main

import (
)

const SERVO_NUM=3
var ServoId [SERVO_NUM]byte

func init(){
	ServoId[0]=1
	ServoId[1]=6
	ServoId[2]=10
}

type ServoPos struct{

	Pos[SERVO_NUM] int;
}

func ReadAllServo() (sp ServoPos, err error){

	for  i:=0;i<SERVO_NUM;i++ {
		sp.Pos[i], err=ReadPosition(ServoId[i])
		if err!=nil {
			return sp,err
		}
	}
	return sp,nil
}

