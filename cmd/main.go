package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sekai-overlay-go/internal/config"
	"sekai-overlay-go/internal/generator"
	"sekai-overlay-go/internal/modules"
	"sekai-overlay-go/internal/ui"
)

func main() {
	console := ui.NewConsole()
	console.PrintBanner()

	// 起動時に最新リリースとobj2の状態を確認して通知する（自動置換は行わない）
	console.PrintInfo("起動チェック: 最新リリースと @SekaiObjects.obj2 の状態を確認します...")
	if err := modules.CheckAndNotifyUpdates(console); err != nil {
		console.PrintError(fmt.Sprintf("起動チェック中にエラーが発生しました: %v", err))
	}

	// メインループ
	for {
		showMainMenu(console)
		choice := getUserChoice(console)

		switch choice {
		case "1":
			runGeneration(console)
		case "2":
			runSetup(console)
		case "3":
			console.PrintInfo("ご利用ありがとうございました。")
			os.Exit(0)
		default:
			console.PrintError("無効な選択です。1-3の数字を入力してください。")
		}

		console.PrintInfo("\n続けるにはEnterキーを押してください...")
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
		console.ClearScreen()
		console.PrintBanner()
	}
}

func showMainMenu(console *ui.Console) {
	console.PrintHeader("メインメニュー")
	fmt.Println("1. 譜面データ生成")
	fmt.Println("2. セットアップ")
	fmt.Println("3. 終了")
	fmt.Print("\n選択してください (1-3): ")
}

func getUserChoice(console *ui.Console) string {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		console.PrintError("入力エラー")
		return ""
	}
	return strings.TrimSpace(input)
}

func runGeneration(console *ui.Console) {
	console.PrintHeader("譜面データ生成")

	// 譜面IDの入力
	console.PrintInfo("譜面IDを入力してください (例: chcy-XXXX): ")
	levelID := getUserChoice(console)
	if levelID == "" {
		console.PrintError("譜面IDは必須です。")
		return
	}

	// 曲タイトルの入力
	console.PrintInfo("曲タイトルを入力してください (空白でlevel.jsonの値を使用): ")
	title := getUserChoice(console)

	// 譜面制作者の入力
	console.PrintInfo("譜面制作者を入力してください (空白でlevel.jsonの値を使用): ")
	author := getUserChoice(console)

	// チーム総合力の入力
	console.PrintInfo("チーム総合力を入力してください (デフォルト: 250000): ")
	powerInput := getUserChoice(console)
	teamPower := 250000.0
	if powerInput != "" {
		if power, err := strconv.ParseFloat(powerInput, 64); err == nil {
			teamPower = power
		} else {
			console.PrintError("無効な数値です。デフォルト値を使用します。")
		}
	}

	// 背景バージョンの選択
	console.PrintInfo("背景バージョンを選択してください (3または1、デフォルト: 3): ")
	versionInput := getUserChoice(console)
	bgVersion := "3"
	if versionInput == "1" {
		bgVersion = "1"
	} else if versionInput != "" && versionInput != "3" {
		console.PrintError("無効なバージョンです。デフォルト値(3)を使用します。")
	}

	// 難易度の入力
	console.PrintInfo("難易度を入力してください (デフォルト: master): ")
	difficulty := getUserChoice(console)
	if difficulty == "" {
		difficulty = "master"
	}

	// 設定の作成
	cfg := config.Config{
		FullLevelID: levelID,
		BgVersion:   bgVersion,
		TeamPower:   teamPower,
		AppVersion:  config.AppVersion,
		ExtraData: map[string]interface{}{
			"difficulty": difficulty,
			"title":      title,
			"author":     author,
		},
	}

	// 生成前に設定サマリを表示
	console.PrintInfo("生成設定:")
	summary := map[string]string{
		"譜面ID":    cfg.FullLevelID,
		"背景バージョン": cfg.BgVersion,
		"チーム総合力":  fmt.Sprintf("%.0f", cfg.TeamPower),
		"難易度":     fmt.Sprintf("%v", cfg.ExtraData["difficulty"]),
		"タイトル":    fmt.Sprintf("%v", cfg.ExtraData["title"]),
		"作者":      fmt.Sprintf("%v", cfg.ExtraData["author"]),
	}
	console.PrintKVTable(summary)

	// ジェネレータの実行
	gen := generator.NewGenerator(cfg, console)
	if err := gen.Run(); err != nil {
		console.PrintError(fmt.Sprintf("生成処理に失敗しました: %v", err))
		return
	}

	console.PrintSuccess("処理が完了しました！")
}

func runSetup(console *ui.Console) {
	console.PrintHeader("セットアップ")

	console.PrintInfo("セットアップを開始します...")
	console.PrintInfo("この処理には管理者権限が必要になる場合があります。")
	console.PrintInfo("続行しますか？ (y/N): ")

	choice := getUserChoice(console)
	if strings.ToLower(choice) != "y" && strings.ToLower(choice) != "yes" {
		console.PrintInfo("セットアップをキャンセルしました。")
		return
	}

	if err := modules.CheckAndRunSetup(); err != nil {
		console.PrintError(fmt.Sprintf("セットアップに失敗しました: %v", err))
		return
	}

	console.PrintSuccess("セットアップが完了しました。")
}
