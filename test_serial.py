#!/usr/bin/python3

import serial
import time

if __name__ == "__main__":
    
    serialHandle = serial.Serial("/dev/ttyUSB0", 115200)
    while True:
        serialHandle.write(b"HELLO WORLD\r\n")
        time.sleep(1)


