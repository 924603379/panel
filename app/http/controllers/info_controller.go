package controllers

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"panel/app/models"
	"panel/app/services"
	"panel/pkg/tools"
)

type MenuItem struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Icon  string `json:"icon"`
	Jump  string `json:"jump"`
}

type InfoController struct {
	plugin  services.Plugin
	setting services.Setting
}

func NewInfoController() *InfoController {
	return &InfoController{
		plugin:  services.NewPluginImpl(),
		setting: services.NewSettingImpl(),
	}
}

// Name 获取面板名称
func (r *InfoController) Name(ctx http.Context) http.Response {
	return Success(ctx, http.Json{
		"name": r.setting.Get(models.SettingKeyName),
	})
}

// HomePlugins 获取首页插件
func (r *InfoController) HomePlugins(ctx http.Context) http.Response {
	var plugins []models.Plugin
	err := facades.Orm().Query().Where("show", 1).Find(&plugins)
	if err != nil {
		facades.Log().Request(ctx.Request()).Tags("面板", "基础信息").With(map[string]any{
			"error": err.Error(),
		}).Info("获取首页插件失败")
		return ErrorSystem(ctx)
	}

	type pluginsData struct {
		models.Plugin
		Name string `json:"name"`
	}

	var pluginsJson []pluginsData
	for _, plugin := range plugins {
		pluginsJson = append(pluginsJson, pluginsData{
			Plugin: plugin,
			Name:   services.NewPluginImpl().GetBySlug(plugin.Slug).Name,
		})
	}

	return Success(ctx, pluginsJson)
}

// NowMonitor 获取当前监控信息
func (r *InfoController) NowMonitor(ctx http.Context) http.Response {
	return Success(ctx, tools.GetMonitoringInfo())
}

// SystemInfo 获取系统信息
func (r *InfoController) SystemInfo(ctx http.Context) http.Response {
	monitorInfo := tools.GetMonitoringInfo()

	return Success(ctx, http.Json{
		"os_name":       monitorInfo.Host.Platform + " " + monitorInfo.Host.PlatformVersion,
		"uptime":        fmt.Sprintf("%.2f", float64(monitorInfo.Host.Uptime)/86400),
		"panel_version": facades.Config().GetString("panel.version"),
	})
}

// CountInfo 获取面板统计信息
func (r *InfoController) CountInfo(ctx http.Context) http.Response {
	var websiteCount int64
	err := facades.Orm().Query().Model(models.Website{}).Count(&websiteCount)
	if err != nil {
		websiteCount = -1
	}

	var mysql models.Plugin
	mysqlInstalled := true
	err = facades.Orm().Query().Where("slug like ?", "mysql%").FirstOrFail(&mysql)
	if err != nil {
		mysqlInstalled = false
	}
	var postgresql models.Plugin
	postgresqlInstalled := true
	err = facades.Orm().Query().Where("slug like ?", "postgresql%").FirstOrFail(&postgresql)
	if err != nil {
		postgresqlInstalled = false
	}
	var databaseCount int64
	if mysqlInstalled {
		status, err := tools.Exec("systemctl status mysqld | grep Active | grep -v grep | awk '{print $2}'")
		if status == "active" && err == nil {
			rootPassword := r.setting.Get(models.SettingKeyMysqlRootPassword)
			type database struct {
				Name string `json:"name"`
			}

			db, err := sql.Open("mysql", "root:"+rootPassword+"@unix(/tmp/mysql.sock)/")
			if err != nil {
				facades.Log().Request(ctx.Request()).Tags("面板", "基础信息").With(map[string]any{
					"error": err.Error(),
				}).Info("获取数据库列表失败")
				databaseCount = -1
			} else {
				defer db.Close()
				rows, err := db.Query("SHOW DATABASES")
				if err != nil {
					facades.Log().Request(ctx.Request()).Tags("面板", "基础信息").With(map[string]any{
						"error": err.Error(),
					}).Info("获取数据库列表失败")
					databaseCount = -1
				} else {
					defer rows.Close()
					var databases []database
					for rows.Next() {
						var d database
						err := rows.Scan(&d.Name)
						if err != nil {
							continue
						}
						if d.Name == "information_schema" || d.Name == "performance_schema" || d.Name == "mysql" || d.Name == "sys" {
							continue
						}

						databases = append(databases, d)
					}
					databaseCount = int64(len(databases))
				}
			}
		}
	}
	if postgresqlInstalled {
		status, err := tools.Exec("systemctl status postgresql | grep Active | grep -v grep | awk '{print $2}'")
		if status == "active" && err == nil {
			raw, err := tools.Exec(`echo "\l" | su - postgres -c "psql"`)
			if err == nil {
				databases := strings.Split(raw, "\n")
				if len(databases) >= 4 {
					databases = databases[3 : len(databases)-1]
					for _, db := range databases {
						parts := strings.Split(db, "|")
						if len(parts) != 9 || len(strings.TrimSpace(parts[0])) == 0 || strings.TrimSpace(parts[0]) == "template0" || strings.TrimSpace(parts[0]) == "template1" || strings.TrimSpace(parts[0]) == "postgres" {
							continue
						}

						databaseCount++
					}
				}
			}
		}
	}

	var ftpCount int64
	var ftpPlugin = r.plugin.GetInstalledBySlug("pureftpd")
	if ftpPlugin.ID != 0 {
		listRaw, err := tools.Exec("pure-pw list")
		if len(listRaw) != 0 && err == nil {
			listArr := strings.Split(listRaw, "\n")
			ftpCount = int64(len(listArr))
		}
	}

	var cronCount int64
	err = facades.Orm().Query().Model(models.Cron{}).Count(&cronCount)
	if err != nil {
		cronCount = -1
	}

	return Success(ctx, http.Json{
		"website":  websiteCount,
		"database": databaseCount,
		"ftp":      ftpCount,
		"cron":     cronCount,
	})
}

