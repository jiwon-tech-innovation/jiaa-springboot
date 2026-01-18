package service

import (
	"math"
)

// ScoreService 점수 산정 로직 (JIAA Algorithm)
type ScoreService struct{}

func NewScoreService() *ScoreService {
	return &ScoreService{}
}

// CalculateInput 점수 계산에 필요한 입력 데이터
type CalculateInput struct {
	EyesClosedDurationSec float64 // 눈 감은 지속 시간 (초)
	HeadPitch             float64 // 고개 각도 (Pitch)
	URLCategory           string  // URL 카테고리 (PLAY/STUDY/WORK/NEUTRAL)
	OSActivityCount       int     // 키보드+마우스 입력 횟수
	VisionScore           int     // 시선 집중도 (0-100)
	CurrentScore          int     // 현재 점수
}

// CalculateResult 계산 결과
type CalculateResult struct {
	FinalScore int
	State      string
}

// CalculateScore 사용자의 상태와 점수를 계산
func (s *ScoreService) CalculateScore(input CalculateInput) CalculateResult {
	// 1단계: "즉결 처형" (Sanctions) - 최우선 순위

	// 졸음 감지 (Sleep)
	// 조건: eyes_closed == true (3초 이상 지속) OR head_pitch < -20 (고개 숙임)
	isSleeping := input.EyesClosedDurationSec >= 3.0 || input.HeadPitch < -20.0
	if isSleeping {
		return CalculateResult{FinalScore: 0, State: "SLEEPING"}
	}

	// 유해 사이트 (Banned URL)
	// 조건: url_category == "PLAY" (게임, 유튜브 등)
	if input.URLCategory == "PLAY" {
		return CalculateResult{FinalScore: 10, State: "DISTRACTED"}
	}

	// 2단계: "생각 모드 보호" (Thinking Mode Protection)
	// 조건: os_activity == 0 (입력 없음), vision_score > 70 (화면은 잘 봄), head_pitch > -10 (고개 듬)
	if input.OSActivityCount == 0 && input.VisionScore > 70 && input.HeadPitch > -10.0 {
		// 점수: 90점 고정 (혹은 MAX(현재점수, 85)) -> 요청대로 90점 고정 로직 적용하되, 기존 점수가 더 높으면 유지
		newScore := 90
		if input.CurrentScore > 90 {
			newScore = input.CurrentScore
		}
		return CalculateResult{FinalScore: newScore, State: "THINKING"}
	}

	// 3단계: "일반 집중 모드" (Active Focus)
	// 조건: os_activity > 0 (입력 있음)
	if input.OSActivityCount > 0 {
		// OS_Norm: 마우스/키보드 횟수를 0~100으로 정규화 (예: 1초에 5타 이상이면 100점)
		// 입력 input.OSActivityCount가 1초 기준이라고 가정
		osNorm := float64(input.OSActivityCount) * 20.0
		if osNorm > 100.0 {
			osNorm = 100.0
		}

		// Final = (Vision * 0.6) + (OS_Norm * 0.4)
		weighted := (float64(input.VisionScore) * 0.6) + (osNorm * 0.4)
		finalScore := int(math.Round(weighted))

		return CalculateResult{FinalScore: finalScore, State: "FOCUSING"}
	}

	// 4단계: "멍 때리기" (Idling)
	// 조건: os_activity == 0 AND vision_score < 50
	if input.OSActivityCount == 0 && input.VisionScore < 50 {
		// 점수: 감점 (Decay) - 1초마다 5점씩 깎음
		newScore := input.CurrentScore - 5
		if newScore < 0 {
			newScore = 0
		}
		return CalculateResult{FinalScore: newScore, State: "IDLING"}
	}

	// 그 외 (Gray Area: 입력 없고 50 <= 시선 <= 70)
	// 현상 유지 또는 완만한 감점? -> 일단 현상 유지 (NEUTRAL)
	return CalculateResult{FinalScore: input.CurrentScore, State: "NEUTRAL"}
}
