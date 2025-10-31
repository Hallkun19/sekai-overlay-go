package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppVersion     = "0.1.0"
	UpdateCheckURL = "https://raw.githubusercontent.com/Hallkun19/SekaiOverlay/refs/heads/main/data.json"
	ReleasePageURL = "https://github.com/Hallkun19/SekaiOverlay/releases/latest"
)

// Config はアプリケーション設定を保持する構造体
type Config struct {
	FullLevelID string                 `json:"full_level_id"`
	BgVersion   string                 `json:"bg_version"`
	TeamPower   float64                `json:"team_power"`
	AppVersion  string                 `json:"app_version"`
	ExtraData   map[string]interface{} `json:"extra_data"`
}

// GetConfigDir は設定ディレクトリのパスを取得する
func GetConfigDir() string {
	var configDir string
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		configDir = filepath.Join(appData, "SekaiOverlay")
	} else {
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".sekai-overlay")
	}
	return configDir
}

// GetConfigPath は設定ファイルのパスを取得する
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.ini")
}

// AviUtlScriptDir はAviUtlスクリプトディレクトリのパス
var AviUtlScriptDir = `C:\ProgramData\aviutl2\Script`

// UnmultAnmURL はunmult.anm2のダウンロードURL
var UnmultAnmURL = "https://gist.githubusercontent.com/mes51/f90331af552231f39adb5ed3847ebe86/raw/121c5a97d7d776270bdb81febdcf12e79b257466/unmult.anm2"

// DkjsonLuaURL はdkjson.luaのダウンロードURL
var DkjsonLuaURL = "https://raw.githubusercontent.com/LuaDist/dkjson/refs/heads/master/dkjson.lua"

// ServerMap はサーバーURLマッピング
var ServerMap = map[string]string{
	"chcy":               "https://cc.sevenc7c.com/sonolus/levels/",
	"ptlv":               "https://ptlv.sevenc7c.com/sonolus/levels/",
	"UnCh":               "https://untitledcharts.com/sonolus/levels/",
	"coconut-next-sekai": "https://coconut.sonolus.com/next-sekai/levels/",
}

// WeightMap はノーツの重み付けマップ
var WeightMap = map[string]float64{
	// CC
	"#BPM_CHANGE": 0, "Initialization": 0, "InputManager": 0, "Stage": 0,
	"NormalTapNote": 1, "CriticalTapNote": 2, "NormalFlickNote": 1, "CriticalFlickNote": 3,
	"NormalSlideStartNote": 1, "CriticalSlideStartNote": 2, "NormalSlideEndNote": 1, "CriticalSlideEndNote": 2,
	"NormalSlideEndFlickNote": 1, "CriticalSlideEndFlickNote": 3, "HiddenSlideTickNote": 0,
	"NormalSlideTickNote": 0.1, "CriticalSlideTickNote": 0.2, "IgnoredSlideTickNote": 0.1,
	"NormalAttachedSlideTickNote": 0.1, "CriticalAttachedSlideTickNote": 0.2,
	"NormalSlideConnector": 0, "CriticalSlideConnector": 0, "SimLine": 0,
	"NormalSlotEffect": 0, "SlideSlotEffect": 0, "FlickSlotEffect": 0, "CriticalSlotEffect": 0,
	"NormalSlotGlowEffect": 0, "SlideSlotGlowEffect": 0, "FlickSlotGlowEffect": 0, "CriticalSlotGlowEffect": 0,
	"NormalTraceNote": 0.1, "CriticalTraceNote": 0.2, "NormalTraceSlotEffect": 0, "NormalTraceSlotGlowEffect": 0,
	"DamageNote": 0.1, "DamageSlotEffect": 0, "DamageSlotGlowEffect": 0,
	"NormalTraceFlickNote": 1, "CriticalTraceFlickNote": 3, "NonDirectionalTraceFlickNote": 1,
	"NormalTraceSlideStartNote": 0.1, "NormalTraceSlideEndNote": 0.1,
	"CriticalTraceSlideStartNote": 0.2, "CriticalTraceSlideEndNote": 0.2,
	"TimeScaleGroup": 0, "TimeScaleChange": 0,

	// NS
	"#TIMESCALE_CHANGE": 0, "#TIMESCALE_GROUP": 0, "_InputManager": 0, "SlideManager": 0,
	"Connector": 0, "SlotGlowEffect": 0, "SlotEffect": 0, "NormalHeadTapNote": 1,
	"CriticalHeadTapNote": 2, "NormalHeadFlickNote": 1, "CriticalHeadFlickNote": 3,
	"NormalHeadTraceNote": 0.1, "CriticalHeadTraceNote": 0.2, "NormalHeadTraceFlickNote": 1,
	"CriticalHeadTraceFlickNote": 3, "NormalHeadReleaseNote": 1, "CriticalHeadReleaseNote": 2,
	"NormalTailTapNote": 1, "CriticalTailTapNote": 2, "NormalTailFlickNote": 1,
	"CriticalTailFlickNote": 3, "NormalTailTraceNote": 0.1, "CriticalTailTraceNote": 0.2,
	"NormalTailTraceFlickNote": 1, "CriticalTailTraceFlickNote": 3,
	"NormalTailReleaseNote": 1, "CriticalTailReleaseNote": 2, "TransientHiddenTickNote": 0.1,
	"NormalTickNote": 0.1, "CriticalTickNote": 0.2, "AnchorNote": 0,
	"FakeNormalTapNote": 0, "FakeCriticalTapNote": 0, "FakeNormalFlickNote": 0,
	"FakeCriticalFlickNote": 0, "FakeNormalTraceNote": 0, "FakeCriticalTraceNote": 0,
	"FakeNormalTraceFlickNote": 0, "FakeCriticalTraceFlickNote": 0,
	"FakeNormalReleaseNote": 0, "FakeCriticalReleaseNote": 0,
	"FakeNormalHeadTapNote": 0, "FakeCriticalHeadTapNote": 0, "FakeNormalHeadFlickNote": 0,
	"FakeCriticalHeadFlickNote": 0, "FakeNormalHeadTraceNote": 0, "FakeCriticalHeadTraceNote": 0,
	"FakeNormalHeadTraceFlickNote": 0, "FakeCriticalHeadTraceFlickNote": 0,
	"FakeNormalHeadReleaseNote": 0, "FakeCriticalHeadReleaseNote": 0,
	"FakeNormalTailTapNote": 0, "FakeCriticalTailTapNote": 0, "FakeNormalTailFlickNote": 0,
	"FakeCriticalTailFlickNote": 0, "FakeNormalTailTraceNote": 0, "FakeCriticalTailTraceNote": 0,
	"FakeNormalTailTraceFlickNote": 0, "FakeCriticalTailTraceFlickNote": 0,
	"FakeNormalTailReleaseNote": 0, "FakeCriticalTailReleaseNote": 0,
	"FakeTransientHiddenTickNote": 0, "FakeNormalTickNote": 0, "FakeCriticalTickNote": 0,
	"FakeAnchorNote": 0, "FakeDamageNote": 0,
}
