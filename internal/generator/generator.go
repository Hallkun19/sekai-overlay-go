package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sekai-overlay-go/internal/config"
	"sekai-overlay-go/internal/modules"
	"sekai-overlay-go/internal/ui"
	"sekai-overlay-go/internal/utils"
)

// Generator は全ての生成処理を管理する構造体
type Generator struct {
	config  config.Config
	console *ui.Console
	appRoot string
}

// NewGenerator は新しいGeneratorインスタンスを作成する
func NewGenerator(cfg config.Config, console *ui.Console) *Generator {
	return &Generator{
		config:  cfg,
		console: console,
		appRoot: utils.GetAppRoot(),
	}
}

// Run は全ての生成処理を実行する
func (g *Generator) Run() error {
	var fullLevelID string

	// 譜面IDの検証と分割
	fullLevelIDInput := g.config.FullLevelID
	if !strings.Contains(fullLevelIDInput, "-") {
		return fmt.Errorf("無効な譜面ID形式です (例: chcy-test-1)")
	}

	parts := strings.Split(fullLevelIDInput, "-")
	if len(parts) < 2 {
		return fmt.Errorf("無効な譜面ID形式です (例: chcy-test-1)")
	}

	prefix := strings.Join(parts[:len(parts)-1], "-")
	idPart := parts[len(parts)-1]
	fullLevelID = fmt.Sprintf("%s-%s", prefix, idPart)

	// 出力先ディレクトリの作成
	distDir := filepath.Join(g.appRoot, "dist", fullLevelID)
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗しました: %w", err)
	}

	// 1. ダウンロード
	g.console.PrintStatus(fmt.Sprintf("[%s] データをダウンロード中...", fullLevelID))
	levelID, err := modules.DownloadAndPrepareAssets(prefix, idPart, distDir)
	if err != nil {
		return fmt.Errorf("データダウンロードに失敗しました: %w", err)
	}

	// 2. 背景画像生成
	g.console.PrintStatus("背景画像を生成中...")
	if err := modules.GenerateBackgroundImage(levelID, g.config.BgVersion, distDir); err != nil {
		return fmt.Errorf("背景画像生成に失敗しました: %w", err)
	}

	// 3. スコアオブジェクト生成
	g.console.PrintStatus("スコアオブジェクトを生成中...")
	lastNoteTime, err := modules.GenerateSkobjData(levelID, distDir, g.config.TeamPower, g.config.AppVersion)
	if err != nil {
		return fmt.Errorf("スコアオブジェクト生成に失敗しました: %w", err)
	}

	// 4. エイリアスオブジェクト生成
	g.console.PrintStatus("エイリアスオブジェクトを生成中...")
	title, err := modules.GenerateAliasObject(levelID, distDir, lastNoteTime, g.config.ExtraData)
	if err != nil {
		return fmt.Errorf("エイリアスオブジェクト生成に失敗しました: %w", err)
	}

	// 5. クリーンアップ
	g.cleanup(distDir)

	// 6. 出力フォルダを開く
	g.console.PrintStatus("出力フォルダを開いています...")
	g.openOutputFolder(distDir)

	g.console.PrintSuccess(fmt.Sprintf("譜面 '%s' のファイル生成が完了しました。", title))
	return nil
}

// cleanup は一時ファイルをクリーンアップする
func (g *Generator) cleanup(distDir string) {
	g.console.PrintStatus("一時ファイルをクリーンアップ中...")

	filesToRemove := []string{"level.json", "chart.json"}
	for _, filename := range filesToRemove {
		path := filepath.Join(distDir, filename)
		if _, err := os.Stat(path); err == nil {
			if err := os.Remove(path); err != nil {
				g.console.PrintError(fmt.Sprintf("ファイル削除エラー (%s): %v", filename, err))
			}
		}
	}
}

// openOutputFolder は出力フォルダをOSのファイルエクスプローラーで開く
func (g *Generator) openOutputFolder(path string) {
	if err := g.console.OpenFolder(path); err != nil {
		// フォルダを開けなくても処理全体は成功しているので、エラーにはしない
		fmt.Printf("出力フォルダを自動で開けませんでした: %v\n", err)
	}
}
