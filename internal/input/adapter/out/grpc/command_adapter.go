package grpc

import (
	"context"

	"jiaa-server-core/internal/input/domain"
	"jiaa-server-core/pkg/proto"
)

// commandAdapter SabotageCommandService gRPC 클라이언트 (Driven Adapter)
// 기존 proto 정의와 호환성 유지
type commandAdapter struct {
	client proto.SabotageCommandServiceClient
}

// NewCommandAdapter commandAdapter 생성자
func NewCommandAdapter(client proto.SabotageCommandServiceClient) *commandAdapter {
	return &commandAdapter{client: client}
}

// SendSabotage 사보타주 명령 전송 (CommandPort 구현)
func (a *commandAdapter) SendSabotage(cmd domain.SabotageAction) error {
	_, err := a.client.ExecuteSabotage(context.Background(), &proto.SabotageRequest{
		ClientId:   cmd.ClientID,
		ActionType: string(cmd.ActionType),
		Intensity:  int32(cmd.Intensity),
		Message:    cmd.Message,
	})
	return err
}
