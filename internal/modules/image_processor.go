package modules

import (
	"fmt"
	"image"
	"path/filepath"

	"sekai-overlay-go/internal/utils"

	"github.com/disintegration/imaging"
)

// GenerateBackgroundImage は背景画像を生成する
func GenerateBackgroundImage(levelID, version, distDir string) error {
	fmt.Println("背景画像の生成を開始します...")

	coverImagePath := filepath.Join(distDir, "jacket.jpg")
	outputImagePath := filepath.Join(distDir, "background.png")

	// カバー画像を読み込み
	targetImage, err := imaging.Open(coverImagePath)
	if err != nil {
		return fmt.Errorf("カバー画像の読み込みに失敗しました: %w", err)
	}

	var finalImage *image.NRGBA

	// バージョンに応じてレンダリング
	switch version {
	case "3":
		finalImage, err = renderV3(targetImage)
	case "1":
		finalImage, err = renderV1(targetImage)
	default:
		return fmt.Errorf("バージョン '%s' は現在サポートされていません", version)
	}

	if err != nil {
		return fmt.Errorf("背景画像の生成に失敗しました: %w", err)
	}

	// 生成した画像を保存
	if err := imaging.Save(finalImage, outputImagePath); err != nil {
		return fmt.Errorf("背景画像の保存に失敗しました: %w", err)
	}

	fmt.Printf("背景画像を '%s' に保存しました。\n", outputImagePath)
	return nil
}

// renderV3 はv3の背景画像を生成する
func renderV3(targetImage image.Image) (*image.NRGBA, error) {
	// アセット画像の読み込み
	base, err := loadAssetImage("assets/background/v3/base.png")
	if err != nil {
		return nil, err
	}

	bottom, err := loadAssetImage("assets/background/v3/bottom.png")
	if err != nil {
		return nil, err
	}

	centerCover, err := loadAssetImage("assets/background/v3/center_cover.png")
	if err != nil {
		return nil, err
	}

	centerMask, err := loadAssetImage("assets/background/v3/center_mask.png")
	if err != nil {
		return nil, err
	}

	sideCover, err := loadAssetImage("assets/background/v3/side_cover.png")
	if err != nil {
		return nil, err
	}

	sideMask, err := loadAssetImage("assets/background/v3/side_mask.png")
	if err != nil {
		return nil, err
	}

	windows, err := loadAssetImage("assets/background/v3/windows.png")
	if err != nil {
		return nil, err
	}

	baseBounds := base.Bounds()
	baseSize := image.Rect(0, 0, baseBounds.Dx(), baseBounds.Dy())

	// サイドジャケットの生成
	sideJackets := imaging.New(baseBounds.Dx(), baseBounds.Dy(), image.Transparent)

	// モーフィング処理（簡易実装）
	leftNormal := morphImage(targetImage, []image.Point{{566, 161}, {1183, 134}, {633, 731}, {1226, 682}}, baseSize)
	rightNormal := morphImage(targetImage, []image.Point{{966, 104}, {1413, 72}, {954, 525}, {1390, 524}}, baseSize)
	leftMirror := morphImage(targetImage, []image.Point{{633, 1071}, {1256, 1045}, {598, 572}, {1197, 569}}, baseSize)
	rightMirror := morphImage(targetImage, []image.Point{{954, 1122}, {1393, 1167}, {942, 702}, {1366, 717}}, baseSize)

	sideJackets = imaging.Overlay(sideJackets, leftNormal, image.Point{}, 1.0)
	sideJackets = imaging.Overlay(sideJackets, rightNormal, image.Point{}, 1.0)
	sideJackets = imaging.Overlay(sideJackets, leftMirror, image.Point{}, 1.0)
	sideJackets = imaging.Overlay(sideJackets, rightMirror, image.Point{}, 1.0)
	sideJackets = imaging.Overlay(sideJackets, sideCover, image.Point{}, 1.0)

	// センタージャケットの生成
	center := imaging.New(baseBounds.Dx(), baseBounds.Dy(), image.Transparent)
	centerNormal := morphImage(targetImage, []image.Point{{824, 227}, {1224, 227}, {833, 608}, {1216, 608}}, baseSize)
	centerMirror := morphImage(targetImage, []image.Point{{830, 1017}, {1214, 1017}, {833, 676}, {1216, 676}}, baseSize)

	center = imaging.Overlay(center, centerNormal, image.Point{}, 1.0)
	center = imaging.Overlay(center, centerMirror, image.Point{}, 1.0)
	center = imaging.Overlay(center, centerCover, image.Point{}, 1.0)

	// マスキング処理
	sideJackets = applyMask(sideJackets, sideMask)
	center = applyMask(center, centerMask)

	// 最終的な合成
	finalImage := imaging.Clone(base)
	finalImage = imaging.Overlay(finalImage, sideJackets, image.Point{}, 1.0)
	finalImage = imaging.Overlay(finalImage, sideCover, image.Point{}, 1.0)
	finalImage = imaging.Overlay(finalImage, windows, image.Point{}, 1.0)
	finalImage = imaging.Overlay(finalImage, center, image.Point{}, 1.0)
	finalImage = imaging.Overlay(finalImage, bottom, image.Point{}, 1.0)

	return finalImage, nil
}

