//project main.go
package main

import (
	"errors"
	"fmt"
	"github.com/tarm/goserial"
	"io"
	"sync"
	"time"
	"os"
)

const MAXRDLEN = 8000
const MAX2BYTE = 256*256 - 1
const DefaultReadDelay = 3 * time.Millisecond
const SerialBaud = 115200
const SerialReadTimeoutInMs = 3
const ReadTimeoutInSec = 5
const DebugSerial = true

const ServoCmdMax = 36

const ServoCmdMoveWrite = 1
const ServoCmdMoveStop = 12
const ServoCmdPosRead = 28
const ServoCmdMotorModeWrite = 29
const ServoCmdLoadUnloadWrite = 31

var ServoCmdLen [ServoCmdMax + 1]byte
var SerialPaths []string

func init() {
	ServoCmdLen[ServoCmdMoveWrite] = 7
	ServoCmdLen[ServoCmdMoveStop] = 3
	ServoCmdLen[ServoCmdPosRead] = 3
	ServoCmdLen[ServoCmdMotorModeWrite] = 7
	ServoCmdLen[ServoCmdLoadUnloadWrite] = 4

	SerialPaths = []string{}
	SerialPaths = append(SerialPaths, "/dev/ttyUSB0" )
	SerialPaths = append(SerialPaths, "/dev/ttyUSB1" )
	SerialPaths = append(SerialPaths, "/dev/ttyUSB2" )
	SerialPaths = append(SerialPaths, "/dev/ttyUSB3" )
	SerialPaths = append(SerialPaths, "/dev/ttyUSB4" )
	SerialPaths = append(SerialPaths, "/dev/ttyUSB5" )
	SerialPaths = append(SerialPaths, "/dev/ttyUSB6" )
	SerialPaths = append(SerialPaths, "/dev/ttyUSB7" )
	SerialPaths = append(SerialPaths, "/dev/ttyUSB8" )
	SerialPaths = append(SerialPaths, "/dev/tty.SLAB_USBtoUART" )

	FNF=FrameNotFinError{}
	WFR=WaitingForResponseError{}
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

func findValidSerialPath() (path string, err error) {
	for _,path = range SerialPaths{
		_, err = os.Stat(path)
		if err == nil {
			return path, nil;
		}
	}
	return "", err
}

func SerialOpen() error {
	serialPath, err := findValidSerialPath()
	if err != nil {
		return err
	}

	if DebugSerial {
		fmt.Printf("using serial path: %s \n", serialPath)
	}

	lock()
	defer unlock()
	cfg := &serial.Config{Name: serialPath, Baud: SerialBaud, ReadTimeout: SerialReadTimeoutInMs}

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

func ServoWriteCmd(id byte, cmd byte, pos uint16, duration uint16) error {
	if DebugSerial {
		fmt.Printf("DebugSerial:prepare to WriteCmd(%v,%v,%v,%v)\n", id, cmd, pos, duration)
	}

	//open serial if needed
	if serial_handler.iorwc == nil {
		SerialOpen()
	}

	buf := []byte{0x55, 0x55}

	buf = append(buf, id&0xff, 0, cmd&0xff) //len is set to 0 for now, will be set in later

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

	buf = append(buf, calcCheckSum(buf))

	lock()
	defer unlock()
	if DebugSerial {
		fmt.Printf("DebugSerial:write_buf:%#v \n", buf)
	}
	_, err := serial_handler.iorwc.Write(buf)
	return err
}

func calcCheckSum(buf []byte) byte {
	var sum uint32
	for _, eachByte := range buf {
		sum += uint32(eachByte)
	}
	sum = sum - uint32(0x55) - uint32(0x55)
	sum = ^sum
	return byte(sum & 0xff)
}

type PosReader struct {
	Id        []byte
	ReqDelay  time.Duration
	BufferLen int
	buffer    []byte
	parser    posParser
	timeout   DelayTimer
}

func NewPosReader(inputId []byte) (pr *PosReader, err error) {
	pr = &PosReader{Id: inputId}
	err = pr.init()
	if err!= nil {
		return nil, err
	}
	return pr, nil
}

func (pr *PosReader)init() (err error) {
	if len(pr.Id) == 0 {
		return errors.New("empty id slice")
	}
	if pr.ReqDelay == 0 {
		pr.ReqDelay = DefaultReadDelay
	}
	if pr.BufferLen == 0 {
		pr.BufferLen = MAXRDLEN
	}
	pr.buffer = make([]byte, pr.BufferLen)
	pr.timeout = DelayTimer{}

	pr.parser = posParser{Id: pr.Id}
	err = pr.parser.init()
	if err != nil {
		return  err
	}

	//open serial if needed
	if serial_handler.iorwc == nil {
		SerialOpen()
	}


	lock()
	defer unlock()
	//clear buffer
	_, err = serial_handler.iorwc.Read(pr.buffer)
	if err != nil && err != io.EOF {
		return  err
	}
	return nil
}

func (pr *PosReader) ReadPosition() (pos []uint16, err error) {

	err = pr.init()
	if err != nil {
		return nil, err
	}

	for _, id := range pr.Id {
		err = ServoWriteCmd(id, ServoCmdPosRead, 0, 0)
		if err != nil {
			return nil, err
		}
		time.Sleep(pr.ReqDelay)
	}

	pr.timeout.Start(ReadTimeoutInSec * time.Second)
	for {
		lock()
		num, _ := serial_handler.iorwc.Read(pr.buffer)
		defer unlock()

		if num > 0 {
			if DebugSerial {
				fmt.Printf("ReadBuf:%#v\n", pr.buffer[:num])
			}

			pos, err := pr.parser.parse(pr.buffer[:num])
			if err != nil {
				if DebugSerial {
					fmt.Printf("parse warning:%s", err.Error())
				}
			} else {
				return pos, nil
			}
		}

		if pr.timeout.Timeout {
			return nil, errors.New(fmt.Sprintf("Read Timeout in sec:%d", ReadTimeoutInSec))
		}
	}

}

type WaitingForResponseError struct{}

func (w *WaitingForResponseError) Error() string {
	return "still waiting for some id to response"
}

var WFR WaitingForResponseError

type posParser struct {
	Id  []byte
	pos []uint16
	fin []bool
	sm  parseSM
}

func (p *posParser) init() error {
	if len(p.Id) == 0 {
		return errors.New("id slice is empty")
	}
	if p.anyDupId() {
		return errors.New("duplicate items in input id")
	}
	p.pos = make([]uint16, len(p.Id))
	p.fin = make([]bool, len(p.Id))
	for i := range p.fin {
		p.fin[i] = false
	}

	p.sm.init()

	return nil
}

func (p *posParser) anyDupId() bool {
	for i, ci := range p.Id {
		if i+1 >= len(p.Id ) {
			continue
		}
		for _, cj := range p.Id[i+1:] {
			if ci == cj {
				return true
			}
		}
	}
	return false
}

func (p *posParser) parse(recvBuffer []byte) (pos []uint16, err error) {
	err = &WFR

	for _, eachByte := range recvBuffer {
		if DebugSerial {
			fmt.Printf("eachByte %v \n", eachByte)
		}
		tid, tpos, terr := p.sm.parse(eachByte)
		if terr == nil {
			p.updatePosForId(tid, tpos)
			if p.allIdFin() {
				return p.pos, nil
			}
		} else {
			if err == nil || err == &FNF || err == &WFR {
				err = terr
			}
		}
	}

	return nil, err
}

func (p *posParser) updatePosForId(tid byte, tpos uint16) error {
	for whichId, mid := range p.Id {
		if tid == mid {
			p.pos[whichId] = tpos
			p.fin[whichId] = true
			return nil
		}
	}
	return errors.New(fmt.Sprintf("recv id(%d) not found in array", tid))
}
func (p *posParser) allIdFin() bool {
	for _, isFin := range p.fin {
		if !isFin {
			return false
		}
	}
	return true

}

type parseSM struct {
	state       int
	parseBuffer []byte
	parseId     byte
	parsePos    uint16
}

type FrameNotFinError struct {
	State int
}

var FNF FrameNotFinError

func (e *FrameNotFinError) Error() string {
	return fmt.Sprintf("frame not finished. state(%d)", e.State)
}

func (sm *parseSM) init(){
	sm.parseBuffer = make([]byte, 16)
}

func (sm *parseSM) parse(eachByte byte) (parseId byte, parsePos uint16, err error) {

	switch sm.state {
	case 0, 1:
		if eachByte == 0x55 {
			sm.state++
		} else {
			err = errors.New(fmt.Sprintf("protocol err. wrong start. state(%d)", sm.state))
			sm.state = 0
		}
	case 2:
		sm.parseId = eachByte
		sm.state++
	case 3:
		//length
		sm.state++
	case 4:
		if eachByte == 0x1C {
			sm.state++
		} else {
			err = errors.New(fmt.Sprintf("protocal err, wrong cmd byte. state(%d)", sm.state))
			sm.state = 0
		}
	case 5:
		sm.parsePos = uint16(eachByte & 0xff)
		sm.state++
	case 6:
		sm.parsePos += uint16(eachByte&0xff) * 256
		sm.state++
	case 7:
		calcSum := calcCheckSum(sm.parseBuffer)
		if eachByte == calcSum {
			sm.state = 0
			return sm.parseId, sm.parsePos, nil
		} else {
			err = errors.New(fmt.Sprintf("protocal err, wrong checksum. state(%d)", sm.state))
			sm.state = 0
		}

	default:
		err = errors.New(fmt.Sprintf("protocal err, wrong state. state(%d)", sm.state))
		sm.state = 0
	}

	if err == nil {
		sm.parseBuffer[sm.state] = eachByte
		FNF.State = sm.state
		return 0, 0, &FNF
	} else {
		return 0, 0, err
	}
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
	} else {
		enable_u16 = 0
	}
	return ServoWriteCmd(id, ServoCmdMotorModeWrite, enable_u16, uint16(speed))
}
