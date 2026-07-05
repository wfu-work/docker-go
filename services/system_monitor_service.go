package services

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	commonServices "github.com/wfu-work/nav-common-go-lib/services"
	commonUtils "github.com/wfu-work/nav-common-go-lib/utils"
)

var (
	SystemMonitorServiceApp = new(SystemMonitorService)
	serverStartedAt         = time.Now()
)

type SystemMonitorService struct{}

type ServiceRuntimeInfo struct {
	Name            string `json:"name"`
	Status          string `json:"status"`
	PID             int    `json:"pid"`
	StartedAt       int64  `json:"startedAt"`
	UptimeSeconds   int64  `json:"uptimeSeconds"`
	WorkingDir      string `json:"workingDir"`
	Executable      string `json:"executable"`
	GoVersion       string `json:"goVersion"`
	GOOS            string `json:"goos"`
	Compiler        string `json:"compiler"`
	NumCPU          int    `json:"numCpu"`
	NumGoroutine    int    `json:"numGoroutine"`
	AllocBytes      uint64 `json:"allocBytes"`
	SysBytes        uint64 `json:"sysBytes"`
	HeapAllocBytes  uint64 `json:"heapAllocBytes"`
	HeapInuseBytes  uint64 `json:"heapInuseBytes"`
	LastGCPauseNano uint64 `json:"lastGcPauseNano"`
}

type SystemMonitorInfo struct {
	Service   ServiceRuntimeInfo `json:"service"`
	CPU       commonUtils.Cpu    `json:"cpu"`
	RAM       commonUtils.Ram    `json:"ram"`
	Disk      []commonUtils.Disk `json:"disk"`
	Warnings  []string           `json:"warnings"`
	CheckedAt int64              `json:"checkedAt"`
}

func (s SystemMonitorService) Runtime() *SystemMonitorInfo {
	server, err := commonServices.OsServiceApp.GetServerInfo()
	warnings := make([]string, 0)
	if err != nil {
		warnings = append(warnings, err.Error())
	}
	if server == nil {
		server = &commonUtils.Server{}
	}
	return &SystemMonitorInfo{
		Service:   currentServiceRuntime(),
		CPU:       server.Cpu,
		RAM:       server.Ram,
		Disk:      server.Disk,
		Warnings:  warnings,
		CheckedAt: time.Now().UnixMilli(),
	}
}

func currentServiceRuntime() ServiceRuntimeInfo {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	workingDir, _ := os.Getwd()
	executable, _ := os.Executable()
	name := filepath.Base(executable)
	if name == "." || name == string(filepath.Separator) || name == "" {
		name = "nav-docker-server"
	}

	var lastGCPause uint64
	if stats.NumGC > 0 {
		lastGCPause = stats.PauseNs[(stats.NumGC+255)%256]
	}

	return ServiceRuntimeInfo{
		Name:            name,
		Status:          "running",
		PID:             os.Getpid(),
		StartedAt:       serverStartedAt.UnixMilli(),
		UptimeSeconds:   int64(time.Since(serverStartedAt).Seconds()),
		WorkingDir:      workingDir,
		Executable:      executable,
		GoVersion:       runtime.Version(),
		GOOS:            runtime.GOOS,
		Compiler:        runtime.Compiler,
		NumCPU:          runtime.NumCPU(),
		NumGoroutine:    runtime.NumGoroutine(),
		AllocBytes:      stats.Alloc,
		SysBytes:        stats.Sys,
		HeapAllocBytes:  stats.HeapAlloc,
		HeapInuseBytes:  stats.HeapInuse,
		LastGCPauseNano: lastGCPause,
	}
}
