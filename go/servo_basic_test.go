package main

import (
	"testing"
	"time"
)

func Test_write_basic(t *testing.T) {
	t.Log("test start")
	err := SerialOpen()
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = ServoWriteCmd(1, 1, 300, 1000)
	if err != nil {
		t.Error(err.Error())
		return
	}
	time.Sleep(time.Duration(2 * time.Second))

	err = ServoWriteCmd(1, 1, 700, 1000)
	if err != nil {
		t.Error(err.Error())
		return
	}
	time.Sleep(time.Duration(2 * time.Second))
}
func Test_read_pos(t *testing.T) {
	t.Log("test start")
	err := SerialOpen()
	if err != nil {
		t.Error(err.Error())
		return
	}

	pos,err := ReadPosition(1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("read pos:", pos)
}
