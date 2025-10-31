package modules

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"sekai-overlay-go/internal/config"

	"golang.org/x/image/draw"
)

// DownloadAndPrepareAssets は指定サーバーから譜面データをダウンロードし、ジャケットをリサイズする
func DownloadAndPrepareAssets(prefix, idPart, distDir string) (string, error) {
	baseURL, exists := config.ServerMap[prefix]
	if !exists {
		return "", fmt.Errorf("サポートされていないサーバー接頭辞です: %s", prefix)
	}

	apiURL := fmt.Sprintf("%s%s-%s", baseURL, prefix, idPart)
	fullLevelID := fmt.Sprintf("%s-%s", prefix, idPart)

	fmt.Printf("APIにアクセスしています: %s\n", apiURL)

	// APIリクエスト
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("APIリクエストに失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("APIレスポンスエラー: %d", resp.StatusCode)
	}

	var apiResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", fmt.Errorf("JSONデコードに失敗しました: %w", err)
	}

	// ディレクトリ作成
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return "", fmt.Errorf("ディレクトリ作成に失敗しました: %w", err)
	}

	// level.json保存
	levelPath := filepath.Join(distDir, "level.json")
	levelFile, err := os.Create(levelPath)
	if err != nil {
		return "", fmt.Errorf("level.json作成に失敗しました: %w", err)
	}
	defer levelFile.Close()

	encoder := json.NewEncoder(levelFile)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(apiResponse); err != nil {
		return "", fmt.Errorf("level.json書き込みに失敗しました: %w", err)
	}

	fmt.Printf("ファイルを '%s' に保存します。\n", distDir)

	item, ok := apiResponse["item"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("APIレスポンスにitemフィールドが見つかりません")
	}

	// ジャケットダウンロードとリサイズ
	cover, ok := item["cover"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("cover情報が見つかりません")
	}
	coverURL, ok := cover["url"].(string)
	if !ok {
		return "", fmt.Errorf("cover URLが見つかりません")
	}

	jacketPath := filepath.Join(distDir, "jacket.jpg")
	if err := downloadFile(coverURL, jacketPath); err != nil {
		return "", fmt.Errorf("ジャケットダウンロードに失敗しました: %w", err)
	}

	if err := resizeJacket(jacketPath); err != nil {
		return "", fmt.Errorf("ジャケットリサイズに失敗しました: %w", err)
	}

	// BGMダウンロード
	bgm, ok := item["bgm"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("bgm情報が見つかりません")
	}
	bgmURL, ok := bgm["url"].(string)
	if !ok {
		return "", fmt.Errorf("bgm URLが見つかりません")
	}

	musicPath := filepath.Join(distDir, "music.mp3")
	if err := downloadFile(bgmURL, musicPath); err != nil {
		return "", fmt.Errorf("BGMダウンロードに失敗しました: %w", err)
	}

	// チャートデータダウンロードと解凍
	data, ok := item["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("data情報が見つかりません")
	}
	dataURL, ok := data["url"].(string)
	if !ok {
		return "", fmt.Errorf("data URLが見つかりません")
	}

	chartGzPath := filepath.Join(distDir, "chart.json.gz")
	if err := downloadFile(dataURL, chartGzPath); err != nil {
		return "", fmt.Errorf("チャートデータダウンロードに失敗しました: %w", err)
	}

	chartPath := filepath.Join(distDir, "chart.json")
	if err := unzipGz(chartGzPath, chartPath); err != nil {
		return "", fmt.Errorf("チャートデータ解凍に失敗しました: %w", err)
	}

	return fullLevelID, nil
}

// downloadFile はファイルをダウンロードする
func downloadFile(url, destPath string) error {
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

// resizeJacket はジャケット画像をリサイズする
func resizeJacket(imagePath string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// サイズが512x512でない場合のみリサイズ
	targetSize := 512
	if img.Bounds().Dx() == targetSize && img.Bounds().Dy() == targetSize {
		return nil
	}

	fmt.Printf("  -> jacket.jpgを%dx%dにリサイズしています...\n", targetSize, targetSize)

	// 新しい画像を作成
	dst := image.NewRGBA(image.Rect(0, 0, targetSize, targetSize))

	// リサイズ
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	// ファイルを書き込みモードで開き直す
	file, err = os.Create(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// JPEGとして保存
	return jpeg.Encode(file, dst, &jpeg.Options{Quality: 95})
}

// unzipGz はgzファイルを解凍する
func unzipGz(gzPath, destPath string) error {
	gzFile, err := os.Open(gzPath)
	if err != nil {
		return err
	}
	defer gzFile.Close()

	gzReader, err := gzip.NewReader(gzFile)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, gzReader)
	if err != nil {
		return err
	}

	// ファイルを閉じてから削除する
	gzReader.Close()
	gzFile.Close()

	// 元のgzファイルを削除
	return os.Remove(gzPath)
}
