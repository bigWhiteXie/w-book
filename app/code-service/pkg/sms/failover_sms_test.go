package sms

import (
	"context"
	"errors"
	"testing"
	"time"

	mock_sms "codexie.com/w-book-code/mocks/pkg/sms"
	"github.com/golang/mock/gomock"
	"github.com/zeromicro/go-zero/core/logx"
)

func TestSmsService_SendSms(t *testing.T) {
	type fields struct {
		smsProviders map[string]*SmsProvider
	}
	type args struct {
		ctx   context.Context
		phone string
		args  map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		mock    func(ctrl *gomock.Controller) *SmsService
		args    args
		wantErr bool
		num     int
	}{
		{
			name: "正常发送",
			mock: func(ctrl *gomock.Controller) *SmsService {
				c1 := mock_sms.NewMockSmsClient(ctrl)
				c1.EXPECT().SendSms(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

				s := &SmsService{
					smsProviders: map[string]*SmsProvider{
						"c1": {
							Name:      "c1",
							Client:    c1,
							Weight:    4,
							Status:    1,
							FailLimit: 3,
						},
					},
				}
				return s
			},
			args: args{
				ctx:   context.Background(),
				phone: "13800138000",
				args: map[string]string{
					"code": "123456",
				},
			},
			wantErr: false,
			num:     1,
		},
		{
			name: "失败切换服务商",
			mock: func(ctrl *gomock.Controller) *SmsService {
				c1 := mock_sms.NewMockSmsClient(ctrl)
				c1.EXPECT().SendSms(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error"))
				c2 := mock_sms.NewMockSmsClient(ctrl)
				c2.EXPECT().SendSms(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				s := &SmsService{
					smsProviders: map[string]*SmsProvider{
						"c1": {
							Name:      "c1",
							Client:    c1,
							Weight:    99,
							Status:    1,
							FailLimit: 3,
						},
						"c2": {
							Name:      "c2",
							Client:    c2,
							Weight:    1,
							Status:    1,
							FailLimit: 3,
						},
					},
				}
				return s
			},
			args: args{
				ctx:   context.Background(),
				phone: "13800138000",
				args: map[string]string{
					"code": "123456",
				},
			},
			wantErr: false,
			num:     2,
		},
		{
			name: "超时3次切换服务商",
			mock: func(ctrl *gomock.Controller) *SmsService {
				c1 := mock_sms.NewMockSmsClient(ctrl)
				c1.EXPECT().SendSms(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, phone string, args map[string]string) error {
					time.Sleep(4 * time.Second)
					return nil
				})
				c1.EXPECT().SendSms(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, phone string, args map[string]string) error {
					time.Sleep(4 * time.Second)
					return nil
				})
				c1.EXPECT().SendSms(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, phone string, args map[string]string) error {
					time.Sleep(4 * time.Second)
					return nil
				})
				c2 := mock_sms.NewMockSmsClient(ctrl)
				c2.EXPECT().SendSms(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				s := &SmsService{
					smsProviders: map[string]*SmsProvider{
						"c1": {
							Name:      "c1",
							Client:    c1,
							Weight:    99,
							Status:    1,
							FailLimit: 3,
						},
						"c2": {
							Name:      "c2",
							Client:    c2,
							Weight:    1,
							Status:    1,
							FailLimit: 3,
						},
					},
				}
				return s
			},
			args: args{
				ctx:   context.Background(),
				phone: "13800138000",
				args: map[string]string{
					"code": "123456",
				},
			},
			wantErr: false,
			num:     4,
		},
	}
	ctrl := gomock.NewController(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.mock(ctrl)
			var err error
			for i := range tt.num {
				logx.Infof("第%d次发送", i)
				err = s.SendSms(tt.args.ctx, tt.args.phone, tt.args.args)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("SmsService.SendSms() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
