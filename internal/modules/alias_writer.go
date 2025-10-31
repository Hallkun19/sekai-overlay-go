package modules

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"sekai-overlay-go/internal/utils"
)

// GenerateAliasObject はエイリアスオブジェクトを生成する
func GenerateAliasObject(levelID, distDir string, lastNoteTime float64, extraData map[string]interface{}) (string, error) {
	fmt.Println("エイリアスオブジェクトの生成を開始します...")

	// テンプレートファイルのパスを取得
	templatePath := utils.ResourcePath("assets/alias/template.object")
	levelJSONPath := filepath.Join(distDir, "level.json")
	outputPath := filepath.Join(distDir, "main.object")

	// テンプレートファイルを読み込み
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("テンプレートファイルの読み込みに失敗しました: %w", err)
	}

	// level.jsonを読み込み
	levelFile, err := os.Open(levelJSONPath)
	if err != nil {
		return "", fmt.Errorf("level.jsonの読み込みに失敗しました: %w", err)
	}
	defer levelFile.Close()

	var levelData map[string]interface{}
	if err := json.NewDecoder(levelFile).Decode(&levelData); err != nil {
		return "", fmt.Errorf("level.jsonの解析に失敗しました: %w", err)
	}

	// プレースホルダー用の値を取得
	itemData, _ := levelData["item"].(map[string]interface{})

	finalTitle := getStringValue(extraData, "title", itemData, "title", "-")
	finalAuthor := getStringValue(extraData, "author", itemData, "author", "-")

	difficultyInput := getStringValue(extraData, "difficulty", nil, "", "custom")
	standardDifficulties := []string{"easy", "normal", "hard", "expert", "master", "append"}
	difficultyImgVal := strings.ToLower(difficultyInput)
	isStandard := false
	for _, std := range standardDifficulties {
		if difficultyImgVal == std {
			isStandard = true
			break
		}
	}
	if !isStandard {
		difficultyImgVal = "custom"
	}

	vocalInput := getStringValue(extraData, "vocal", nil, "", "")
	var vocalText string
	if vocalInput != "" {
		vocalText = fmt.Sprintf("Vo. %s", vocalInput)
	} else {
		vocalText = "Inst. ver."
	}

	// 置換マップを作成
	replacements := map[string]string{
		"{title}":          finalTitle,
		"{author}":         finalAuthor,
		"{words}":          getStringValue(extraData, "words", nil, "", "-"),
		"{music}":          getStringValue(extraData, "music", nil, "", "-"),
		"{arrange}":        getStringValue(extraData, "arrange", nil, "", "-"),
		"{vocal}":          vocalText,
		"{difficulty}":     strings.ToUpper(difficultyInput),
		"{difficulty_img}": difficultyImgVal,
	}

	// 空白だった場合のデフォルト値を設定
	for key, value := range replacements {
		if value == "" {
			if key == "{vocal}" {
				replacements[key] = "Inst. ver."
			} else {
				replacements[key] = "-"
			}
		}
	}

	// パス情報
	distFullPath := strings.ReplaceAll(filepath.ToSlash(distDir), "/", "\\")
	assetsFullPath := strings.ReplaceAll(filepath.ToSlash(utils.ResourcePath("assets")), "/", "\\")
	replacements["{distPath}"] = distFullPath
	replacements["{assetsPath}"] = assetsFullPath

	// フレーム計算
	videoStartFrame := int(math.Round((lastNoteTime+1.0)*60)) + 316
	fadeStartFrame := videoStartFrame + 161
	fadeStopFrame := fadeStartFrame + 142
	endFrame := fadeStopFrame + 124

	replacements["{videoStartFrame}"] = fmt.Sprintf("%d", videoStartFrame)
	replacements["{fadeStartFrame}"] = fmt.Sprintf("%d", fadeStartFrame)
	replacements["{fadeStopFrame}"] = fmt.Sprintf("%d", fadeStopFrame)
	replacements["{endFrame}"] = fmt.Sprintf("%d", endFrame)

	// 文字列を一括置換
	outputContent := string(templateContent)
	for placeholder, value := range replacements {
		outputContent = strings.ReplaceAll(outputContent, placeholder, value)
	}

	// 結果を書き出し
	if err := os.WriteFile(outputPath, []byte(outputContent), 0644); err != nil {
		return "", fmt.Errorf("出力ファイルの書き込みに失敗しました: %w", err)
	}

	fmt.Printf("エイリアスオブジェクトを '%s' に保存しました。\n", outputPath)
	return finalTitle, nil
}

// getStringValue はマップから文字列値を取得するヘルパー関数
func getStringValue(primary map[string]interface{}, primaryKey string, secondary map[string]interface{}, secondaryKey, defaultValue string) string {
	if primary != nil {
		if value, exists := primary[primaryKey]; exists {
			if str, ok := value.(string); ok && str != "" {
				return str
			}
		}
	}
	if secondary != nil {
		if value, exists := secondary[secondaryKey]; exists {
			if str, ok := value.(string); ok && str != "" {
				return str
			}
		}
	}
	return defaultValue
}
