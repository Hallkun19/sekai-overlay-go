package modules

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sekai-overlay-go/internal/config"
	"sekai-overlay-go/internal/utils"
)

// BpmChange はBPM変更を表す構造体
type BpmChange struct {
	Beat float64
	BPM  float64
}

// ScoreFrame はスコアフレームデータを表す構造体
type ScoreFrame struct {
	Seconds  float64 `json:"seconds"`
	Combo    int     `json:"combo"`
	Score    int     `json:"score"`
	AddScore int     `json:"add_score"`
	Rank     string  `json:"rank"`
	ScoreBar float64 `json:"score_bar"`
}

// SkobjData は出力データ構造体
type SkobjData struct {
	AssetPath string       `json:"asset_path"`
	Version   string       `json:"version"`
	Objects   []ScoreFrame `json:"objects"`
}

// getValueFromData はデータ配列から指定された名前の値を取得する
func getValueFromData(data []map[string]interface{}, name string) float64 {
	for _, item := range data {
		if item["name"] == name {
			if value, ok := item["value"].(float64); ok {
				return value
			}
		}
	}
	return 0.0
}

// getTimeFromBpmChanges はBPM変更リストから指定されたビート位置の時間を計算する
func getTimeFromBpmChanges(bpmChanges []BpmChange, beat float64) float64 {
	var retTime float64
	for i, bpmChange := range bpmChanges {
		if i == len(bpmChanges)-1 {
			retTime += (beat - bpmChange.Beat) * (60 / bpmChange.BPM)
			break
		}
		nextBpmChange := bpmChanges[i+1]
		if bpmChange.Beat <= beat && beat < nextBpmChange.Beat {
			retTime += (beat - bpmChange.Beat) * (60 / bpmChange.BPM)
			break
		} else if beat >= nextBpmChange.Beat {
			retTime += (nextBpmChange.Beat - bpmChange.Beat) * (60 / bpmChange.BPM)
		} else {
			break
		}
	}
	return retTime
}

// calculateScoreFrames はスコア、コンボ、秒数、ランク、スコアバーのフレームリストを計算する
func calculateScoreFrames(levelInfo map[string]interface{}, levelData map[string]interface{}, power float64) ([]ScoreFrame, float64) {
	rating, _ := levelInfo["rating"].(float64)
	entities, _ := levelData["entities"].([]interface{})

	// レーティングを5-40の範囲にクランプ
	clampedRating := math.Max(5, math.Min(rating, 40))

	// ランク境界を計算
	rankBorder := 1200000 + (clampedRating-5)*4100
	rankS := 1040000 + (clampedRating-5)*5200
	rankA := 840000 + (clampedRating-5)*4200
	rankB := 400000 + (clampedRating-5)*2000
	rankC := 20000 + (clampedRating-5)*100

	// スコアバーの位置定数
	const (
		posBorder = 1.0
		posS      = 0.890
		posA      = 0.742
		posB      = 0.591
		posC      = 0.447
	)

	// 重み付けされたノーツ数を計算
	var weightedNotesCount float64
	for _, entity := range entities {
		if entityMap, ok := entity.(map[string]interface{}); ok {
			if archetype, ok := entityMap["archetype"].(string); ok {
				if weight, exists := config.WeightMap[archetype]; exists {
					weightedNotesCount += weight
				}
			}
		}
	}

	if weightedNotesCount == 0 {
		return []ScoreFrame{{Seconds: 0, Combo: 0, Score: 0, AddScore: 0, Rank: "d", ScoreBar: 0}}, 0
	}

	var bpmChanges []BpmChange
	var noteEntities []map[string]interface{}

	// エンティティを分類
	for _, entity := range entities {
		if entityMap, ok := entity.(map[string]interface{}); ok {
			archetype, _ := entityMap["archetype"].(string)
			if archetype == "#BPM_CHANGE" {
				if data, ok := entityMap["data"].([]interface{}); ok {
					var dataSlice []map[string]interface{}
					for _, d := range data {
						if dm, ok := d.(map[string]interface{}); ok {
							dataSlice = append(dataSlice, dm)
						}
					}
					beat := getValueFromData(dataSlice, "#BEAT")
					bpm := getValueFromData(dataSlice, "#BPM")
					if bpm > 0 {
						bpmChanges = append(bpmChanges, BpmChange{Beat: beat, BPM: bpm})
					}
				}
			} else if weight, exists := config.WeightMap[archetype]; exists && weight > 0 {
				if _, ok := entityMap["data"].([]interface{}); ok {
					noteEntities = append(noteEntities, entityMap)
				}
			}
		}
	}

	// BPM変更をソート
	sort.Slice(bpmChanges, func(i, j int) bool {
		return bpmChanges[i].Beat < bpmChanges[j].Beat
	})

	// ノーツエンティティをビート順にソート
	sort.Slice(noteEntities, func(i, j int) bool {
		var dataI, dataJ []map[string]interface{}
		if d, ok := noteEntities[i]["data"].([]interface{}); ok {
			for _, item := range d {
				if dm, ok := item.(map[string]interface{}); ok {
					dataI = append(dataI, dm)
				}
			}
		}
		if d, ok := noteEntities[j]["data"].([]interface{}); ok {
			for _, item := range d {
				if dm, ok := item.(map[string]interface{}); ok {
					dataJ = append(dataJ, dm)
				}
			}
		}
		beatI := getValueFromData(dataI, "#BEAT")
		beatJ := getValueFromData(dataJ, "#BEAT")
		return beatI < beatJ
	})

	frames := []ScoreFrame{{Seconds: 0, Combo: 0, Score: 0, AddScore: 0, Rank: "none", ScoreBar: 0}}
	levelFax := (rating-5)*0.005 + 1
	comboFax := 1.0
	score := 0.0
	var lastNoteTime float64

	for i, entity := range noteEntities {
		comboCounter := i + 1

		if comboCounter%100 == 1 && comboCounter > 1 {
			comboFax += 0.01
		}
		if comboFax > 1.1 {
			comboFax = 1.1
		}

		archetype, _ := entity["archetype"].(string)
		weight := config.WeightMap[archetype]

		var dataSlice []map[string]interface{}
		if data, ok := entity["data"].([]interface{}); ok {
			for _, d := range data {
				if dm, ok := d.(map[string]interface{}); ok {
					dataSlice = append(dataSlice, dm)
				}
			}
		}

		addScore := (power / weightedNotesCount) * 4 * weight * 1 * levelFax * comboFax * 1
		score += addScore

		beat := getValueFromData(dataSlice, "#BEAT")
		time := getTimeFromBpmChanges(bpmChanges, beat)
		lastNoteTime = time

		// ランクとスコアバーを計算
		rank := ""
		scoreBar := 0.0

		switch {
		case score >= rankBorder:
			rank = "s"
			scoreBar = posBorder
		case score >= rankS:
			rank = "s"
			scoreBar = ((score-rankS)/(rankBorder-rankS))*(posBorder-posS) + posS
		case score >= rankA:
			rank = "a"
			scoreBar = ((score-rankA)/(rankS-rankA))*(posS-posA) + posA
		case score >= rankB:
			rank = "b"
			scoreBar = ((score-rankB)/(rankA-rankB))*(posA-posB) + posB
		case score >= rankC:
			rank = "c"
			scoreBar = ((score-rankC)/(rankB-rankC))*(posB-posC) + posC
		case score == 0:
			rank = "none"
			scoreBar = 0.0
		default:
			rank = "d"
			if rankC > 0 {
				scoreBar = (score / rankC) * posC
			}
		}

		frames = append(frames, ScoreFrame{
			Seconds:  math.Round(time*1000000) / 1000000,
			Combo:    comboCounter,
			Score:    int(math.Round(score)),
			AddScore: int(math.Round(addScore)),
			Rank:     rank,
			ScoreBar: math.Round(scoreBar*1000000) / 1000000,
		})
	}

	return frames, lastNoteTime
}

