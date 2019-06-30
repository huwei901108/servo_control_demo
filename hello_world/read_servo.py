#!/usr/bin/python3

import serial
import time
import struct

serialHandle = serial.Serial("/dev/ttyUSB0", 115200)  #bo te rate 115200

command = {"MOVE_WRITE":1, "POS_READ":28, "LOAD_UNLOAD_WRITE": 31}

def servoWriteCmd(id, cmd, par1 = None, par2 = None):
    buf = bytearray(b'\x55\x55')
    try:
        len = 3   #empty cmd len is 3
        buf1 = bytearray(b'')
        
        ## handle para
        if par1 is not None:
            len += 2  #data len +2
            par1 = 0xffff & par1
            buf1.extend([(0xff & par1), (0xff & (par1 >> 8))])  #low 8 high 8
        if par2 is not None:
            len += 2
            par2 = 0xffff & par2
            buf1.extend([(0xff & par2), (0xff & (par2 >> 8))])  #low 8 high 8
    
        buf.extend([(0xff & id), (0xff & len), (0xff & cmd)]) #append id len cmd
        buf.extend(buf1) # append para
        
        ##calc checksum
        sum = 0x00
        for b in buf:
            sum += b
        sum = sum - 0x55 - 0x55  #remove starting 0x55
        sum = ~sum
        buf.append(0xff & sum)  #append low 8
        
        serialHandle.write(buf) #send
        
    except Exception as e:
        print(e)


def readPosition(id):
    serialHandle.flushInput()
    servoWriteCmd(id, command["POS_READ"])
    time.sleep(0.00034)  #wait until cmd is sent
    time.sleep(0.005)  #wait until data is get
    time.sleep(1)
    count = serialHandle.inWaiting() #get serial buffer len
    pos = None
    print("count=%d" %count)
    if count != 0:
        recv_str = serialHandle.read(count)
        recv_data = []
        for tmpStr in recv_str:
            tmpInt, = struct.unpack('B', tmpStr)
            recv_data.append(tmpInt)
        if count == 8: #correct num for read pos
            if recv_data[0] == 0x55 and recv_data[1] == 0x55 and recv_data[4] == 0x1C:
                 #correct format
                 print 'correct format ', recv_data[5], ' and ', recv_data[6]
                 pos= 0xffff & (recv_data[5] | (0xff00 & (recv_data[6] << 8))) #sum up pos

    return pos

servoWriteCmd(1, command["LOAD_UNLOAD_WRITE"],0)  #make Moto power off, can be moved by hand
while True:
    try:
        pos = readPosition(10) #read(id)
        print(pos)
        time.sleep(1)
        
    except Exception as e:
        print(e)
        break
