package main

import (
	"testing"
	"time"
)

func Test_motor(t *testing.T) {
	t.Log("test start")
	var err error
	SerialOpen()
	ServoMove(1, 300, 1000)
	time.Sleep(time.Duration(3 * time.Second))
	defer ServoMove(1, 500, 1000)

	err = ServoMotorMove(1, true, 0)
	defer ServoMotorMove(1, false, 0)

	if err != nil {
		t.Error(err.Error())
		return
	}
	time.Sleep(time.Duration(2 * time.Second))

	err = ServoMotorMove(1, true, 200)
	if err != nil {
		t.Error(err.Error())
		return
	}
	time.Sleep(time.Duration(2 * time.Second))

	err = ServoMotorMove(1, true, 0)
	err = ServoMotorMove(1, true, -200)
	if err != nil {
		t.Error(err.Error())
		return
	}
	time.Sleep(time.Duration(2 * time.Second))

	err = ServoMotorMove(1, false, 0)
	if err != nil {
		t.Error(err.Error())
		return
	}
	time.Sleep(time.Duration(2 * time.Second))
}

func Test_move(t *testing.T) {
	t.Log("test start")
	var err error
	//err := SerialOpen()
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = ServoMove(1, 300, 1000)
	if err != nil {
		t.Error(err.Error())
		return
	}
	time.Sleep(time.Duration(2 * time.Second))

	err = ServoMove(1, 700, 1000)
	if err != nil {
		t.Error(err.Error())
		return
	}
	time.Sleep(time.Duration(2 * time.Second))
}
func Test_read_pos(t *testing.T) {
	t.Log("test start")
	var err error
	//err := SerialOpen()
	if err != nil {
		t.Error(err.Error())
		return
	}
	for i := 0; i < 2; i++ {
		pos, err := ReadPosition(1)
		if err != nil {
			t.Error(err.Error())
			return
		}
		t.Log("read pos:", pos)

		pos, err = ReadPosition(255)
		t.Log("err", err.Error())
	}
}
