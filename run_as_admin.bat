@echo off
cd /d "%~dp0"
powershell -Command "Start-Process 'go' -ArgumentList 'run .' -Verb RunAs"
pause