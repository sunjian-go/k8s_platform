package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s_platform/config"
	"k8s_platform/controller"
	"k8s_platform/service"
)

func main() {
	service.K8s.K8sInit()                                                                                     //初始化k8s client
	pods, err := service.K8s.K8sClientSet.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{}) //通过clientset获取到pod数组
	if err != nil {
		panic("K8sClientSet: " + err.Error())
	}
	//fmt.Printf("type is %T\n", pods.Items)
	for _, pod := range pods.Items { //遍历pod信息
		fmt.Println(pod.Name, pod.Namespace)
	}

	r := gin.Default()                 //初始化路由引擎
	controller.Router.InitApiRouter(r) //使用初始化路由方法
	r.Run(config.ListenAddr)           //启动gin监听
}