// renderV1 はv1の背景画像を生成する
func renderV1(targetImage image.Image) (*image.NRGBA, error) {
	// アセット画像の読み込み
	base, err := loadAssetImage("assets/background/v1/base.png")
	if err != nil {
		return nil, err
	}

	sideMask, err := loadAssetImage("assets/background/v1/side_mask.png")
	if err != nil {
		return nil, err
	}

	centerMask, err := loadAssetImage("assets/background/v1/center_mask.png")
	if err != nil {
		return nil, err
	}

	mirrorMask, err := loadAssetImage("assets/background/v1/mirror_mask.png")
	if err != nil {
		return nil, err
	}

	frames, err := loadAssetImage("assets/background/v1/frames.png")
	if err != nil {
		return nil, err
	}

	baseBounds := base.Bounds()
	baseSize := image.Rect(0, 0, baseBounds.Dx(), baseBounds.Dy())

	// サイドジャケットの生成
	sideJackets := imaging.New(baseBounds.Dx(), baseBounds.Dy(), image.Transparent)
	leftNormal := morphImage(targetImage, []image.Point{{449, 114}, {1136, 99}, {465, 804}, {1152, 789}}, baseSize)
	rightNormal := morphImage(targetImage, []image.Point{{1018, 92}, {1635, 51}, {1026, 756}, {1630, 740}}, baseSize)

	sideJackets = imaging.Overlay(sideJackets, leftNormal, image.Point{}, 1.0)
	sideJackets = imaging.Overlay(sideJackets, rightNormal, image.Point{}, 1.0)

	// センタージャケットの生成
	center := imaging.New(baseBounds.Dx(), baseBounds.Dy(), image.Transparent)
	centerNormal := morphImage(targetImage, []image.Point{{798, 193}, {1252, 193}, {801, 635}, {1246, 635}}, baseSize)
	centerMirror := morphImage(targetImage, []image.Point{{798, 1152}, {1252, 1152}, {795, 713}, {1252, 713}}, baseSize)

	centerNormal = applyMask(centerNormal, centerMask)
	centerMirror = applyMask(centerMirror, mirrorMask)

	center = imaging.Overlay(center, centerNormal, image.Point{}, 1.0)
	center = imaging.Overlay(center, centerMirror, image.Point{}, 1.0)

	sideJackets = applyMask(sideJackets, sideMask)

	// 最終的な合成
	finalImage := imaging.Clone(base)
	finalImage = imaging.Overlay(finalImage, sideJackets, image.Point{}, 1.0)
	finalImage = imaging.Overlay(finalImage, center, image.Point{}, 1.0)
	finalImage = imaging.Overlay(finalImage, frames, image.Point{}, 1.0)

	return finalImage, nil
}

// loadAssetImage はアセット画像を読み込む
func loadAssetImage(assetPath string) (image.Image, error) {
	fullPath := utils.ResourcePath(assetPath)
	return imaging.Open(fullPath)
}

// morphImage は画像をモーフィングする（簡易実装）
func morphImage(src image.Image, targetCoords []image.Point, targetSize image.Rectangle) *image.NRGBA {
	// 簡易的なリサイズのみ実装
	// 本来は射影変換が必要だが、複雑なので一旦リサイズで代用
	width := targetSize.Dx()
	height := targetSize.Dy()

	return imaging.Resize(src, width, height, imaging.Lanczos)
}

// applyMask はマスクを適用する
func applyMask(img, mask image.Image) *image.NRGBA {
	// 簡易的なアルファブレンド
	return imaging.Overlay(img, mask, image.Point{}, 1.0)
}
