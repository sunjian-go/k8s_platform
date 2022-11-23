package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"k8s_platform/service"
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
		return
	}
	data, err := service.Pod.GetPods(params.FilterName, params.Namespace, params.Limit, params.Page) //返回pod的数组和数量
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"mas":  err.Error(),
			"data": nil,
		})
		return
	}
	ctx.JSON(200, gin.H{ //成功返回含有pod数组和数量的结构体
		"msg":  "获取pod列表成功",
		"data": data,
	})
}

//获取pod详细信息的路由处理函数
func (p *pod) GetPodDetail(ctx *gin.Context) {
	//1.先定义需要接收的客户端发送的pod信息的结构体
	params := new(struct {
		PodName   string `form:"pod_name"`
		Namespace string `form:"namespace"`
	})
	err := ctx.Bind(params) //2.将客户端发送的信息绑定到结构体
	if err != nil {
		logger.Error(errors.New("绑定参数失败," + err.Error()))
		ctx.JSON(500, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	data, err := service.Pod.GetPodDetail(params.PodName, params.Namespace) //3.调用函数获取pod详细信息，返回的是一个原生pod信息
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}
	ctx.JSON(200, gin.H{ //4.将信息返回客户端
		"msg":  "获取信息成功",
		"data": data,
	})

}

//删除pod路由处理函数
func (p *pod) DelPod(ctx *gin.Context) {
	params := new(struct {
		PodName   string `json:"podName"`
		Namespace string `json:"namespace"`
	})
	//form格式适用于Bind方法，json格式适用于ShouldBindJSON方法
	err := ctx.ShouldBindJSON(params)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "绑定数据失败",
			"data": nil,
		})
		return
	}
	err = service.Pod.Deletepod(params.PodName, params.Namespace)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "删除失败",
			"data": nil,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"msg": "删除pod：" + params.PodName + "成功",
	})
}

//更新pod路由处理函数
func (p *pod) UpdatePod(ctx *gin.Context) {
	params := new(struct {
		Namespace string `json:"namespace"`
		Content   string `json:"content"`
	})
	err := ctx.ShouldBindJSON(params)
	if err != nil {
		logger.Error("Bind请求参数失败, " + err.Error())
		ctx.JSON(500, gin.H{
			"msg":  "绑定数据失败" + err.Error(),
			"data": nil,
		})
		return
	}
	err = service.Pod.UpdatePod(params.Namespace, params.Content)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "更新pod失败",
			"data": nil,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"msg":  "更新pod成功",
		"data": nil,
	})
}

//获取容器名函数
func (p *pod) GetContName(ctx *gin.Context) {
	params := new(struct {
		PodName   string `form:"podName"`
		Namespace string `form:"namespace"`
	})
	err := ctx.Bind(params)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "数据绑定失败",
			"data": nil,
		})
		return
	}
	conts, err := service.Pod.GetPodContainer(params.PodName, params.Namespace)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "获取容器名失败",
			"data": nil,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"msg":  "获取容器名成功",
		"data": conts,
	})
}

//获取日志
func (p *pod) GetLogs(ctx *gin.Context) {
	params := new(struct {
		ContName  string `form:"contName"`
		PodName   string `form:"podName"`
		Namespace string `form:"namespace"`
	})
	err := ctx.Bind(params)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "绑定数据失败",
			"data": nil,
		})
		return
	}
	var podlogs string
	podlogs, err = service.Pod.GetPodLog(params.ContName, params.PodName, params.Namespace)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "获取容器日志失败",
			"data": nil,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"msg":  "获取容器日志成功",
		"data": podlogs,
	})
}

//获取pod数量
func (p *pod) GetPodNum(ctx *gin.Context) {
	podsNps, err := service.Pod.GetPodNum()
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "获取pod数量失败",
			"data": nil,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"msg":  "获取pod数量成功",
		"data": podsNps,
	})
}
