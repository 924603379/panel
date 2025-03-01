package plugins

import (
	"database/sql"
	"fmt"
	"regexp"

	"github.com/goravel/framework/contracts/http"
	"github.com/spf13/cast"

	"panel/app/http/controllers"
	"panel/app/models"
	"panel/app/services"
	"panel/pkg/tools"
)

type Mysql80Controller struct {
	setting services.Setting
	backup  services.Backup
}

func NewMysql80Controller() *Mysql80Controller {
	return &Mysql80Controller{
		setting: services.NewSettingImpl(),
		backup:  services.NewBackupImpl(),
	}
}

// Status 获取运行状态
func (r *Mysql80Controller) Status(ctx http.Context) http.Response {
	status, err := tools.ServiceStatus("mysqld")
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, "获取MySQL状态失败")
	}

	return controllers.Success(ctx, status)
}

// Reload 重载配置
func (r *Mysql80Controller) Reload(ctx http.Context) http.Response {
	if err := tools.ServiceReload("mysqld"); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, "重载MySQL配置失败")
	}

	return controllers.Success(ctx, nil)
}

// Restart 重启服务
func (r *Mysql80Controller) Restart(ctx http.Context) http.Response {
	if err := tools.ServiceRestart("mysqld"); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, "重启MySQL服务失败")
	}

	return controllers.Success(ctx, nil)
}

// Start 启动服务
func (r *Mysql80Controller) Start(ctx http.Context) http.Response {
	if err := tools.ServiceStart("mysqld"); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, "启动MySQL服务失败")
	}

	return controllers.Success(ctx, nil)
}

// Stop 停止服务
func (r *Mysql80Controller) Stop(ctx http.Context) http.Response {
	if err := tools.ServiceStop("mysqld"); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, "停止MySQL服务失败")
	}

	return controllers.Success(ctx, nil)
}

// GetConfig 获取配置
func (r *Mysql80Controller) GetConfig(ctx http.Context) http.Response {
	config, err := tools.Read("/www/server/mysql/conf/my.cnf")
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, "获取MySQL配置失败")
	}

	return controllers.Success(ctx, config)
}

// SaveConfig 保存配置
func (r *Mysql80Controller) SaveConfig(ctx http.Context) http.Response {
	config := ctx.Request().Input("config")
	if len(config) == 0 {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, "配置不能为空")
	}

	if err := tools.Write("/www/server/mysql/conf/my.cnf", config, 0644); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, "写入MySQL配置失败")
	}

	return r.Restart(ctx)
}

