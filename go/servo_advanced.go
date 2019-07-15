package main

import (
	"fmt"
	"github.com/pkg/term"
	"time"
)

const SERVO_NUM = 3

type ServoInfoS struct {
	Id		byte
	KeyUp		string
	KeyDn		string
	MoveEverSent	bool
	MoveLastSent	uint16
}

var ServoInfo [SERVO_NUM]ServoInfoS

const CtrlUIDelay = 1 * time.Millisecond
const CtrlUIKeyStop = " "
const CtrlUIInc = 3

const MoveRepeatSend = 0

func init() {
	ServoInfo[0] = ServoInfoS{1, "q", "a", false, 0}
	ServoInfo[1] = ServoInfoS{6, "e", "d", false, 0}
	ServoInfo[2] = ServoInfoS{10,"w", "s", false, 0}
}

type ServoPos struct {
	Pos [SERVO_NUM]uint16
}

func ReadAllServo() (sp ServoPos, err error) {

	for i := 0; i < SERVO_NUM; i++ {
		sp.Pos[i], err = ReadPosition(ServoInfo[i].Id)
		if err != nil {
			return sp, err
		}
	}
	return sp, nil
}

func MotorCtlAndRead() (ServoPos, error) {
	SerialOpen()
	for _, info := range ServoInfo {
		ServoMotorMove(info.Id, true, 0)
	}

	defer func() {
		for _, info := range ServoInfo {
			ServoMotorMove(info.Id, false, 0)
		}
	}()

	return ReadAllServo()
}

func CtrlUI() (ServoPos, error) {
	err := SerialOpen()
	target_sp, err := ReadAllServo()
	follow_target := true

	go func(){
		for follow_target=true; follow_target; {
			time.Sleep(CtrlUIDelay)
			MoveAllServoImmIfNeed(target_sp)
		}
	}()

	// start tty handling
	t, _ := term.Open("/dev/tty")
	defer func() {
		t.Restore()
		t.Close()
		fmt.Println("")
	}()
	term.RawMode(t)
	t.SetReadTimeout(CtrlUIDelay)

	fmt.Printf("Type KeyUp or KeyDn to move target. Type '%s' to confirm\n", CtrlUIKeyStop)
	fmt.Printf("%#v \n", ServoInfo)
	for {
		bytes := make([]byte, 1)
		numBytes, _ := t.Read(bytes)
		fmt.Printf("\r \r %v ", target_sp)
		if numBytes != 0 {
			getStr := string(bytes[0])
			//fmt.Printf(" %s ", getStr)
			if getStr == CtrlUIKeyStop {
				break
			}
			foundKey:= false
			for idx, info := range ServoInfo {
				if getStr == info.KeyUp {
					target_sp.Pos[idx] += CtrlUIInc
					foundKey=true
					break
				} else if getStr == info.KeyDn {
					target_sp.Pos[idx] -= CtrlUIInc
					foundKey=true
					break
				}
			}
			if !foundKey {
				fmt.Printf("%20s","unkown input")
			} else {
				fmt.Printf("%*s",20,"")
			}
		} else {
			fmt.Printf("%*s",20,"")
		}
	}

	return target_sp,err
}

func MoveAllServoImmIfNeed(targetSp ServoPos) (err error) {
	for idx, info := range ServoInfo {
		thisTarget:=targetSp.Pos[idx]
		//fmt.Printf("/r idx%d,thisTar%d,info%#v",idx,thisTarget,info)
		if info.MoveEverSent && thisTarget == info.MoveLastSent {
			// skip move
		} else {
			err = ServoMove(info.Id, thisTarget, 0)
		}
		ServoInfo[idx].MoveLastSent = thisTarget
		ServoInfo[idx].MoveEverSent = true
	}
	return err
}
