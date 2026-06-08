package messages

import "context"

type RealtimeHub interface {
	BroadcastMessageReceived(ctx context.Context, msg *Message)
}
