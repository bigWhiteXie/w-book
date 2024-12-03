package ioc

import (
	"codexie.com/w-book-code/internal/event"
	"codexie.com/w-book-code/internal/logic"
	"codexie.com/w-book-code/internal/server"
	"codexie.com/w-book-code/internal/svc"
)

type App struct {
	Server         *server.SMSServer
	SmsEvtListener *event.SmsEvtListener
}

func NewSmsApp(svc *svc.ServiceContext, codeLogic *logic.CodeLogic, smsListener *event.SmsEvtListener) *App {
	smsServer := server.NewSMSServer(svc, codeLogic)
	return &App{
		Server:         smsServer,
		SmsEvtListener: smsListener,
	}
}
