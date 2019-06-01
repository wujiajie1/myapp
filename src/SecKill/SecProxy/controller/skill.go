package controller

import (
	"SecKill/SecProxy/service"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"strconv"
	"strings"
	"time"
)

type SkillController struct {
	beego.Controller
}
//秒杀接入
///seckill?product=20&source=android&authcode=xx&time=xx&nance=xx
func (p *SkillController) SecKill(){
	productId,err := p.GetInt("product_id")
	result := make(map[string]interface{})
	result["code"] = 200
	result["message"] = "success"
	defer func() {
		p.Data["json"] = result
		p.ServeJSON()
	}()
	if err != nil {
		result["code"] = 1001
		result["message"] = "invalid product_id"
		return
	}
	source := p.GetString("src")
	authcode := p.GetString("authcode")
	secTime := p.GetString("time")
	nance := p.GetString("nance")

	secRequest := &service.SecRequest{}
	secRequest.ProductId = productId
	secRequest.Source = source
	secRequest.AuthCode = authcode
	secRequest.SecTime = secTime
	secRequest.Nance = nance
	secRequest.UserAuthSign = p.Ctx.GetCookie("userAuthSign")
	secRequest.UserId,err = strconv.Atoi(p.Ctx.GetCookie("userId"))
	secRequest.AccessTime = time.Now()
	if len(p.Ctx.Request.RemoteAddr) > 0 {
		secRequest.ClientAddr = strings.Split(p.Ctx.Request.RemoteAddr, ":")[0]
	}

	if err != nil {
		result["code"] = service.ErrInvalidRequest
		result["message"] = err.Error()

		logs.Error("invalid request, get userId failed, err:%v", err)
		return
	}
	//service处理业务逻辑
	data, code, err := service.SecKill(secRequest)
	if err != nil {
		result["code"] = service.ErrInvalidRequest
		result["message"] = err.Error()
		return
	}
	result["code"] = data
	result["message"] = code
	return

}
//秒杀信息
func (p *SkillController) SecInfo(){
	productId,err := p.GetInt("product_id")
	result := make(map[string]interface{})
	result["code"] = 200
	result["message"] = "success"
	defer func() {
		p.Data["json"] = result
		p.ServeJSON()
	}()
	if err != nil {
		data, code, err := service.SecInfoList()
		if err != nil {
			result["code"] = code
			result["message"] = err.Error()

			logs.Error("invalid request, get product_id failed, err:%v", err)
			return
		}

		result["code"] = code
		result["data"] = data
	}else {
		data,code,err := service.SecInfo(productId)
		if err != nil {
			result["code"] = code
			result["message"] = err.Error()
			logs.Error("invalid request, get product_id failed, err:%v", err)
			return
		}
		result["code"] = code
		result["data"] = data
	}

}