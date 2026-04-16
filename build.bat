@echo off
REM build.bat - Build mq-top binary inside Linux Docker container
REM Output: .\out\mq_top  (Linux amd64 binary)

setlocal

set IMAGE=mq-top-builder
set OUTDIR=%~dp0out

if not exist "%OUTDIR%" mkdir "%OUTDIR%"

echo [1/2] Building Docker image...
docker build -t %IMAGE% -f Dockerfile.build . || (echo Docker build failed & exit /b 1)

echo [2/2] Compiling mq_top binary...
docker run --rm -v "%OUTDIR%:/go/out" %IMAGE% || (echo Compilation failed & exit /b 1)

echo.
echo Build complete!
echo.
echo Outputs:
echo   %OUTDIR%\mq_top                  (standalone binary - requires MQ client installed)
echo   %OUTDIR%\mq_top_bundle.tar.gz    (portable bundle  - no MQ client required)
echo.
echo Usage - standalone (MQ container has the client):
echo   docker cp %OUTDIR%\mq_top mqcontainer:/tmp/
echo   docker exec -it mqcontainer sh -c "TERM=xterm-256color /tmp/mq_top"
echo.
echo Usage - portable bundle (any Linux machine):
echo   tar xzf %OUTDIR%\mq_top_bundle.tar.gz
echo   TERM=xterm-256color ./mq_top/run.sh -ibmmq.connName "hostname(1414)" -ibmmq.queueManager QM1
