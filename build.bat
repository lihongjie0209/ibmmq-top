@echo off
REM build.bat - Build ibmmq-top binary and portable bundles inside Linux Docker container
REM Outputs:
REM   .\out\mq_top                         (standalone binary)
REM   .\out\mq_top_bundle_mq93.tar.gz      (portable bundle with MQ 9.3.0.37 LTS libs)
REM   .\out\mq_top_bundle_mq94.tar.gz      (portable bundle with MQ 9.4.5.0 libs)

setlocal

set IMAGE=mq-top-builder
set OUTDIR=%~dp0out

if not exist "%OUTDIR%" mkdir "%OUTDIR%"

echo [1/2] Building Docker image...
docker build -t %IMAGE% -f Dockerfile.build . || (echo Docker build failed & exit /b 1)

echo [2/2] Compiling binary and creating bundles...
docker run --rm -v "%OUTDIR%:/go/out" %IMAGE% || (echo Compilation failed & exit /b 1)

echo.
echo Build complete!
echo.
echo Outputs:
echo   %OUTDIR%\mq_top                        (standalone - requires MQ client installed on target)
echo   %OUTDIR%\mq_top_bundle_mq93.tar.gz     (portable bundle with MQ 9.3.0.37 LTS libs)
echo   %OUTDIR%\mq_top_bundle_mq94.tar.gz     (portable bundle with MQ 9.4.5.0 libs)
echo.
echo Usage - standalone (inside MQ container):
echo   docker cp %OUTDIR%\mq_top mqcontainer:/tmp/
echo   docker exec -it mqcontainer sh -c "TERM=xterm-256color /tmp/mq_top"
echo.
echo Usage - portable bundle (any Linux machine, no MQ client needed):
echo   tar xzf mq_top_bundle_mq94.tar.gz
echo   TERM=xterm-256color ./mq_top/run.sh -ibmmq.connName "hostname(1414)" -ibmmq.queueManager QM1
echo.
echo Note: bundle-mq93 targets MQ 9.2.x/9.3.x servers; bundle-mq94 targets MQ 9.4.x servers.