// Load 获取负载
func (r *Mysql80Controller) Load(ctx http.Context) http.Response {
	rootPassword := r.setting.Get(models.SettingKeyMysqlRootPassword)
	if len(rootPassword) == 0 {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, "MySQL root密码为空")
	}

	status, err := tools.ServiceStatus("mysqld")
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, "获取MySQL状态失败")
	}
	if !status {
		return controllers.Error(ctx, http.StatusInternalServerError, "MySQL 未运行")
	}

	raw, err := tools.Exec("/www/server/mysql/bin/mysqladmin -uroot -p" + rootPassword + " extended-status 2>&1")
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, "获取MySQL负载失败")
	}

	var data []map[string]string
	expressions := []struct {
		regex string
		name  string
	}{
		{`Uptime\s+\|\s+(\d+)\s+\|`, "运行时间"},
		{`Queries\s+\|\s+(\d+)\s+\|`, "总查询次数"},
		{`Connections\s+\|\s+(\d+)\s+\|`, "总连接次数"},
		{`Com_commit\s+\|\s+(\d+)\s+\|`, "每秒事务"},
		{`Com_rollback\s+\|\s+(\d+)\s+\|`, "每秒回滚"},
		{`Bytes_sent\s+\|\s+(\d+)\s+\|`, "发送"},
		{`Bytes_received\s+\|\s+(\d+)\s+\|`, "接收"},
		{`Threads_connected\s+\|\s+(\d+)\s+\|`, "活动连接数"},
		{`Max_used_connections\s+\|\s+(\d+)\s+\|`, "峰值连接数"},
		{`Key_read_requests\s+\|\s+(\d+)\s+\|`, "索引命中率"},
		{`Innodb_buffer_pool_reads\s+\|\s+(\d+)\s+\|`, "Innodb索引命中率"},
		{`Created_tmp_disk_tables\s+\|\s+(\d+)\s+\|`, "创建临时表到磁盘"},
		{`Open_tables\s+\|\s+(\d+)\s+\|`, "已打开的表"},
		{`Select_full_join\s+\|\s+(\d+)\s+\|`, "没有使用索引的量"},
		{`Select_full_range_join\s+\|\s+(\d+)\s+\|`, "没有索引的JOIN量"},
		{`Select_range_check\s+\|\s+(\d+)\s+\|`, "没有索引的子查询量"},
		{`Sort_merge_passes\s+\|\s+(\d+)\s+\|`, "排序后的合并次数"},
		{`Table_locks_waited\s+\|\s+(\d+)\s+\|`, "锁表次数"},
	}

	for _, expression := range expressions {
		re := regexp.MustCompile(expression.regex)
		matches := re.FindStringSubmatch(raw)
		if len(matches) > 1 {
			d := map[string]string{"name": expression.name, "value": matches[1]}
			if expression.name == "发送" || expression.name == "接收" {
				d["value"] = tools.FormatBytes(cast.ToFloat64(matches[1]))
			}

			data = append(data, d)
		}
	}

	// 索引命中率
	readRequests := cast.ToFloat64(data[9]["value"])
	reads := cast.ToFloat64(data[10]["value"])
	data[9]["value"] = fmt.Sprintf("%.2f%%", readRequests/(reads+readRequests)*100)
	// Innodb 索引命中率
	bufferPoolReads := cast.ToFloat64(data[11]["value"])
	bufferPoolReadRequests := cast.ToFloat64(data[12]["value"])
	data[10]["value"] = fmt.Sprintf("%.2f%%", bufferPoolReadRequests/(bufferPoolReads+bufferPoolReadRequests)*100)

	return controllers.Success(ctx, data)
}

// ErrorLog 获取错误日志
func (r *Mysql80Controller) ErrorLog(ctx http.Context) http.Response {
	log, err := tools.Exec("tail -n 100 /www/server/mysql/mysql-error.log")
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, log)
	}

	return controllers.Success(ctx, log)
}

// ClearErrorLog 清空错误日志
func (r *Mysql80Controller) ClearErrorLog(ctx http.Context) http.Response {
	if out, err := tools.Exec("echo '' > /www/server/mysql/mysql-error.log"); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}

	return controllers.Success(ctx, nil)
}

// SlowLog 获取慢查询日志
func (r *Mysql80Controller) SlowLog(ctx http.Context) http.Response {
	log, err := tools.Exec("tail -n 100 /www/server/mysql/mysql-slow.log")
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, log)
	}

	return controllers.Success(ctx, log)
}

// ClearSlowLog 清空慢查询日志
func (r *Mysql80Controller) ClearSlowLog(ctx http.Context) http.Response {
	if out, err := tools.Exec("echo '' > /www/server/mysql/mysql-slow.log"); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}
	return controllers.Success(ctx, nil)
}

// GetRootPassword 获取root密码
func (r *Mysql80Controller) GetRootPassword(ctx http.Context) http.Response {
	rootPassword := r.setting.Get(models.SettingKeyMysqlRootPassword)
	if len(rootPassword) == 0 {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, "MySQL root密码为空")
	}

	return controllers.Success(ctx, rootPassword)
}

