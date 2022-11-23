package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"k8s_platform/service"
)

var Deployment deployment

type deployment struct {
}

//获取deployment列表
func (d *deployment) GetDeployments(ctx *gin.Context) {
	params := new(struct {
		FilterName string `form:"filter_name"`
		Namespace  string `form:"namespace"`
		Limit      int    `form:"limit"`
		Page       int    `form:"page"`
	})
	if err := ctx.Bind(params); err != nil {
		logger.Error(errors.New("绑定失败," + err.Error()))
		ctx.JSON(500, gin.H{
			"msg":  "绑定失败",
			"data": nil,
		})
		return
	}
	deploymentsResp, err := service.Deployment.GetDeployments(params.FilterName, params.Namespace, params.Limit, params.Page)
	if err != nil {
		logger.Error(errors.New("获取deployment列表失败," + err.Error()))
		ctx.JSON(500, gin.H{
			"msg":  "获取deployment列表失败",
			"data": nil,
		})
		return
	}
	if len(deploymentsResp.Items) > 0 {
		ctx.JSON(200, gin.H{
			"msg":  "获取deployment列表成功",
			"data": deploymentsResp,
		})
	} else {
		ctx.JSON(500, gin.H{
			"msg":  "获取deployment列表失败",
			"data": nil,
		})
	}
}

//获取deployment详情
func (d *deployment) GetDeploymentDetail(ctx *gin.Context) {
	params := new(struct {
		DeploymentName string `form:"deploymentName"`
		Namespace      string `form:"namespace"`
	})
	err := ctx.Bind(params)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "绑定deployment失败",
			"data": nil,
		})
		return
	}
	deploy, err := service.Deployment.GetDeploymentDetail(params.DeploymentName, params.Namespace)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "获取deployment详情失败",
			"data": nil,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"msg":  "获取deployment详情成功",
		"data": deploy,
	})
}

//删除deployment
func (d *deployment) DelDeployment(ctx *gin.Context) {
	params := new(struct {
		DeploymentName string `json:"deploymentName"`
		Namespace      string `json:"namespace"`
	})
	err := ctx.ShouldBindJSON(params)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg": "绑定deployment失败",
		})
		return
	}
	err = service.Deployment.DeleteDeployment(params.DeploymentName, params.Namespace)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg": "删除deployment失败",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"msg": "删除depoyment：" + params.DeploymentName + "成功",
	})

}

//更新deployment
func (d *deployment) UpdateDeployment(ctx *gin.Context) {
	params := new(struct {
		Namespace string `json:"namespace"`
		Content   string `json:"content"`
	})
	if err := ctx.ShouldBindJSON(params); err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "绑定失败",
			"data": nil,
		})
		return
	}
	err := service.Deployment.UpdateDeployment(params.Namespace, params.Content)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "更新deployment失败",
			"data": nil,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"msg": "更新deployment成功",
	})
}

//获取每个namespace里面的deloyment
func (d *deployment) GetDeployNum(ctx *gin.Context) {
	deploymenylist, err := service.Deployment.GetDeployNum()
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "获取deployment失败",
			"data": nil,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"msg":  "获取deployment成功",
		"data": deploymenylist,
	})
}

//创建deployment
func (d *deployment) CreateDeployment(ctx *gin.Context) {
	var (
		deployCreate = new(service.DeployCreate) //定义创建deployment所需的结构体
		err          error
	)
	if err = ctx.ShouldBindJSON(deployCreate); err != nil {
		logger.Error(errors.New("bind请求参数失败," + err.Error()))
		ctx.JSON(500, gin.H{
			"msg":  "bind请求参数失败",
			"data": nil,
		})
		return
	}
	if err = service.Deployment.CreateDeployment(deployCreate); err != nil {
		ctx.JSON(500, gin.H{
			"msg":  "创建deployment: " + deployCreate.Name + " 失败",
			"data": nil,
		})
		return
	}
	ctx.JSON(200, gin.H{
		"msg":  "创建deployment: " + deployCreate.Name + " 成功",
		"data": nil,
	})
}
