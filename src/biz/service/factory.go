package service

import "sync"

var once sync.Once

var (
	commandService *PrivateService
	webService     *WebService
)

func Init() {
	once.Do(func() {
		webService = NewWebService()
		commandService = NewPrivateService()
	})
}

func GetPrivateService() *PrivateService {
	return commandService
}

func GetWebService() *WebService {
	return webService
}
