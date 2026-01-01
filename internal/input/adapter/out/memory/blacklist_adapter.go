package memory

// BlacklistAdapter 인메모리 블랙리스트 어댑터 (Driven Adapter)
// 테스트/개발용 - 프로덕션에서는 Redis 등으로 교체
type BlacklistAdapter struct {
	urlBlacklist map[string]bool
	appBlacklist map[string]bool
}

// NewBlacklistAdapter BlacklistAdapter 생성자
func NewBlacklistAdapter() *BlacklistAdapter {
	return &BlacklistAdapter{
		urlBlacklist: make(map[string]bool),
		appBlacklist: make(map[string]bool),
	}
}

// NewBlacklistAdapterWithDefaults 기본 블랙리스트가 포함된 생성자
func NewBlacklistAdapterWithDefaults() *BlacklistAdapter {
	adapter := NewBlacklistAdapter()

	// 기본 URL 블랙리스트 (예시)
	defaultURLs := []string{
		"youtube.com",
		"netflix.com",
		"twitch.tv",
		"instagram.com",
		"facebook.com",
		"twitter.com",
		"tiktok.com",
		"reddit.com",
	}
	for _, url := range defaultURLs {
		adapter.AddURLToBlacklist(url)
	}

	// 기본 앱 블랙리스트 (예시)
	defaultApps := []string{
		"League of Legends",
		"Steam",
		"Discord",
		"Slack",
		"KakaoTalk",
	}
	for _, app := range defaultApps {
		adapter.AddAppToBlacklist(app)
	}

	return adapter
}

// IsBlacklisted 주어진 URL이 블랙리스트에 있는지 확인
func (a *BlacklistAdapter) IsBlacklisted(url string) bool {
	// 정확한 매칭 체크
	if a.urlBlacklist[url] {
		return true
	}

	// 부분 매칭 체크 (URL에 블랙리스트 도메인이 포함되어 있는지)
	for blacklistedURL := range a.urlBlacklist {
		if containsDomain(url, blacklistedURL) {
			return true
		}
	}

	return false
}

// IsAppBlacklisted 주어진 앱이 블랙리스트에 있는지 확인
func (a *BlacklistAdapter) IsAppBlacklisted(appName string) bool {
	return a.appBlacklist[appName]
}

// AddURLToBlacklist URL을 블랙리스트에 추가
func (a *BlacklistAdapter) AddURLToBlacklist(url string) {
	a.urlBlacklist[url] = true
}

// RemoveURLFromBlacklist URL을 블랙리스트에서 제거
func (a *BlacklistAdapter) RemoveURLFromBlacklist(url string) {
	delete(a.urlBlacklist, url)
}

// AddAppToBlacklist 앱을 블랙리스트에 추가
func (a *BlacklistAdapter) AddAppToBlacklist(appName string) {
	a.appBlacklist[appName] = true
}

// RemoveAppFromBlacklist 앱을 블랙리스트에서 제거
func (a *BlacklistAdapter) RemoveAppFromBlacklist(appName string) {
	delete(a.appBlacklist, appName)
}

// containsDomain URL에 도메인이 포함되어 있는지 확인
func containsDomain(url, domain string) bool {
	// 간단한 문자열 포함 체크
	// 실제 프로덕션에서는 URL 파싱 후 정확한 도메인 비교 필요
	return len(url) >= len(domain) &&
		(url == domain ||
			contains(url, "."+domain) ||
			contains(url, "://"+domain) ||
			contains(url, "/"+domain))
}

// contains 문자열 포함 여부 확인 (strings.Contains 대체)
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
