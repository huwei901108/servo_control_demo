package main

import (
	"testing"
)

func Test_read_servo_all(t *testing.T) {
	t.Log("test start")
	err := SerialOpen()
	if err != nil {
		t.Error(err.Error())
		return
	}

	sp, err := ReadAllServo()
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("read all servo", sp)


}
