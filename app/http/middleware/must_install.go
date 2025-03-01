package middleware

import (
	"strings"
	"sync"

	"github.com/goravel/framework/contracts/http"

	"panel/app/services"
)

// MustInstall 确保已安装插件
func MustInstall() http.Middleware {
	return func(ctx http.Context) {
		path := ctx.Request().Path()
		var slug string
		if strings.HasPrefix(path, "/api/panel/website") {
			slug = "openresty"
		} else {
			pathArr := strings.Split(path, "/")
			if len(pathArr) < 4 {
				ctx.Request().AbortWithStatusJson(http.StatusOK, http.Json{
					"code":    http.StatusForbidden,
					"message": "插件不存在",
				})
				return
			}
			slug = pathArr[3]
		}

		plugin := services.NewPluginImpl().GetBySlug(slug)
		installedPlugin := services.NewPluginImpl().GetInstalledBySlug(slug)
		installedPlugins, err := services.NewPluginImpl().AllInstalled()
		if err != nil {
			ctx.Request().AbortWithStatusJson(http.StatusOK, http.Json{
				"code":    http.StatusInternalServerError,
				"message": "系统内部错误",
			})
			return
		}

		if installedPlugin.Version != plugin.Version || installedPlugin.Slug != plugin.Slug {
			ctx.Request().AbortWithStatusJson(http.StatusOK, http.Json{
				"code":    http.StatusForbidden,
				"message": "插件 " + slug + " 需要更新至 " + plugin.Version + " 版本",
			})
			return
		}

		var lock sync.RWMutex
		pluginsMap := make(map[string]bool)

		for _, p := range installedPlugins {
			lock.Lock()
			pluginsMap[p.Slug] = true
			lock.Unlock()
		}

		for _, require := range plugin.Requires {
			lock.RLock()
			_, requireFound := pluginsMap[require]
			lock.RUnlock()
			if !requireFound {
				ctx.Request().AbortWithStatusJson(http.StatusOK, http.Json{
					"code":    http.StatusForbidden,
					"message": "插件 " + slug + " 需要依赖 " + require + " 插件",
				})
				return
			}
		}

		for _, exclude := range plugin.Excludes {
			lock.RLock()
			_, excludeFound := pluginsMap[exclude]
			lock.RUnlock()
			if excludeFound {
				ctx.Request().AbortWithStatusJson(http.StatusOK, http.Json{
					"code":    http.StatusForbidden,
					"message": "插件 " + slug + " 不兼容 " + exclude + " 插件",
				})
				return
			}
		}

		ctx.Request().Next()
	}
}