// SetRootPassword 设置root密码
func (r *Mysql80Controller) SetRootPassword(ctx http.Context) http.Response {
	status, err := tools.ServiceStatus("mysqld")
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, "获取MySQL状态失败")
	}
	if !status {
		return controllers.Error(ctx, http.StatusInternalServerError, "MySQL 未运行")
	}

	rootPassword := ctx.Request().Input("password")
	if len(rootPassword) == 0 {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, "MySQL root密码不能为空")
	}

	oldRootPassword := r.setting.Get(models.SettingKeyMysqlRootPassword)
	if oldRootPassword != rootPassword {
		if _, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + oldRootPassword + " -e \"ALTER USER 'root'@'localhost' IDENTIFIED BY '" + rootPassword + "';\""); err != nil {
			return controllers.Error(ctx, http.StatusInternalServerError, "设置root密码失败")
		}
		if _, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + oldRootPassword + " -e \"FLUSH PRIVILEGES;\""); err != nil {
			return controllers.Error(ctx, http.StatusInternalServerError, "设置root密码失败")
		}
		err := r.setting.Set(models.SettingKeyMysqlRootPassword, rootPassword)
		if err != nil {
			if _, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"ALTER USER 'root'@'localhost' IDENTIFIED BY '" + oldRootPassword + "';\""); err != nil {
				return nil
			}
			if _, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"FLUSH PRIVILEGES;\""); err != nil {
				return nil
			}
			return controllers.Error(ctx, http.StatusInternalServerError, "设置root密码失败")
		}
	}

	return controllers.Success(ctx, nil)
}

// DatabaseList 获取数据库列表
func (r *Mysql80Controller) DatabaseList(ctx http.Context) http.Response {
	rootPassword := r.setting.Get(models.SettingKeyMysqlRootPassword)
	type database struct {
		Name string `json:"name"`
	}

	db, err := sql.Open("mysql", "root:"+rootPassword+"@unix(/tmp/mysql.sock)/")
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var databases []database
	for rows.Next() {
		var d database
		err := rows.Scan(&d.Name)
		if err != nil {
			continue
		}

		databases = append(databases, d)
	}

	if err := rows.Err(); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, "获取数据库列表失败")
	}

	page := ctx.Request().QueryInt("page", 1)
	limit := ctx.Request().QueryInt("limit", 10)
	startIndex := (page - 1) * limit
	endIndex := page * limit
	if startIndex > len(databases) {
		return controllers.Success(ctx, http.Json{
			"total": 0,
			"items": []database{},
		})
	}
	if endIndex > len(databases) {
		endIndex = len(databases)
	}
	pagedDatabases := databases[startIndex:endIndex]

	return controllers.Success(ctx, http.Json{
		"total": len(databases),
		"items": pagedDatabases,
	})
}

// AddDatabase 添加数据库
func (r *Mysql80Controller) AddDatabase(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]string{
		"database": "required|min_len:1|max_len:255|regex:^[a-zA-Z][a-zA-Z0-9_]+$",
		"user":     "required|min_len:1|max_len:255|regex:^[a-zA-Z][a-zA-Z0-9_]+$",
		"password": "required|min_len:8|max_len:255",
	})
	if err != nil {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, err.Error())
	}
	if validator.Fails() {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, validator.Errors().One())
	}

	rootPassword := r.setting.Get(models.SettingKeyMysqlRootPassword)
	database := ctx.Request().Input("database")
	user := ctx.Request().Input("user")
	password := ctx.Request().Input("password")

	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"CREATE DATABASE IF NOT EXISTS " + database + " DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci;\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"CREATE USER '" + user + "'@'localhost' IDENTIFIED BY '" + password + "';\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"GRANT ALL PRIVILEGES ON " + database + ".* TO '" + user + "'@'localhost';\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"FLUSH PRIVILEGES;\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}

	return controllers.Success(ctx, nil)
}

// DeleteDatabase 删除数据库
func (r *Mysql80Controller) DeleteDatabase(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]string{
		"database": "required|min_len:1|max_len:255|regex:^[a-zA-Z][a-zA-Z0-9_]+$|not_in:information_schema,mysql,performance_schema,sys",
	})
	if err != nil {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, err.Error())
	}
	if validator.Fails() {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, validator.Errors().One())
	}

	rootPassword := r.setting.Get(models.SettingKeyMysqlRootPassword)
	database := ctx.Request().Input("database")
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"DROP DATABASE IF EXISTS " + database + ";\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}

	return controllers.Success(ctx, nil)
}

