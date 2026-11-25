@echo off
for %%S in ("Directory\Background\shell" "Directory\shell" "*\shell") do (
 reg delete "HKCU\Software\Classes\%%~S\HoI4ModNavigator.3C81A9" /f >nul 2>&1
 for /l %%I in (0,1,79) do call :remove "%%~S" %%I
)
reg delete "HKCU\Software\Classes\Directory\Background\shellex\ContextMenuHandlers\{76AE4D21-CAD1-4B8E-A79E-8555CAB57101}" /f >nul 2>&1
reg delete "HKCU\Software\Classes\Directory\shellex\ContextMenuHandlers\{76AE4D21-CAD1-4B8E-A79E-8555CAB57102}" /f >nul 2>&1
reg delete "HKCU\Software\Classes\*\shellex\ContextMenuHandlers\{76AE4D21-CAD1-4B8E-A79E-8555CAB57103}" /f >nul 2>&1
for %%G in (01 02 03) do reg delete "HKCU\Software\Classes\CLSID\{76AE4D21-CAD1-4B8E-A79E-8555CAB571%%G}" /f >nul 2>&1
reg delete "HKCU\Software\HoI4ModNavigator\ShellV2" /f >nul 2>&1
echo Restart Windows Explorer via Task Manager if needed. Settings and game/mod files are unchanged.
pause
exit /b
:remove
set "IDX=00%~2"
reg delete "HKCU\Software\Classes\%~1\HoI4ModNavigator.3C81A9.cmd%IDX:~-3%" /f >nul 2>&1
exit /b
