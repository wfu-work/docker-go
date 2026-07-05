package domains

import (
	"os"

	"github.com/wfu-work/nav-common-go-lib/global"
	"go.uber.org/zap"
)

func RegisterTables() {
	db := global.NAV_DB
	if db == nil {
		return
	}
	if err := db.AutoMigrate(
		Setting{},
		Host{},
		Agent{},
		AgentUpgradePackage{},
		Task{},
		TaskEvent{},
		DockerContainer{},
		DockerImage{},
		HostMetric{},
		ContainerMetric{},
		DeployConfig{},
		DeployConfigVersion{},
		DeployRelease{},
		RegistryCredential{},
		DeployTemplate{},
		OperationApproval{},
		OperationPolicy{},
		OperationAudit{},
	); err != nil {
		global.NAV_LOG.Error("register nav docker business tables failed", zap.Error(err))
		os.Exit(1)
	}
	global.NAV_LOG.Info("register nav docker business tables success")
}