// InstalledDbAndPhp 获取已安装的数据库和 PHP 版本
func (r *InfoController) InstalledDbAndPhp(ctx http.Context) http.Response {
	var php []models.Plugin
	err := facades.Orm().Query().Where("slug like ?", "php%").Find(&php)
	if err != nil {
		return ErrorSystem(ctx)
	}

	var mysql models.Plugin
	mysqlInstalled := true
	err = facades.Orm().Query().Where("slug like ?", "mysql%").FirstOrFail(&mysql)
	if err != nil {
		mysqlInstalled = false
	}

	var postgresql models.Plugin
	postgresqlInstalled := true
	err = facades.Orm().Query().Where("slug like ?", "postgresql%").FirstOrFail(&postgresql)
	if err != nil {
		postgresqlInstalled = false
	}

	type data struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	var phpData []data
	var dbData []data
	phpData = append(phpData, data{Value: "0", Label: "不使用"})
	dbData = append(dbData, data{Value: "0", Label: "不使用"})
	for _, p := range php {
		match := regexp.MustCompile(`php(\d+)`).FindStringSubmatch(p.Slug)
		if len(match) == 0 {
			continue
		}

		phpData = append(phpData, data{Value: strings.ReplaceAll(p.Slug, "php", ""), Label: r.plugin.GetBySlug(p.Slug).Name})
	}

	if mysqlInstalled {
		dbData = append(dbData, data{Value: "mysql", Label: "MySQL"})
	}
	if postgresqlInstalled {
		dbData = append(dbData, data{Value: "postgresql", Label: "PostgreSQL"})
	}

	return Success(ctx, http.Json{
		"php": phpData,
		"db":  dbData,
	})
}

// CheckUpdate 检查面板更新
func (r *InfoController) CheckUpdate(ctx http.Context) http.Response {
	version := facades.Config().GetString("panel.version")
	remote, err := tools.GetLatestPanelVersion()
	if err != nil {
		return Error(ctx, http.StatusInternalServerError, "获取最新版本失败")
	}

	if tools.VersionCompare(version, remote.Version, ">=") {
		return Success(ctx, http.Json{
			"update": false,
		})
	}

	return Success(ctx, http.Json{
		"update": true,
	})
}

// UpdateInfo 获取更新信息
func (r *InfoController) UpdateInfo(ctx http.Context) http.Response {
	version := facades.Config().GetString("panel.version")
	current, err := tools.GetLatestPanelVersion()
	if err != nil {
		return Error(ctx, http.StatusInternalServerError, "获取最新版本失败")
	}

	if tools.VersionCompare(version, current.Version, ">=") {
		return Error(ctx, http.StatusInternalServerError, "当前版本已是最新版本")
	}

	versions, err := tools.GenerateVersions(version, current.Version)
	if err != nil {
		return Error(ctx, http.StatusInternalServerError, "获取更新信息失败")
	}

	var versionInfo []tools.PanelInfo
	for _, v := range versions {
		info, err := tools.GetPanelVersion(v)
		if err != nil {
			continue
		}

		versionInfo = append(versionInfo, info)
	}

	return Success(ctx, versionInfo)
}

// Update 更新面板
func (r *InfoController) Update(ctx http.Context) http.Response {
	var task models.Task
	err := facades.Orm().Query().Where("status", models.TaskStatusRunning).OrWhere("status", models.TaskStatusWaiting).FirstOrFail(&task)
	if err == nil {
		return Error(ctx, http.StatusInternalServerError, "当前有任务正在执行，禁止更新")
	}

	panel, err := tools.GetLatestPanelVersion()
	if err != nil {
		facades.Log().Request(ctx.Request()).Tags("面板", "基础信息").With(map[string]any{
			"error": err.Error(),
		}).Info("获取最新版本失败")
		return Error(ctx, http.StatusInternalServerError, "获取最新版本失败")
	}

	err = tools.UpdatePanel(panel)
	if err != nil {
		facades.Log().Request(ctx.Request()).Tags("面板", "基础信息").With(map[string]any{
			"error": err.Error(),
		}).Info("更新面板失败")
		return Error(ctx, http.StatusInternalServerError, err.Error())
	}

	tools.RestartPanel()
	return Success(ctx, nil)
}

// Restart 重启面板
func (r *InfoController) Restart(ctx http.Context) http.Response {
	var task models.Task
	err := facades.Orm().Query().Where("status", models.TaskStatusRunning).OrWhere("status", models.TaskStatusWaiting).FirstOrFail(&task)
	if err == nil {
		return Error(ctx, http.StatusInternalServerError, "当前有任务正在执行，禁止重启")
	}

	tools.RestartPanel()
	return Success(ctx, nil)
}