// BackupList 获取备份列表
func (r *Mysql80Controller) BackupList(ctx http.Context) http.Response {
	backupList, err := r.backup.MysqlList()
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	page := ctx.Request().QueryInt("page", 1)
	limit := ctx.Request().QueryInt("limit", 10)
	startIndex := (page - 1) * limit
	endIndex := page * limit
	if startIndex > len(backupList) {
		return controllers.Success(ctx, http.Json{
			"total": 0,
			"items": []services.BackupFile{},
		})
	}
	if endIndex > len(backupList) {
		endIndex = len(backupList)
	}
	pagedBackupList := backupList[startIndex:endIndex]
	if pagedBackupList == nil {
		pagedBackupList = []services.BackupFile{}
	}

	return controllers.Success(ctx, http.Json{
		"total": len(backupList),
		"items": pagedBackupList,
	})
}

// UploadBackup 上传备份
func (r *Mysql80Controller) UploadBackup(ctx http.Context) http.Response {
	file, err := ctx.Request().File("file")
	if err != nil {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, "上传文件失败")
	}

	backupPath := r.setting.Get(models.SettingKeyBackupPath) + "/mysql"
	if !tools.Exists(backupPath) {
		if err = tools.Mkdir(backupPath, 0644); err != nil {
			return nil
		}
	}

	name := file.GetClientOriginalName()
	_, err = file.StoreAs(backupPath, name)
	if err != nil {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, "上传文件失败")
	}

	return controllers.Success(ctx, nil)
}

// CreateBackup 创建备份
func (r *Mysql80Controller) CreateBackup(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]string{
		"database": "required|min_len:1|max_len:255|regex:^[a-zA-Z][a-zA-Z0-9_]+$|not_in:information_schema,mysql,performance_schema,sys",
	})
	if err != nil {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, err.Error())
	}
	if validator.Fails() {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, validator.Errors().One())
	}

	database := ctx.Request().Input("database")
	err = r.backup.MysqlBackup(database)
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return controllers.Success(ctx, nil)
}

// DeleteBackup 删除备份
func (r *Mysql80Controller) DeleteBackup(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]string{
		"name": "required|min_len:1|max_len:255",
	})
	if err != nil {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, err.Error())
	}
	if validator.Fails() {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, validator.Errors().One())
	}

	backupPath := r.setting.Get(models.SettingKeyBackupPath) + "/mysql"
	fileName := ctx.Request().Input("name")
	if err := tools.Remove(backupPath + "/" + fileName); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return controllers.Success(ctx, nil)
}

// RestoreBackup 还原备份
func (r *Mysql80Controller) RestoreBackup(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]string{
		"backup":   "required|min_len:1|max_len:255",
		"database": "required|min_len:1|max_len:255|regex:^[a-zA-Z][a-zA-Z0-9_]+$|not_in:information_schema,mysql,performance_schema,sys",
	})
	if err != nil {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, err.Error())
	}
	if validator.Fails() {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, validator.Errors().One())
	}

	err = r.backup.MysqlRestore(ctx.Request().Input("database"), ctx.Request().Input("backup"))
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return controllers.Success(ctx, nil)
}

// UserList 用户列表
func (r *Mysql80Controller) UserList(ctx http.Context) http.Response {
	type user struct {
		User   string   `json:"user"`
		Host   string   `json:"host"`
		Grants []string `json:"grants"`
	}

	rootPassword := r.setting.Get(models.SettingKeyMysqlRootPassword)
	db, err := sql.Open("mysql", "root:"+rootPassword+"@unix(/tmp/mysql.sock)/")
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT user, host FROM mysql.user")
	if err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var userGrants []user

	for rows.Next() {
		var u user
		err := rows.Scan(&u.User, &u.Host)
		if err != nil {
			continue
		}

		// 查询用户权限
		grantsRows, err := db.Query(fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", u.User, u.Host))
		if err != nil {
			continue
		}
		defer grantsRows.Close()

		for grantsRows.Next() {
			var grant string
			err := grantsRows.Scan(&grant)
			if err != nil {
				continue
			}

			u.Grants = append(u.Grants, grant)
		}

		if err := grantsRows.Err(); err != nil {
			continue
		}

		userGrants = append(userGrants, u)
	}

	if err := rows.Err(); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, "获取用户列表失败")
	}

	page := ctx.Request().QueryInt("page", 1)
	limit := ctx.Request().QueryInt("limit", 10)
	startIndex := (page - 1) * limit
	endIndex := page * limit
	if startIndex > len(userGrants) {
		return controllers.Success(ctx, http.Json{
			"total": 0,
			"items": []user{},
		})
	}
	if endIndex > len(userGrants) {
		endIndex = len(userGrants)
	}
	pagedUserGrants := userGrants[startIndex:endIndex]

	return controllers.Success(ctx, http.Json{
		"total": len(userGrants),
		"items": pagedUserGrants,
	})
}

