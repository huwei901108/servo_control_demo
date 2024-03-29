#!/usr/bin/python3

import serial
##import pigpio
import time

## init pigpio lib
##pi=pigpio.pi()

##init serial
serialHandle = serial.Serial("/dev/ttyUSB0", 115200)


##
##
##
def servoWriteCmd(id, cmd, pos = None, duration = None):
    buf = bytearray(b'\x55\x55')
    try:
        len = 3   #default 3
        buf1 = bytearray(b'')

	##handle para
        if pos is not None:
            len += 2  #data len +2
            buf1.extend([(0xff & pos), (0xff & (pos >> 8))])  #high 8 and low 8, put to buffer
        if duration is not None:
            len += 2
            buf1.extend([(0xff & duration), (0xff & (duration >> 8))])  #to buffer
        buf.extend([(0xff & id), (0xff & len), (0xff & cmd)])
        buf.extend(buf1) #append para

	##calc checksum
        sum = 0x00
        for b in buf:  #sum up
            sum += b
        sum = sum - 0x55 - 0x55  #remove starting 0x55
        sum = ~sum  #revert
        buf.append(0xff & sum)  #get lower8 and append to buffer
        serialHandle.write(buf) #send
    except Exception as e:
        print(e)




##loop

while True:
    try:
        servoWriteCmd(1,1,0,1000) #send (id=1, cmd=1, position=0, time=1000ms)
        time.sleep(3)
        servoWriteCmd(1,1,1000,1000)
        time.sleep(3)
    except Exception as e:
        print(e)
        break
