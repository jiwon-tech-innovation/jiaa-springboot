package domain

import "time"

// ActivityType 클라이언트 활동 유형
type ActivityType string

const (
	ActivityURLVisit   ActivityType = "URL_VISIT"
	ActivityAppOpen    ActivityType = "APP_OPEN"
	ActivityAppClose   ActivityType = "APP_CLOSE"
	ActivityIdleStart  ActivityType = "IDLE_START"
	ActivityIdleEnd    ActivityType = "IDLE_END"
	ActivityInputUsage ActivityType = "INPUT_USAGE"
)

// ClientActivity 클라이언트의 활동 데이터를 나타내는 도메인 엔티티
// 클라이언트로부터 수신한 원시 활동 데이터
type ClientActivity struct {
	ClientID     string            // 클라이언트 식별자
	URL          string            // 방문 URL (URL_VISIT인 경우)
	AppName      string            // 앱 이름 (APP_OPEN/CLOSE인 경우)
	Timestamp    time.Time         // 활동 발생 시간
	ActivityType ActivityType      // 활동 유형
	Metadata     map[string]string // 추가 메타데이터
}

// NewClientActivity ClientActivity 생성자
func NewClientActivity(clientID string, activityType ActivityType) *ClientActivity {
	return &ClientActivity{
		ClientID:     clientID,
		ActivityType: activityType,
		Timestamp:    time.Now(),
		Metadata:     make(map[string]string),
	}
}

// WithURL URL 설정
func (c *ClientActivity) WithURL(url string) *ClientActivity {
	c.URL = url
	return c
}

// WithAppName 앱 이름 설정
func (c *ClientActivity) WithAppName(appName string) *ClientActivity {
	c.AppName = appName
	return c
}

// WithTimestamp 타임스탬프 설정
func (c *ClientActivity) WithTimestamp(timestamp time.Time) *ClientActivity {
	c.Timestamp = timestamp
	return c
}

// AddMetadata 메타데이터 추가
func (c *ClientActivity) AddMetadata(key, value string) *ClientActivity {
	c.Metadata[key] = value
	return c
}

// IsURLActivity URL 관련 활동인지 확인
func (c *ClientActivity) IsURLActivity() bool {
	return c.ActivityType == ActivityURLVisit && c.URL != ""
}

// IsAppActivity 앱 관련 활동인지 확인
func (c *ClientActivity) IsAppActivity() bool {
	return (c.ActivityType == ActivityAppOpen || c.ActivityType == ActivityAppClose) && c.AppName != ""
}
