package main

import (
	"testing"
)

func Test_ctrl_UI(t *testing.T) {

	sp, err := CtrlUI()
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("read sp ", sp)

}
func Test_motor_ctl_and_read(t *testing.T) {
	t.Log("test start")
	sp, err := MotorCtlAndRead()
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("read all servo:", sp)
}

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
	t.Log("read all servo:", sp)
}
