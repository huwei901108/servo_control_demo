//project main.go
package main

import (
	"errors"
	"fmt"
	"github.com/tarm/goserial"
	"io"
	"sync"
	"time"
)

const MAXRDLEN = 8000
const MAX2BYTE = 256*256 - 1
const SerialName = "/dev/ttyUSB0"
const SerialBaud = 115200
const SerialReadTimeoutInMs = 3
const ReadTimeoutInSec = 5
const DebugSerial = true

const ServoCmdMax=36

const ServoCmdMoveWrite = 1
const ServoCmdMoveStop = 12
const ServoCmdPosRead = 28
const ServoCmdMotorModeWrite = 29
const ServoCmdLoadUnloadWrite = 31

var ServoCmdLen [ServoCmdMax+1]byte

func init(){
	ServoCmdLen[ServoCmdMoveWrite] = 7
	ServoCmdLen[ServoCmdMoveStop] = 3
	ServoCmdLen[ServoCmdPosRead] = 3
	ServoCmdLen[ServoCmdMotorModeWrite] = 7
	ServoCmdLen[ServoCmdLoadUnloadWrite] = 4
}

type serial_handle struct {
	iorwc io.ReadWriteCloser
	mtx   sync.Mutex
}

var serial_handler serial_handle

func lock() {
	serial_handler.mtx.Lock()
}
func unlock() {
	serial_handler.mtx.Unlock()
}

func SerialOpen() error {
	lock()
	defer unlock()

	cfg := &serial.Config{Name: SerialName, Baud: SerialBaud, ReadTimeout: SerialReadTimeoutInMs}

	var err error
	serial_handler.iorwc, err = serial.OpenPort(cfg)
	if err != nil {
		return err
	}
	return nil
}
func SerialClose() {
	lock()
	defer unlock()

	if serial_handler.iorwc != nil {
		serial_handler.iorwc.Close()
	}
}

// pass any val not in range [0,MAX2BYTE] to pos or duration will make the proc ignore the parameter
func ServoWriteCmd(id byte, cmd byte, pos uint16, duration uint16) error {
	if DebugSerial {
		fmt.Printf("DebugSerial:prepare to WriteCmd(%v,%v,%v,%v)\n", id, cmd, pos, duration)
	}

	//open serial if needed
	if serial_handler.iorwc == nil {
		SerialOpen()
	}

	buf := []byte{0x55, 0x55}

	buf = append(buf, (id & 0xff), 0, (cmd & 0xff)) //len is set to 0 for now, will be set in later

	if cmd > ServoCmdMax {
		cmd = 0
	}

	if ServoCmdLen[cmd] >= 5 {
		buf = append(buf, byte(pos&0xff), byte((pos>>8)&0xff))
	}
	if ServoCmdLen[cmd] >= 7 {
		buf = append(buf, byte(duration&0xff), byte((duration>>8)&0xff))
	}

	buf[3] = byte((len(buf) - 2) & 0xff) // remove first 2 0x55

	// calc checksum
	var sum uint32
	for _, eachByte := range buf {
		sum += uint32(eachByte)
	}
	sum = sum - uint32(0x55) - uint32(0x55)
	sum = ^sum
	buf = append(buf, byte(sum&0xff))

	lock()
	defer unlock()
	if DebugSerial {
		fmt.Printf("DebugSerial:write_buf:%#v \n", buf)
	}
	_, err := serial_handler.iorwc.Write(buf)
	return err
}

func ReadPosition(id byte) (pos int, err error) {
	buffer := make([]byte, MAXRDLEN)

	lock()
	//clear buffer before send
	num, err := serial_handler.iorwc.Read(buffer)
	if err != nil && err != io.EOF {
		return 0, err
	}
	unlock()

	err = ServoWriteCmd(id, ServoCmdPosRead, 0, 0)
	if err != nil {
		return 0, err
	}

	p := posParser{}
	timeout := DelayTimer{}
	timeout.Start(ReadTimeoutInSec * time.Second)
	lock()
	defer unlock()
	for {
		num, err = serial_handler.iorwc.Read(buffer)
		if num > 0 {
			if DebugSerial {
				fmt.Printf("ReadBuf:%#v\n", buffer[:num])
			}

			pos, err := p.parse(buffer[:num])
			if err != nil {
				if DebugSerial {
					fmt.Printf("parse warning:%s", err.Error())
				}
			} else {
				return pos, nil
			}
		}

		if timeout.Timeout {
			return 0, errors.New(fmt.Sprintf("Read Timeout in sec:%d", ReadTimeoutInSec))
		}
	}

}

type posParser struct {
	state int
	pos   int
}

func (p *posParser) parse(buffer []byte) (pos int, err error) {

	for _, eachByte := range buffer {
		if DebugSerial {
			fmt.Printf("eachByte %v \n", eachByte)
		}
		switch p.state {
		case 0, 1:
			if eachByte == 0x55 {
				p.state++
			} else {
				err = errors.New(fmt.Sprintf("protocol err. state(%d)", p.state))
				p.state = 0
			}
		case 2, 3:
			p.state++
		case 4:
			if eachByte == 0x1C {
				p.state++
			} else {
				err = errors.New(fmt.Sprintf("protocal err. state(%d)", p.state))
				p.state = 0
			}
		case 5:
			p.pos = int(eachByte & 0xff)
			p.state++
		case 6:
			p.pos += int(eachByte&0xff) * 256
			p.state++
			return p.pos, err
		}
	}
	if err == nil {
		err = errors.New(fmt.Sprintf("frame not finished. state(%d)", p.state))
	}
	return 0, err

}

type DelayTimer struct {
	Timeout bool
}

func (dt *DelayTimer) Start(delay time.Duration) {
	dt.Timeout = false
	func() {
		time.Sleep(delay)
		dt.Timeout = true
	}()
}

func ServoMove(id byte, pos uint16, duration uint16) error {
	return ServoWriteCmd(id, ServoCmdMoveWrite, pos, duration)
}

func ServoMotorMove(id byte, enable bool, speed int16) error {
	var enable_u16 uint16
	if enable {
		enable_u16 = 1
	}else{
		enable_u16 = 0
	}
	return ServoWriteCmd(id, ServoCmdMotorModeWrite, enable_u16, uint16(speed))
}

