package main

import "docker-go/inits"

// @title                       NAV Docker Gateway
// @version                     v0.1.0
// @description                 AI Docker gateway backend based on nav-common-go-lib.
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @BasePath                    /
func main() {
	inits.Init()
}
