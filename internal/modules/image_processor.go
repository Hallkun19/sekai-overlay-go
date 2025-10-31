package modules

import (
	"fmt"
	"image"
	"image/color"
	"math"
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

// morphImage は画像をモーフィングする（射影変換を使用）
func morphImage(src image.Image, targetCoords []image.Point, targetSize image.Rectangle) *image.NRGBA {
	if len(targetCoords) != 4 {
		return imaging.New(targetSize.Dx(), targetSize.Dy(), image.Transparent)
	}

	// 元画像のサイズを取得
	bounds := src.Bounds()
	srcNRGBA := imaging.Clone(src)

	// ソース座標（元画像の四隅）
	srcCoords := []image.Point{
		{bounds.Min.X, bounds.Min.Y},         // 左上
		{bounds.Max.X - 1, bounds.Min.Y},     // 右上
		{bounds.Min.X, bounds.Max.Y - 1},     // 左下
		{bounds.Max.X - 1, bounds.Max.Y - 1}, // 右下
	}

	// 射影変換行列を計算（ターゲット → ソースの逆変換）
	matrix := calculatePerspectiveMatrix(targetCoords, srcCoords)

	// 出力画像の作成
	dst := imaging.New(targetSize.Dx(), targetSize.Dy(), image.Transparent)

	// targetCoordsの範囲を計算
	minX := targetCoords[0].X
	minY := targetCoords[0].Y
	maxX := targetCoords[0].X
	maxY := targetCoords[0].Y
	for _, p := range targetCoords {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	// ターゲット領域内のピクセルのみを処理
	for y := minY; y <= maxY; y++ {
		if y < 0 || y >= targetSize.Dy() {
			continue
		}
		for x := minX; x <= maxX; x++ {
			if x < 0 || x >= targetSize.Dx() {
				continue
			}

			// 元画像の座標を計算
			sx, sy := applyInversePerspective(matrix, float64(x), float64(y))

			// 元画像の範囲内かチェック
			if sx >= float64(bounds.Min.X) && sx < float64(bounds.Max.X) &&
				sy >= float64(bounds.Min.Y) && sy < float64(bounds.Max.Y) {
				// バイリニア補間で色を取得
				color := bilinearInterpolation(srcNRGBA, sx, sy)
				dst.Set(x, y, color)
			}
		}
	}

	return dst
}

// calculatePerspectiveMatrix は射影変換行列を計算する
func calculatePerspectiveMatrix(src, dst []image.Point) [9]float64 {
	// 8つの方程式を解くための行列を作成
	a := make([]float64, 64)
	b := make([]float64, 8)

	for i := 0; i < 4; i++ {
		x := float64(src[i].X)
		y := float64(src[i].Y)
		u := float64(dst[i].X)
		v := float64(dst[i].Y)

		// x'の方程式
		a[i*8+0] = x
		a[i*8+1] = y
		a[i*8+2] = 1
		a[i*8+3] = 0
		a[i*8+4] = 0
		a[i*8+5] = 0
		a[i*8+6] = -x * u
		a[i*8+7] = -y * u
		b[i] = u

		// y'の方程式
		a[(i+4)*8+0] = 0
		a[(i+4)*8+1] = 0
		a[(i+4)*8+2] = 0
		a[(i+4)*8+3] = x
		a[(i+4)*8+4] = y
		a[(i+4)*8+5] = 1
		a[(i+4)*8+6] = -x * v
		a[(i+4)*8+7] = -y * v
		b[i+4] = v
	}

	// ガウス・ジョルダン法で連立方程式を解く
	x := solveLinearSystem(a, b)

	// 射影変換行列を返す（逆行列なので順序を変更）
	return [9]float64{x[0], x[1], x[2], x[3], x[4], x[5], x[6], x[7], 1}
}

// solveLinearSystem はガウス・ジョルダン法で連立方程式を解く
func solveLinearSystem(a []float64, b []float64) []float64 {
	n := 8
	for i := 0; i < n; i++ {
		// ピボット選択
		maxEl := math.Abs(a[i*n+i])
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(a[k*n+i]) > maxEl {
				maxEl = math.Abs(a[k*n+i])
				maxRow = k
			}
		}

		// 行の交換
		if maxRow != i {
			for j := 0; j < n; j++ {
				a[i*n+j], a[maxRow*n+j] = a[maxRow*n+j], a[i*n+j]
			}
			b[i], b[maxRow] = b[maxRow], b[i]
		}

		// 前進消去
		for k := i + 1; k < n; k++ {
			c := -a[k*n+i] / a[i*n+i]
			for j := i; j < n; j++ {
				if i == j {
					a[k*n+j] = 0
				} else {
					a[k*n+j] += c * a[i*n+j]
				}
			}
			b[k] += c * b[i]
		}
	}

	// 後退代入
	x := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		x[i] = b[i]
		for j := i + 1; j < n; j++ {
			x[i] -= a[i*n+j] * x[j]
		}
		x[i] /= a[i*n+i]
	}

	return x
}

