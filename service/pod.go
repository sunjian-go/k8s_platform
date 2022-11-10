package service

import (
	"context"
	"errors"
	"github.com/wonderivan/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Pod pod

type pod struct {
}

//定义列表的返回内容，Items是pod元素列表，Total是元素数量
type PodResp struct {
	PodNum int          `json:"podNum"`
	PodArr []corev1.Pod `json:"podArr"`
}

//获取pod列表，支持过滤、排序、分页
func (p *pod) GetPods(filterName, namespace string, limit, page int) (data *PodResp, err error) {
	//获取podList类型的pod列表
	//context.TODO()用于声明一个空的context上下文，用于List方法内设置这个请求的超时（源码），这里的常用用法
	//metav1.ListOptions{}用于过滤List数据，如使用label，field等
	//kubectl get services --all-namespaces --field-seletor metadata.namespace != default
	podList, err := K8s.K8sClientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Info("获取pod列表失败，" + err.Error())            //用于开发者自己调试看
		return nil, errors.New("获取pod列表失败，" + err.Error()) //用于返回给客户端查看
	}
	//实例化dataSelector结构体，组装数据
	selectableData := &dataSelector{
		GenericDataList: p.toCells(podList.Items),
		dataSelectQuery: &DataSelectQuery{
			&FilterQuery{
				Name: filterName,
			},
			&paginateQuery{
				Limit: limit,
				Page:  page,
			},
		},
	}
	//先过滤
	filtered := selectableData.Filter()    //selectableData.Filter返回值是一个dataselect类型的指针
	total := len(filtered.GenericDataList) //所以可以点出GenericDataList方法
	//再排序和分页
	dataslter := filtered.Sort().Paginate()
	//将DataCell类型转成pod
	pods := p.formCells(dataslter.GenericDataList)
	return &PodResp{
		PodNum: total,
		PodArr: pods,
	}, nil
}

//类型转换的方法，将corev1.pod转换为DataCell类型
func (p *pod) toCells(pods []corev1.Pod) []DataCell {
	cells := make([]DataCell, len(pods))
	for i := range pods {
		cells[i] = podCell(pods[i])
	}
	return cells
}

//formCells方法用于将DataCell类型数组，转换成corev1.pod类型数组
func (p *pod) formCells(cells []DataCell) []corev1.Pod {
	pods := make([]corev1.Pod, len(cells))
	for i := range cells {
		//cells[i].(podCell)就使用到了断言，断言后转换成了podCell类型，然后又转成了pod类型
		pods[i] = corev1.Pod(cells[i].(podCell))
	}
	return pods
}
