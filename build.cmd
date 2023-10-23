@echo off

echo Building...
set GOOS=linux

go build -o status

set GOOS=windows