// GenerateSkobjData は譜面データを読み込み、スコアオブジェクトデータを計算してJSONファイルに出力する
func GenerateSkobjData(levelID, distDir string, teamPower float64, appVersion string) (float64, error) {
	levelInfoPath := filepath.Join(distDir, "level.json")
	chartPath := filepath.Join(distDir, "chart.json")

	// ファイルを読み込み
	levelInfoFile, err := os.Open(levelInfoPath)
	if err != nil {
		return 0, fmt.Errorf("level.jsonの読み込みに失敗しました: %w", err)
	}
	defer levelInfoFile.Close()

	var levelInfoData map[string]interface{}
	if err := json.NewDecoder(levelInfoFile).Decode(&levelInfoData); err != nil {
		return 0, fmt.Errorf("level.jsonの解析に失敗しました: %w", err)
	}

	levelInfo, ok := levelInfoData["item"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("level.jsonにitemフィールドが見つかりません")
	}

	chartFile, err := os.Open(chartPath)
	if err != nil {
		return 0, fmt.Errorf("chart.jsonの読み込みに失敗しました: %w", err)
	}
	defer chartFile.Close()

	var levelData map[string]interface{}
	if err := json.NewDecoder(chartFile).Decode(&levelData); err != nil {
		return 0, fmt.Errorf("chart.jsonの解析に失敗しました: %w", err)
	}

	fmt.Println("スコアオブジェクトデータの生成を開始します...")

	scoreFrames, lastNoteTime := calculateScoreFrames(levelInfo, levelData, teamPower)
	assetsFullPath := strings.Replace(filepath.ToSlash(filepath.Join(utils.GetAppRoot(), "assets")), "/", "\\", -1) + "\\"

	outputData := SkobjData{
		AssetPath: assetsFullPath,
		Version:   appVersion,
		Objects:   scoreFrames,
	}

	outputPath := filepath.Join(distDir, "skobj_data.json")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return 0, fmt.Errorf("出力ファイルの作成に失敗しました: %w", err)
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(outputData); err != nil {
		return 0, fmt.Errorf("JSON出力に失敗しました: %w", err)
	}

	fmt.Printf("スコアオブジェクトデータを '%s' に保存しました。\n", outputPath)
	return lastNoteTime, nil
}
