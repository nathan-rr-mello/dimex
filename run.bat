@echo off

del mxOUT.txt
del snapshot0.txt
del snapshot1.txt
del snapshot2.txt
start "0" cmd /k "go run useDIMEX-f.go 0 127.0.0.1:5000 127.0.0.1:6001 127.0.0.1:7002 > t3.txt"
start "1" cmd /k "go run useDIMEX-f.go 1 127.0.0.1:5000 127.0.0.1:6001 127.0.0.1:7002 > t2.txt"
start "2" cmd /k "go run useDIMEX-f.go 2 127.0.0.1:5000 127.0.0.1:6001 127.0.0.1:7002 > t1.txt"