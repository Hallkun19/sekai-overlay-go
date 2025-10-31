package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

// GetAppRoot はアプリケーションのルートパスを取得する
func GetAppRoot() string {
	// 開発環境(.go)とビルド環境(.exe)の両方で動作する
	if _, ok := os.LookupEnv("SNAP"); ok {
		return filepath.Dir(os.Args[0])
	}

	exec, err := os.Executable()
	if err != nil {
		panic(err)
	}

	return filepath.Dir(exec)
}

// ResourcePath はアセットへの絶対パスを取得する（読み込み専用）
func ResourcePath(relativePath string) string {
	basePath := GetAppRoot()

	// ビルド環境の場合、_internalフォルダを確認
	internalPath := filepath.Join(basePath, "_internal")
	if _, err := os.Stat(internalPath); err == nil {
		basePath = internalPath
	}

	return filepath.Join(basePath, relativePath)
}

// IsAdmin は現在のプロセスが管理者権限で実行されているかを確認する（Windows専用）
func IsAdmin() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	isAdmin := kernel32.NewProc("IsUserAnAdmin")
	ret, _, _ := isAdmin.Call()
	return ret != 0
}

// RunAsAdmin はアプリケーションを管理者権限で再起動する（Windows専用）
func RunAsAdmin() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	// ShellExecuteWを使用
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecute := shell32.NewProc("ShellExecuteW")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	argsPtr, _ := syscall.UTF16PtrFromString(args)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)

	ret, _, _ := shellExecute.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(exePtr)),
		uintptr(unsafe.Pointer(argsPtr)),
		uintptr(unsafe.Pointer(cwdPtr)),
		1, // SW_SHOWNORMAL
	)

	if ret <= 32 {
		return fmt.Errorf("管理者権限での再起動に失敗しました")
	}

	return nil
}
