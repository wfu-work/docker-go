package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/wfu-work/nav-common-go-lib/middlewares"
)

type RegistryCredentialRouter struct{}

func (r *RegistryCredentialRouter) InitRegistryCredentialRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("registry-credentials")
	logger := privateGroup.Group("registry-credentials").Use(middlewares.ApiLogger())
	{
		group.GET("list", registryCredentialApi.List)
		group.GET(":guid", registryCredentialApi.Get)

		logger.POST("", registryCredentialApi.Create)
		logger.PUT(":guid", registryCredentialApi.Update)
		logger.DELETE(":guid", registryCredentialApi.Delete)
	}
}

type DeployTemplateRouter struct{}

func (r *DeployTemplateRouter) InitDeployTemplateRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("deploy-templates")
	logger := privateGroup.Group("deploy-templates").Use(middlewares.ApiLogger())
	{
		group.GET("list", deployTemplateApi.List)
		group.GET(":guid", deployTemplateApi.Get)

		logger.POST("", deployTemplateApi.Create)
		logger.PUT(":guid", deployTemplateApi.Update)
		logger.DELETE(":guid", deployTemplateApi.Delete)
		logger.POST(":guid/render", deployTemplateApi.Render)
		logger.POST(":guid/create-config", deployTemplateApi.CreateConfig)
	}
}

type OperationRouter struct{}

func (r *OperationRouter) InitOperationRouter(privateGroup *gin.RouterGroup) {
	policies := privateGroup.Group("operation-policies")
	policyLogger := privateGroup.Group("operation-policies").Use(middlewares.ApiLogger())
	{
		policies.GET("list", operationApi.ListPolicies)
		policies.GET(":guid", operationApi.GetPolicy)

		policyLogger.POST("", operationApi.CreatePolicy)
		policyLogger.PUT(":guid", operationApi.UpdatePolicy)
		policyLogger.DELETE(":guid", operationApi.DeletePolicy)
	}

	approvals := privateGroup.Group("operation-approvals")
	approvalLogger := privateGroup.Group("operation-approvals").Use(middlewares.ApiLogger())
	{
		approvals.GET("list", operationApi.ListApprovals)
		approvals.GET(":guid", operationApi.GetApproval)

		approvalLogger.POST("", operationApi.CreateApproval)
		approvalLogger.POST(":guid/approve", operationApi.Approve)
		approvalLogger.POST(":guid/reject", operationApi.Reject)
		approvalLogger.POST(":guid/cancel", operationApi.Cancel)
	}

	audits := privateGroup.Group("operation-audits")
	{
		audits.GET("list", operationApi.ListAudits)
		audits.GET(":guid", operationApi.GetAudit)
	}
}