// applyInversePerspective は逆射影変換を適用する
func applyInversePerspective(m [9]float64, x, y float64) (float64, float64) {
	w := m[6]*x + m[7]*y + m[8]
	if w == 0 {
		return 0, 0
	}
	sx := (m[0]*x + m[1]*y + m[2]) / w
	sy := (m[3]*x + m[4]*y + m[5]) / w
	return sx, sy
}

// bilinearInterpolation はバイリニア補間で色を取得する
func bilinearInterpolation(img *image.NRGBA, x, y float64) color.Color {
	x1 := int(math.Floor(x))
	x2 := x1 + 1
	y1 := int(math.Floor(y))
	y2 := y1 + 1

	if x1 < 0 || x2 >= img.Bounds().Dx() || y1 < 0 || y2 >= img.Bounds().Dy() {
		return color.Transparent
	}

	dx := x - float64(x1)
	dy := y - float64(y1)

	c11 := img.NRGBAAt(x1, y1)
	c12 := img.NRGBAAt(x1, y2)
	c21 := img.NRGBAAt(x2, y1)
	c22 := img.NRGBAAt(x2, y2)

	r := bilinearValue(c11.R, c12.R, c21.R, c22.R, dx, dy)
	g := bilinearValue(c11.G, c12.G, c21.G, c22.G, dx, dy)
	b := bilinearValue(c11.B, c12.B, c21.B, c22.B, dx, dy)
	a := bilinearValue(c11.A, c12.A, c21.A, c22.A, dx, dy)

	return color.NRGBA{R: r, G: g, B: b, A: a}
}

// bilinearValue は指定された4つの値からバイリニア補間した値を計算する
func bilinearValue(c11, c12, c21, c22 uint8, dx, dy float64) uint8 {
	fx1 := float64(c11)*(1-dx)*(1-dy) + float64(c21)*dx*(1-dy)
	fx2 := float64(c12)*(1-dx)*dy + float64(c22)*dx*dy
	return uint8(fx1 + fx2)
}

// applyMask はマスクを適用する
func applyMask(img, mask image.Image) *image.NRGBA {
	imgBounds := img.Bounds()
	maskBounds := mask.Bounds()

	// 出力画像を元画像のクローンとして作成
	dst := imaging.Clone(img)

	// マスクのサイズが異なる場合はリサイズ
	if imgBounds != maskBounds {
		mask = imaging.Resize(mask, imgBounds.Dx(), imgBounds.Dy(), imaging.Lanczos)
	}

	// 各ピクセルに対してマスクのアルファ値を適用
	for y := 0; y < imgBounds.Dy(); y++ {
		for x := 0; x < imgBounds.Dx(); x++ {
			// マスクのアルファ値を取得
			_, _, _, maskAlpha := mask.At(x, y).RGBA()
			maskAlpha = maskAlpha >> 8 // 16bit から 8bit に変換

			// 元画像のピクセルの色情報を保持したまま、マスクのアルファ値を適用
			c := dst.NRGBAAt(x, y)
			dst.Set(x, y, color.NRGBA{
				R: c.R,
				G: c.G,
				B: c.B,
				A: uint8(maskAlpha),
			})
		}
	}

	return dst
}
