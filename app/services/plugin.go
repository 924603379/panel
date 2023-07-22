// Package services 插件服务
package services

import (
	"github.com/goravel/framework/facades"

	"panel/app/models"
	"panel/app/plugins/mysql57"
	"panel/app/plugins/mysql80"
	"panel/app/plugins/openresty"
	"panel/app/plugins/php74"
	"panel/app/plugins/php80"
	"panel/app/plugins/phpmyadmin"
)

// PanelPlugin 插件元数据结构
type PanelPlugin struct {
	Name        string
	Author      string
	Description string
	Slug        string
	Version     string
	Requires    []string
	Excludes    []string
	Install     string
	Uninstall   string
	Update      string
}

type Plugin interface {
	AllInstalled() ([]models.Plugin, error)
	All() []PanelPlugin
	GetBySlug(slug string) PanelPlugin
	GetInstalledBySlug(slug string) models.Plugin
}

type PluginImpl struct {
}

func NewPluginImpl() *PluginImpl {
	return &PluginImpl{}
}

// AllInstalled 获取已安装的所有插件
func (r *PluginImpl) AllInstalled() ([]models.Plugin, error) {
	var plugins []models.Plugin
	if err := facades.Orm().Query().Get(&plugins); err != nil {
		return plugins, err
	}

	return plugins, nil
}

// All 获取所有插件
func (r *PluginImpl) All() []PanelPlugin {
	var p []PanelPlugin

	p = append(p, PanelPlugin{
		Name:        openresty.Name,
		Author:      openresty.Author,
		Description: openresty.Description,
		Slug:        openresty.Slug,
		Version:     openresty.Version,
		Requires:    openresty.Requires,
		Excludes:    openresty.Excludes,
		Install:     openresty.Install,
		Uninstall:   openresty.Uninstall,
		Update:      openresty.Update,
	})
	p = append(p, PanelPlugin{
		Name:        mysql57.Name,
		Author:      mysql57.Author,
		Description: mysql57.Description,
		Slug:        mysql57.Slug,
		Version:     mysql57.Version,
		Requires:    mysql57.Requires,
		Excludes:    mysql57.Excludes,
		Install:     mysql57.Install,
		Uninstall:   mysql57.Uninstall,
		Update:      mysql57.Update,
	})
	p = append(p, PanelPlugin{
		Name:        mysql80.Name,
		Author:      mysql80.Author,
		Description: mysql80.Description,
		Slug:        mysql80.Slug,
		Version:     mysql80.Version,
		Requires:    mysql80.Requires,
		Excludes:    mysql80.Excludes,
		Install:     mysql80.Install,
		Uninstall:   mysql80.Uninstall,
		Update:      mysql80.Update,
	})
	p = append(p, PanelPlugin{
		Name:        php74.Name,
		Author:      php74.Author,
		Description: php74.Description,
		Slug:        php74.Slug,
		Version:     php74.Version,
		Requires:    php74.Requires,
		Excludes:    php74.Excludes,
		Install:     php74.Install,
		Uninstall:   php74.Uninstall,
		Update:      php74.Update,
	})
	p = append(p, PanelPlugin{
		Name:        php80.Name,
		Author:      php80.Author,
		Description: php80.Description,
		Slug:        php80.Slug,
		Version:     php80.Version,
		Requires:    php80.Requires,
		Excludes:    php80.Excludes,
		Install:     php80.Install,
		Uninstall:   php80.Uninstall,
		Update:      php80.Update,
	})
	p = append(p, PanelPlugin{
		Name:        phpmyadmin.Name,
		Author:      phpmyadmin.Author,
		Description: phpmyadmin.Description,
		Slug:        phpmyadmin.Slug,
		Version:     phpmyadmin.Version,
		Requires:    phpmyadmin.Requires,
		Excludes:    phpmyadmin.Excludes,
		Install:     phpmyadmin.Install,
		Uninstall:   phpmyadmin.Uninstall,
		Update:      phpmyadmin.Update,
	})

	return p
}

// GetBySlug 根据slug获取插件
func (r *PluginImpl) GetBySlug(slug string) PanelPlugin {
	for _, item := range r.All() {
		if item.Slug == slug {
			return item
		}
	}

	return PanelPlugin{}
}

// GetInstalledBySlug 根据slug获取已安装的插件
func (r *PluginImpl) GetInstalledBySlug(slug string) models.Plugin {
	var plugin models.Plugin
	if err := facades.Orm().Query().Where("slug", slug).Get(&plugin); err != nil {
		return plugin
	}

	return plugin
}
