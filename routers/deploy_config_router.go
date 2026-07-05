package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/middlewares"
)

type DeployConfigRouter struct{}

func (r *DeployConfigRouter) InitDeployConfigRouter(privateGroup *gin.RouterGroup) {
	configs := privateGroup.Group("deploy-configs")
	configLogger := privateGroup.Group("deploy-configs").Use(middlewares.ApiLogger())
	{
		configs.GET("list", deployConfigApi.ListDeployConfigs)
		configs.GET(":guid", deployConfigApi.GetDeployConfig)
		configs.GET(":guid/versions", deployConfigApi.ListDeployConfigVersions)
		configs.GET(":guid/versions/:versionGuid", deployConfigApi.GetDeployConfigVersion)

		configLogger.POST("", deployConfigApi.CreateDeployConfig)
		configLogger.PUT(":guid", deployConfigApi.UpdateDeployConfig)
		configLogger.DELETE(":guid", deployConfigApi.DeleteDeployConfig)
		configLogger.POST(":guid/versions", deployConfigApi.CreateDeployConfigVersion)
		configLogger.POST(":guid/validate", deployConfigApi.ValidateDeployConfig)
		configLogger.POST(":guid/deploy", deployConfigApi.DeployConfig)
		configLogger.POST(":guid/rollback", deployConfigApi.RollbackDeployConfig)
	}

	releases := privateGroup.Group("deploy-releases")
	{
		releases.GET("list", deployConfigApi.ListDeployReleases)
		releases.GET(":guid", deployConfigApi.GetDeployRelease)
	}
}
