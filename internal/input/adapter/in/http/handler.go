package http

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"jiaa-server-core/internal/input/domain"
	portin "jiaa-server-core/internal/input/port/in"
)

// ActivityHandler HTTP 요청을 처리하는 Driving Adapter
// 클라이언트로부터 Activity 데이터를 수신하여 ReflexUseCase에 전달
type ActivityHandler struct {
	reflexUseCase portin.ReflexUseCase
}

// NewActivityHandler ActivityHandler 생성자
func NewActivityHandler(reflexUseCase portin.ReflexUseCase) *ActivityHandler {
	return &ActivityHandler{
		reflexUseCase: reflexUseCase,
	}
}

// ActivityRequest HTTP 요청 본문 구조체
type ActivityRequest struct {
	ClientID     string            `json:"client_id"`
	URL          string            `json:"url,omitempty"`
	AppName      string            `json:"app_name,omitempty"`
	ActivityType string            `json:"activity_type"`
	Timestamp    int64             `json:"timestamp,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ActivityResponse HTTP 응답 구조체
type ActivityResponse struct {
	Status     string `json:"status"`
	Blocked    bool   `json:"blocked"`
	ActionType string `json:"action_type,omitempty"`
	Message    string `json:"message,omitempty"`
}

// HandleActivity 클라이언트 Activity 수신 핸들러
// POST /api/v1/activity
func (h *ActivityHandler) HandleActivity(c echo.Context) error {
	var req ActivityRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validation
	if req.ClientID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "client_id is required",
		})
	}

	// Convert to domain entity
	activity := h.toClientActivity(req)

	// Process through ReflexUseCase
	action, err := h.reflexUseCase.ProcessActivity(activity)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Build response
	response := ActivityResponse{
		Status:  "ok",
		Blocked: action != nil,
	}

	if action != nil {
		response.ActionType = string(action.ActionType)
		response.Message = action.Message
	}

	return c.JSON(http.StatusOK, response)
}

// toClientActivity DTO를 Domain 엔티티로 변환
func (h *ActivityHandler) toClientActivity(req ActivityRequest) domain.ClientActivity {
	activity := domain.ClientActivity{
		ClientID:     req.ClientID,
		URL:          req.URL,
		AppName:      req.AppName,
		ActivityType: domain.ActivityType(req.ActivityType),
		Metadata:     req.Metadata,
	}

	if req.Timestamp > 0 {
		activity.Timestamp = time.UnixMilli(req.Timestamp)
	} else {
		activity.Timestamp = time.Now()
	}

	if activity.Metadata == nil {
		activity.Metadata = make(map[string]string)
	}

	return activity
}

// RegisterRoutes Echo 라우터에 핸들러 등록
func (h *ActivityHandler) RegisterRoutes(e *echo.Echo) {
	api := e.Group("/api/v1")
	api.POST("/activity", h.HandleActivity)
}
