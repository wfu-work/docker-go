package inits

import (
	"docker-go/domains"
	"docker-go/routers"
	"docker-go/utils"
	_ "embed"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/wfu-work/nav-common-go-lib/inits"
	commonScheduleds "github.com/wfu-work/nav-common-go-lib/scheduleds"
)

//go:embed config.yaml
var defaultConfig []byte

func Init() {
	if err := utils.NewDefaultConfigManager(defaultConfig).Ensure(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "prepare config failed: %v\n", err)
		os.Exit(1)
	}
	sysInit := inits.SysInit{}
	sysInit.OnTableInit(func() {
		domains.RegisterTables()
	})
	sysInit.OnRouterInit(func(publicGroup *gin.RouterGroup, privateGroup *gin.RouterGroup) {
		routers.RouterGroupApp.InitRouters(publicGroup, privateGroup)
	})
	sysInit.OnWebInit(func(engine *gin.Engine) {

	})
	sysInit.OnOtherInit(func() {

	})
	sysInit.OnScheInit(func(timers commonScheduleds.Timer, options []cron.Option) {

	})
	sysInit.OnClearInit(func() []commonScheduleds.ClearDB {
		return []commonScheduleds.ClearDB{}
	})
	sysInit.Init()
}
