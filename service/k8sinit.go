package service

import (
	"github.com/wonderivan/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s_platform/config"
)

var K8s k8s //用于其他文件调用本包内方法函数等

type k8s struct {
	K8sClientSet *kubernetes.Clientset //用于下面获得k8s clinetset，供给其他文件使用
}

func (k *k8s) K8sInit() {
	//将拷贝出来的kubeconfig文件（.kube/config文件）转换为reset.config类型的对象，通过文件中的ip证书等信息访问集群
	conf, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	if err != nil {
		panic("获取k8s client配置失败: " + err.Error())
	}
	//根据reset.config类型的对象，new一个clientset出来
	clientset, err := kubernetes.NewForConfig(conf)
	if err != nil {
		panic("创建k8s clientset失败:" + err.Error())
	} else {
		logger.Info("k8s clientset 初始化成功")
	}
	K8s.K8sClientSet = clientset //将获取到的clientset赋值给k8s结构体中的K8sClientSet指针供于其他文件使用
}
