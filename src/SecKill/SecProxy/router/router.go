package router

import (
	"SecKill/SecProxy/controller"
	"github.com/astaxie/beego"
)

func init() {
	//初始化路由
	beego.Router("/seckill",&controller.SkillController{},"*:SecKill")
	//查看活动的状态
	beego.Router("/secinfo",&controller.SkillController{},"*:SecInfo")
}
