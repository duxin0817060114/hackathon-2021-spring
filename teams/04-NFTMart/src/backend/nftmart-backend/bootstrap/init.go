package bootstrap

import (
	_ "nftmart/app/core/destroy" // 监听程序退出信号，用于资源的释放
	"nftmart/app/global/my_errors"
	"nftmart/app/global/variable"
	"nftmart/app/http/validator/common/register_validator"
	"nftmart/app/service/sys_log_hook"
	"nftmart/app/utils/gorm_v2"
	"nftmart/app/utils/snow_flake"
	"nftmart/app/utils/websocket/core"
	"nftmart/app/utils/yml_config"
	"nftmart/app/utils/zap_factory"
	"log"
	"os"
)

// 检查项目必须的非编译目录是否存在，避免编译后调用的时候缺失相关目录
func checkRequiredFolders() {
	//1.检查配置文件是否存在
	if _, err := os.Stat(variable.BasePath + "/config/config.yml"); err != nil {
		log.Fatal(my_errors.ErrorsConfigYamlNotExists + err.Error())
	}
	if _, err := os.Stat(variable.BasePath + "/config/gorm_v2.yml"); err != nil {
		log.Fatal(my_errors.ErrorsConfigGormNotExists + err.Error())
	}
	//2.检查public目录是否存在
	if _, err := os.Stat(variable.BasePath + "/public/"); err != nil {
		log.Fatal(my_errors.ErrorsPublicNotExists + err.Error())
	}
	////3.检查Storage/logs 目录是否存在
	//if _, err := os.Stat(variable.BasePath + "/storage/logs/"); err != nil {
	//	log.Fatal(my_errors.ErrorsStorageLogsNotExists + err.Error())
	//}
}

func init() {
	// 1. 初始化 项目根路径，参见 variable 常量包，相关路径：app\global\variable\variable.go

	//2.检查配置文件以及日志目录等非编译性的必要条件
	checkRequiredFolders()

	//3.初始化表单参数验证器，注册在容器
	register_validator.RegisterValidator()

	// 4.启动针对配置文件(confgi.yml、gorm_v2.yml)变化的监听， 配置文件操作指针，初始化为全局变量
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()
	// config>gorm_v2.yml 启动文件变化监听事件
	variable.ConfigGormv2Yml = variable.ConfigYml.Clone("gorm_v2")
	variable.ConfigGormv2Yml.ConfigFileChangeListen()

	// 5.初始化全局日志句柄，并载入日志钩子处理函数
	variable.ZapLog = zap_factory.CreateZapFactory(sys_log_hook.ZapLogHandler)

	// 6.根据配置初始化 gorm mysql 全局 *gorm.Db
	if variable.ConfigGormv2Yml.GetInt("Gormv2.Mysql.IsInitGolobalGormMysql") == 1 {
		if dbMysql, err := gorm_v2.GetOneMysqlClient(); err != nil {
			log.Fatal(my_errors.ErrorsGormInitFail + err.Error())
		} else {
			variable.GormDbMysql = dbMysql
		}
	}
	// 根据配置初始化 gorm sqlserver 全局 *gorm.Db
	if variable.ConfigGormv2Yml.GetInt("Gormv2.Sqlserver.IsInitGolobalGormSqlserver") == 1 {
		if dbSqlserver, err := gorm_v2.GetOneSqlserverClient(); err != nil {
			log.Fatal(my_errors.ErrorsGormInitFail + err.Error())
		} else {
			variable.GormDbSqlserver = dbSqlserver
		}
	}
	// 根据配置初始化 gorm postgresql 全局 *gorm.Db
	if variable.ConfigGormv2Yml.GetInt("Gormv2.PostgreSql.IsInitGolobalGormPostgreSql") == 1 {
		if dbPostgre, err := gorm_v2.GetOnePostgreSqlClient(); err != nil {
			log.Fatal(my_errors.ErrorsGormInitFail + err.Error())
		} else {
			variable.GormDbPostgreSql = dbPostgre
		}
	}

	// 7.雪花算法全局变量
	variable.SnowFlake = snow_flake.CreateSnowflakeFactory()

	// 8.websocket Hub中心启动
	if variable.ConfigYml.GetInt("Websocket.Start") == 1 {
		// websocket 管理中心hub全局初始化一份
		variable.WebsocketHub = core.CreateHubFactory()
		if Wh, ok := variable.WebsocketHub.(*core.Hub); ok {
			go Wh.Run()
		}
	}

}