// AddUser 添加用户
func (r *Mysql80Controller) AddUser(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]string{
		"database": "required|min_len:1|max_len:255|regex:^[a-zA-Z][a-zA-Z0-9_]+$",
		"user":     "required|min_len:1|max_len:255|regex:^[a-zA-Z][a-zA-Z0-9_]+$",
		"password": "required|min_len:8|max_len:255",
	})
	if err != nil {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, err.Error())
	}
	if validator.Fails() {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, validator.Errors().One())
	}

	rootPassword := r.setting.Get(models.SettingKeyMysqlRootPassword)
	user := ctx.Request().Input("user")
	password := ctx.Request().Input("password")
	database := ctx.Request().Input("database")
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"CREATE USER '" + user + "'@'localhost' IDENTIFIED BY '" + password + ";'\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"GRANT ALL PRIVILEGES ON " + database + ".* TO '" + user + "'@'localhost';\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"FLUSH PRIVILEGES;\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}

	return controllers.Success(ctx, nil)
}

// DeleteUser 删除用户
func (r *Mysql80Controller) DeleteUser(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]string{
		"user": "required|min_len:1|max_len:255|regex:^[a-zA-Z][a-zA-Z0-9_]+$",
	})
	if err != nil {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, err.Error())
	}
	if validator.Fails() {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, validator.Errors().One())
	}

	rootPassword := r.setting.Get(models.SettingKeyMysqlRootPassword)
	user := ctx.Request().Input("user")
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"DROP USER '" + user + "'@'localhost';\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}

	return controllers.Success(ctx, nil)
}

// SetUserPassword 设置用户密码
func (r *Mysql80Controller) SetUserPassword(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]string{
		"user":     "required|min_len:1|max_len:255|regex:^[a-zA-Z][a-zA-Z0-9_]+$",
		"password": "required|min_len:8|max_len:255",
	})
	if err != nil {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, err.Error())
	}
	if validator.Fails() {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, validator.Errors().One())
	}

	rootPassword := r.setting.Get(models.SettingKeyMysqlRootPassword)
	user := ctx.Request().Input("user")
	password := ctx.Request().Input("password")
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"ALTER USER '" + user + "'@'localhost' IDENTIFIED BY '" + password + "';\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"FLUSH PRIVILEGES;\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}

	return controllers.Success(ctx, nil)
}

// SetUserPrivileges 设置用户权限
func (r *Mysql80Controller) SetUserPrivileges(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]string{
		"user":     "required|min_len:1|max_len:255|regex:^[a-zA-Z][a-zA-Z0-9_]+$",
		"database": "required|min_len:1|max_len:255",
	})
	if err != nil {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, err.Error())
	}
	if validator.Fails() {
		return controllers.Error(ctx, http.StatusUnprocessableEntity, validator.Errors().One())
	}

	rootPassword := r.setting.Get(models.SettingKeyMysqlRootPassword)
	user := ctx.Request().Input("user")
	database := ctx.Request().Input("database")
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"REVOKE ALL PRIVILEGES ON *.* FROM '" + user + "'@'localhost';\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"GRANT ALL PRIVILEGES ON " + database + ".* TO '" + user + "'@'localhost';\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}
	if out, err := tools.Exec("/www/server/mysql/bin/mysql -uroot -p" + rootPassword + " -e \"FLUSH PRIVILEGES;\""); err != nil {
		return controllers.Error(ctx, http.StatusInternalServerError, out)
	}

	return controllers.Success(ctx, nil)
}
