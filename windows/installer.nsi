; rl-stats-daemon NSIS installer
;
; Build:
;   makensis \
;     -DVERSION=1.2.3 \
;     -DBINARY=/abs/path/to/rl-stats-daemon.exe \
;     -DOUTFILE=/abs/path/to/rl-stats-daemon_setup.exe \
;     windows/installer.nsi

!define APP_NAME "rl-stats-daemon"
!define COMP_NAME "provant.gg"
!define WEB_SITE  "https://github.com/provant-gg/daemon"
!define UNINSTALL_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"

!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!ifndef BINARY
  !error "BINARY must be defined: -DBINARY=path/to/rl-stats-daemon.exe"
!endif
!ifndef OUTFILE
  !define OUTFILE "${APP_NAME}_setup.exe"
!endif

Name "${APP_NAME} ${VERSION}"
OutFile "${OUTFILE}"
InstallDir "$PROGRAMFILES64\${APP_NAME}"
InstallDirRegKey HKLM "Software\${APP_NAME}" "InstallDir"
RequestExecutionLevel admin
ShowInstDetails show
ShowUninstDetails show
Unicode true
SetCompressor /SOLID lzma

VIProductVersion "${VERSION}.0"
VIAddVersionKey "ProductName"     "${APP_NAME}"
VIAddVersionKey "CompanyName"     "${COMP_NAME}"
VIAddVersionKey "FileVersion"     "${VERSION}"
VIAddVersionKey "ProductVersion"  "${VERSION}"
VIAddVersionKey "FileDescription" "${APP_NAME} installer"
VIAddVersionKey "LegalCopyright"  "${COMP_NAME}"

!include "MUI2.nsh"
!include "StrFunc.nsh"
${StrStr}

!define MUI_ABORTWARNING

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Section "Install"
  SetOutPath "$INSTDIR"
  File /oname=${APP_NAME}.exe "${BINARY}"
  WriteUninstaller "$INSTDIR\Uninstall.exe"

  ; Append INSTDIR to system PATH (idempotent).
  ReadRegStr $0 HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path"
  ${StrStr} $1 "$0" "$INSTDIR"
  StrCmp $1 "" 0 pathDone
    StrCpy $0 "$0;$INSTDIR"
    WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path" "$0"
    System::Call 'User32::SendMessageTimeout(p 0xffff, i 0x1A, p 0, t "Environment", i 0, i 5000, *p .r2)'
  pathDone:

  ; Add/Remove Programs registration.
  WriteRegStr   HKLM "${UNINSTALL_KEY}" "DisplayName"     "${APP_NAME}"
  WriteRegStr   HKLM "${UNINSTALL_KEY}" "DisplayVersion"  "${VERSION}"
  WriteRegStr   HKLM "${UNINSTALL_KEY}" "Publisher"       "${COMP_NAME}"
  WriteRegStr   HKLM "${UNINSTALL_KEY}" "URLInfoAbout"    "${WEB_SITE}"
  WriteRegStr   HKLM "${UNINSTALL_KEY}" "InstallLocation" "$INSTDIR"
  WriteRegStr   HKLM "${UNINSTALL_KEY}" "UninstallString" "$\"$INSTDIR\Uninstall.exe$\""
  WriteRegDWORD HKLM "${UNINSTALL_KEY}" "NoModify"        1
  WriteRegDWORD HKLM "${UNINSTALL_KEY}" "NoRepair"        1
  WriteRegStr   HKLM "Software\${APP_NAME}" "InstallDir"  "$INSTDIR"
SectionEnd

Section "Uninstall"
  Delete "$INSTDIR\${APP_NAME}.exe"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir  "$INSTDIR"

  DeleteRegKey HKLM "${UNINSTALL_KEY}"
  DeleteRegKey HKLM "Software\${APP_NAME}"

  ; PATH entry intentionally not removed on uninstall: a stale PATH
  ; entry pointing at a non-existent directory is harmless, and the
  ; search/replace logic to remove a single segment is brittle without
  ; an external plugin.
SectionEnd
