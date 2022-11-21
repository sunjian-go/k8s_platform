package controller

import "github.com/gin-gonic/gin"

//初始化路由结构体,首字母大写，用于跨包调用
var Router router //主要作用是调用下面的方法，如果有多个变量或者方法，都可通过controller.Router.xxxx进行调用，方便使用

//定义路由结构体类型,或者随便一个类型都可以，主要用于其他文件方便调用本包内的函数或者方法、变量等
type router struct {
}

func (r *router) InitApiRouter(router *gin.Engine) { //用router类型的指针去调用这个方法
	router.GET("/api/k8s/pods", Pod.Getpods)
	router.GET("/api/k8s/PodDetail", Pod.GetPodDetail)
	router.DELETE("/api/k8s/delete", Pod.DelPod)
	router.PUT("/api/k8s/update", Pod.UpdatePod)
	router.GET("/api/k8s/GetContName", Pod.GetContName)
}
