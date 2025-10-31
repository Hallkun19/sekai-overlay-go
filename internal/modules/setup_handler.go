package modules

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"sekai-overlay-go/internal/config"
	"sekai-overlay-go/internal/ui"
	"sekai-overlay-go/internal/utils"

	"gopkg.in/ini.v1"
)

// CheckAndRunSetup は設定をチェックし、必要な場合はセットアップを実行する
func CheckAndRunSetup() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	storedVersion := cfg.Section("AppInfo").Key("LastVersion").String()
	setupComplete := cfg.Section("AppInfo").Key("SetupComplete").String() == "true"

	tasks := []string{}
	if storedVersion != config.AppVersion || !setupComplete {
		tasks = append(tasks, "update_obj")
	}
	if !setupComplete {
		tasks = append(tasks, "install_anm")
	}

	if len(tasks) == 0 {
		return nil
	}

	fmt.Println("セットアップを開始します...")

	successMessages := []string{}

	for _, task := range tasks {
		switch task {
		case "update_obj":
			if err := installObjScript(); err != nil {
				return fmt.Errorf("OBJスクリプトのインストールに失敗しました: %w", err)
			}
			if err := updateConfigFile("LastVersion", config.AppVersion); err != nil {
				return fmt.Errorf("設定ファイルの更新に失敗しました: %w", err)
			}
			successMessages = append(successMessages, "・'@SekaiObjects.obj2' をインストール/更新しました。")

		case "install_anm":
			if err := installAnmScript(); err != nil {
				return fmt.Errorf("ANMスクリプトのインストールに失敗しました: %w", err)
			}
			if err := updateConfigFile("SetupComplete", "true"); err != nil {
				return fmt.Errorf("設定ファイルの更新に失敗しました: %w", err)
			}
			successMessages = append(successMessages, "・'unmult.anm2', 'dkjson.lua' をインストールしました。")
		}
	}

	if len(successMessages) > 0 {
		fmt.Println("セットアップが完了しました:")
		for _, msg := range successMessages {
			fmt.Println(msg)
		}
	}

	return nil
}

// CheckAndNotifyUpdates は起動時に最新リリースとインストール済みの @SekaiObjects.obj2 のバージョンを
// 確認し、必要があれば通知する（自動置換は行わない）。
func CheckAndNotifyUpdates(console *ui.Console) error {
	// 最新リリースをGitHub APIから取得
	apiURL := "https://api.github.com/repos/hallkun19/sekai-overlay-go/releases/latest"
	resp, err := http.Get(apiURL)
	if err != nil {
		console.PrintError(fmt.Sprintf("最新リリースの確認に失敗しました: %v", err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		console.PrintError(fmt.Sprintf("最新リリースの確認でエラーが返されました: %d", resp.StatusCode))
		return fmt.Errorf("release check status: %d", resp.StatusCode)
	}

	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		console.PrintError(fmt.Sprintf("最新リリース情報の解析に失敗しました: %v", err))
		return err
	}

	latestTag := body.TagName // 例: v0.2.0
	if latestTag == "" {
		console.PrintError("最新リリースのtag_nameが取得できませんでした。")
		return fmt.Errorf("empty tag_name")
	}

	// アプリ自身のバージョンと比較して通知する
	if latestTag != config.AppVersion {
		console.PrintInfo(fmt.Sprintf("新しいリリースがあります: %s (現在のアプリバージョン: %s)", latestTag, config.AppVersion))
	} else {
		console.PrintInfo("アプリは最新バージョンです。")
	}

	// @SekaiObjects.obj2 のインストール済みバージョンを確認
	objPath := filepath.Join(config.AviUtlScriptDir, "@SekaiObjects.obj2")
	installedObjVer := ""
	if _, err := os.Stat(objPath); err == nil {
		content, err := os.ReadFile(objPath)
		if err == nil {
			lines := strings.Split(string(content), "\n")
			for _, ln := range lines {
				if strings.Contains(ln, "SKOBJ_VERSION") {
					// 例: SKOBJ_VERSION = "v0.1.0"
					parts := strings.SplitN(ln, "=", 2)
					if len(parts) == 2 {
						installedObjVer = strings.TrimSpace(parts[1])
						installedObjVer = strings.Trim(installedObjVer, " \"'")
					}
					break
				}
			}
		}
	}

	if installedObjVer == "" {
		console.PrintInfo("@SekaiObjects.obj2 が見つからないか、バージョン情報が読み取れませんでした。手動でセットアップを実行してください。")
	} else if installedObjVer != latestTag {
		console.PrintInfo(fmt.Sprintf("@SekaiObjects.obj2 が最新ではありません (installed: %s, latest: %s)。自動で置き換えは行いません。必要ならメニューのセットアップを実行してください。", installedObjVer, latestTag))
	} else {
		console.PrintInfo("@SekaiObjects.obj2 は最新です。")
	}

	return nil
}

// loadConfig は設定ファイルを読み込む
func loadConfig() (*ini.File, error) {
	if _, err := os.Stat(config.GetConfigPath()); os.IsNotExist(err) {
		// 設定ファイルが存在しない場合は新規作成
		return ini.Empty(), nil
	}
	return ini.Load(config.GetConfigPath())
}

// updateConfigFile は設定ファイルを更新する
func updateConfigFile(key, value string) error {
	os.MkdirAll(config.GetConfigDir(), 0755)

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if !cfg.HasSection("AppInfo") {
		cfg.NewSection("AppInfo")
	}

	cfg.Section("AppInfo").Key(key).SetValue(value)

	return cfg.SaveTo(config.GetConfigPath())
}

// checkWritePermission は指定されたパスへの書き込み権限があるかチェックする
func checkWritePermission(path string) bool {
	if err := os.MkdirAll(path, 0755); err != nil {
		return false
	}

	tempFile := filepath.Join(path, "temp_permission_check.tmp")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		return false
	}

	os.Remove(tempFile)
	return true
}

// installObjScript は@SekaiObjects.obj2をインストールする
func installObjScript() error {
	srcPath := utils.ResourcePath("assets/scripts/@SekaiObjects.obj2")
	destPath := filepath.Join(config.AviUtlScriptDir, "@SekaiObjects.obj2")

	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("ソースファイルの読み込みに失敗しました: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// 9行目 (インデックス8) を書き換える
	if len(lines) >= 9 {
		lines[8] = fmt.Sprintf("SKOBJ_VERSION = \"%s\"\n", config.AppVersion)
	}

	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(destPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗しました: %w", err)
	}

	fmt.Printf("'%s' へスクリプトをインストールしました。\n", destPath)
	return nil
}

// installAnmScript はunmult.anm2, dkjson.luaをダウンロードしてインストールする
func installAnmScript() error {
	// unmult.anm2のダウンロード
	destPath := filepath.Join(config.AviUtlScriptDir, "unmult.anm2")
	if err := downloadFileSetup(config.UnmultAnmURL, destPath); err != nil {
		return fmt.Errorf("unmult.anm2のダウンロードに失敗しました: %w", err)
	}

	// dkjson.luaのダウンロード
	destPath = filepath.Join(config.AviUtlScriptDir, "dkjson.lua")
	if err := downloadFileSetup(config.DkjsonLuaURL, destPath); err != nil {
		return fmt.Errorf("dkjson.luaのダウンロードに失敗しました: %w", err)
	}

	fmt.Printf("スクリプトを '%s' へダウンロード・インストールしました。\n", config.AviUtlScriptDir)
	return nil
}

// downloadFileSetup はファイルをダウンロードする（setup_handler用）
func downloadFileSetup(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ダウンロードエラー: %d", resp.StatusCode)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
