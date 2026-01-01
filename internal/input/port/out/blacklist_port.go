package out

// BlacklistPort URL 블랙리스트 조회를 위한 Driven Port
// 속도가 생명인 즉각 차단을 위해 사용
type BlacklistPort interface {
	// IsBlacklisted 주어진 URL이 블랙리스트에 있는지 확인
	IsBlacklisted(url string) bool

	// IsAppBlacklisted 주어진 앱이 블랙리스트에 있는지 확인
	IsAppBlacklisted(appName string) bool
}
