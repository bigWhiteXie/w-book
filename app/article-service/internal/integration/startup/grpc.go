package startup

import (
	"testing"

	"codexie.com/w-book-interact/api/grpc"
	"github.com/golang/mock/gomock"
)

func InitInteractClient(t *testing.T) *grpc.MockInteractionClient {
	ctrl := gomock.NewController(t)
	client := grpc.NewMockInteractionClient(ctrl)
	client.EXPECT().QueryInteractionInfo(gomock.Any(), gomock.Any()).Return(&grpc.InteractionResult{
		ReadCnt:     10,
		CollectCnt:  10,
		LikeCnt:     10,
		IsLiked:     true,
		IsCollected: true,
	}, nil).AnyTimes()
	return client
}
