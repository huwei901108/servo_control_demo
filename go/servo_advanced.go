package main

import (
	"fmt"
	"github.com/pkg/term"
	"time"
)

const SERVO_NUM = 3

type ServoInfoS struct {
	Id           byte
	KeyUp        string
	KeyDn        string
	MoveEverSent bool
	MoveLastSent uint16
}

var ServoInfo [SERVO_NUM]ServoInfoS
var ServoId [SERVO_NUM]byte

const CtrlUIDelay = 1 * time.Millisecond
const CtrlUIKeyStop = " "
const CtrlUISpeedNums = 4

var CtrlUISpeed [CtrlUISpeedNums]uint16

const CtrlUISpeedUpAfter = 3

type CtrlUI struct {
	repeatType string
	repeatCnt  int
	speedLvl   int
}

const MoveRepeatSend = 0

func init() {
	ServoInfo[0] = ServoInfoS{1, "q", "a", false, 0}
	ServoInfo[1] = ServoInfoS{6, "e", "d", false, 0}
	ServoInfo[2] = ServoInfoS{10, "w", "s", false, 0}

	CtrlUISpeed[0] = 3
	CtrlUISpeed[1] = 6
	CtrlUISpeed[2] = 9

	for idx, info := range ServoInfo {
		ServoId[idx] = info.Id
	}

}

type ServoPos struct {
	Pos [SERVO_NUM]uint16
}

func ReadAllServo() (sp ServoPos, err error) {

	pr, err := NewPosReader(ServoId[:])
	if err != nil {
		return sp, err
	}

	pos, err := pr.ReadPosition()
	if err != nil {
		return sp, err
	}

	copy(sp.Pos[:], pos)
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

func (ui *CtrlUI) speedOfLvl() uint16 {
	if ui.speedLvl < 0 {
		ui.speedLvl = 0
	} else if ui.speedLvl >= CtrlUISpeedNums {
		ui.speedLvl = CtrlUISpeedNums - 1
	}
	return CtrlUISpeed[ui.speedLvl]
}
func (ui *CtrlUI) speedUp() {
	if ui.speedLvl < CtrlUISpeedNums-1 {
		ui.speedLvl++
	}
}

func (ui *CtrlUI) inputAndGetSpeed(input string) uint16 {
	if input == ui.repeatType {
		ui.repeatCnt++
	} else {
		ui.repeatCnt = 0
		ui.speedLvl = 0
	}
	if ui.repeatCnt >= CtrlUISpeedUpAfter {
		ui.repeatCnt = 0
		ui.speedUp()
	}
	ui.repeatType = input

	return ui.speedOfLvl()
}

func (ui *CtrlUI) Start() (ServoPos, error) {
	err := SerialOpen()
	target_sp, err := ReadAllServo()
	follow_target := true
	defer func() { follow_target = false }()

	go func() {
		for follow_target = true; follow_target; {
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

		var getStr string
		if numBytes == 0 {
			getStr = ""
		} else {
			getStr = string(bytes[0])
		}
		speed := ui.inputAndGetSpeed(getStr)

		fmt.Printf("\r \r %v ", target_sp)
		if numBytes == 0 {
			fmt.Printf("%*s", 20, "")
		} else {
			//fmt.Printf(" %s ", getStr)
			if getStr == CtrlUIKeyStop {
				break
			}
			foundKey := false
			for idx, info := range ServoInfo {
				if getStr == info.KeyUp {
					target_sp.Pos[idx] += speed
					foundKey = true
					break
				} else if getStr == info.KeyDn {
					target_sp.Pos[idx] -= speed
					foundKey = true
					break
				}
			}
			if !foundKey {
				fmt.Printf("%20s", "unkown input")
			} else {
				fmt.Printf("%*s", 20, "")
			}
		}
	}

	return target_sp, err
}

func MoveAllServoImmIfNeed(targetSp ServoPos) (err error) {
	for idx, info := range ServoInfo {
		thisTarget := targetSp.Pos[idx]
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
