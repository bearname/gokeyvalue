@echo off

start goserver.exe -instanceId=0 -port=8000 -isMaster=true -urls=0!localhost:8000,1!localhost:8001 -volume=inst-0\
@REM start goserver.exe -instanceId=0 -port=8000 -isMaster=true -urls=0!localhost:8000,1!localhost:8001,2!localhost:8002 -volume=inst-0\
@REM start goserver.exe -instanceId=1 -port=8001 -isMaster=false -urls=0!localhost:8000,1!localhost:8001,2!localhost:8002 -volume=inst-1\
start goserver.exe -instanceId=1 -port=8001 -isMaster=false -urls=0!localhost:8000,1!localhost:8001 -volume=inst-1\
@REM start goserver.exe -instanceId=1 -port=8002 -isMaster=false -urls=0!localhost:8000,1!localhost:8001,2!localhost:8002 -volume=inst-1\
@REM start goserver.exe -instanceId=2 -port=8002 -isMaster=false -urls=0!localhost:8000,1!localhost:8001,2!localhost:8002 -volume=inst-2\
@REM start goserver.exe -instanceId=3 -port=8003 -isMaster=false -urls=0!localhost:8000,1!localhost:8001,2!localhost:8002 -volume=inst-2\