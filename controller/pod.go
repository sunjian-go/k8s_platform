package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"net/http"
)

var Pod pod

type pod struct {
}

//获取pod列表，支持分页、过滤、排序
func (p *pod) Getpods(ctx *gin.Context) {
	//处理入参
	//匿名结构体，用于定义入参，get请求为form格式，其他请求为json格式
	params := new(struct {
		FilterName string `form:"filter_name"`
		Namespace  string `form:"namespace"`
		Limit      int    `form:"limit"`
		Page       int    `form:"page"`
	})
	if err := ctx.Bind(params); err != nil {
		logger.Error("Bind绑定参数失败" + err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":  "Bind绑定参数失败" + err.Error(),
			"data": nil,
		})
	}
}
