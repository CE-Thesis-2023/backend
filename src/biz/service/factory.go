package service

import "sync"

var once sync.Once

var (
	commandService *CommandService
	webService     *WebService
)

func Init() {
	once.Do(func() {
		webService = NewWebService()
		commandService = NewCommandService()
	})
}

func GetCommandService() *CommandService {
	return commandService
}

func GetWebService() *WebService {
	return webService
}
