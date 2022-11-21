package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/wonderivan/logger"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s_platform/config"
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

	//数据处理后的数据和原始数据的比较
	//处理后的数据
	fmt.Println("排序后的数据")
	for _, pod := range pods {
		fmt.Println(pod.Name, pod.CreationTimestamp.Time)
	}
	//原始数据
	fmt.Println("排序前的数据")
	for _, pod := range podList.Items {
		fmt.Println(pod.Name, pod.CreationTimestamp.Time)
	}
	return &PodResp{
		PodNum: total,
		PodArr: pods,
	}, nil
}

/*
类型转换的方法，将corev1.pod转换为DataCell类型(由于corev1.Pod与podcell类型相等，podcell类型的变量又重写了datecell类型接口的函数，
所以podcell作为桥梁，corev1.Pod等于datecell类型，所以可以直接转换)
*/
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

//获取pod详情
func (p *pod) GetPodDetail(podName, namespace string) (pod *corev1.Pod, err error) {
	pod, err = K8s.K8sClientSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{}) //获取单个pod的详情
	if err != nil {
		logger.Error(errors.New("获取Pod详情失败，" + err.Error()))
		return nil, errors.New("获取Pod详情失败，" + err.Error())
	}
	return pod, nil
}

//删除pod
func (p *pod) Deletepod(podName, namepace string) (err error) {
	err = K8s.K8sClientSet.CoreV1().Pods(namepace).Delete(context.TODO(), podName, metav1.DeleteOptions{}) //删除单个pod
	if err != nil {
		logger.Error(errors.New("删除Pod失败，" + err.Error()))
		return errors.New("删除Pod失败，" + err.Error())
	}
	return nil
}

//更新pod
//content参数是请求中传入的pod对象的json数据
func (p *pod) UpdatePod(namespace, content string) (err error) {
	var pod = &corev1.Pod{} //定义一个corev1.Pod类型的空结构体
	//将json格式数据解码为pod对象
	err = json.Unmarshal([]byte(content), pod) //将json格式数据放入字符切片中，解码存入pod结构体
	if err != nil {
		logger.Error(errors.New("解码失败" + err.Error()))
		return errors.New("解码失败" + err.Error())
	}
	//更新pod
	_, err = K8s.K8sClientSet.CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(errors.New("更新pod失败" + err.Error()))
		return errors.New("更新pod失败" + err.Error())
	}
	return nil
}

//获取pod中的容器名
func (p *pod) GetPodContainer(podName, namespace string) (containers []string, err error) {
	//获取pod详情
	pod, err := p.GetPodDetail(podName, namespace)
	if err != nil {
		return nil, err
	}
	//从pod对象中拿到容器名
	for _, conta := range pod.Spec.Containers {
		containers = append(containers, conta.Name)
	}
	return containers, nil
}

//获取容器日志
func (p *pod) GetPodLog(containerName, podName, namespace string) (log string, err error) {
	//设置日志的配置，容器名、tail的行数
	lineLimit := int64(config.PodLogTailLine)
	option := &corev1.PodLogOptions{Container: containerName, TailLines: &lineLimit}
	//获取request实例
	req := K8s.K8sClientSet.CoreV1().Pods(namespace).GetLogs(podName, option)
	//发起request请求，返回一个io.ReadCloser类型（等同于response.body）
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		logger.Error(errors.New("获取PodLog失败" + err.Error()))
		return nil, errors.New("获取PodLog失败" + err.Error())
	}
	defer podLogs.Close() //记得关闭

	//将response body写入缓冲区，目的是为了转成string返回
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		logger.Error("复制PodLog失败，" + err.Error())
		return " ", errors.New("复制PodLog失败" + err.Error())
	}
	return buf.String(), nil
}
