package out

import "jiaa-server-core/internal/input/domain"

type CommandPort interface {
	SendSabotage(cmd domain.SabotageAction) error
}
