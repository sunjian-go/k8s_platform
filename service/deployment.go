package service

import (
	"context"
	"errors"
	"github.com/wonderivan/logger"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/json"
)

var Deployment deployment

type deployment struct {
}

//定义列表的返回内容，Items是deployment的元素列表，Total为deployment元素数量
type DeploymentResp struct {
	Items []appsv1.Deployment `json:"items"`
	Total int                 `json:"total"`
}

//定义DeploysNp类型，用于返回namespace中deployment的数量
type DeploysNp struct {
	Namespace string `json:"namespace"`
	DeployNum int    `json:"deployNum"`
}

//定义DeployCreate结构体，用于创建deployment需要的参数属性的定义
type DeployCreate struct {
	Name          string            `json:"name"`
	Namespace     string            `json:"namespace"`
	Replicas      int32             `json:"replicas"`
	Image         string            `json:"image"`
	Label         map[string]string `json:"label"`
	Cpu           string            `json:"cpu"`
	Memory        string            `json:"memory"`
	ContainerPort int32             `json:"containerPort"`
	HealthCheck   bool              `json:"healthCheck"`
	HealthPath    string            `json:"healthPath"`
}

/*
类型转换的方法，将appsv1.Deployment转换为DataCell类型(由于appsv1.Deployment与deployCell类型相等，deployCell类型的变量又重写了datecell类型接口的函数，
所以deployCell作为桥梁，appsv1.Deployment等于datecell类型，所以可以直接转换)
*/
func (d *deployment) toCells(deploy []appsv1.Deployment) []DataCell {
	cells := make([]DataCell, len(deploy))
	for i := range deploy {
		cells[i] = deployCell(deploy[i])
	}
	return cells
}

//formCells方法用于将DataCell类型数组，转换成appsv1.Deployment类型数组
func (d *deployment) formCells(cells []DataCell) []appsv1.Deployment {
	deploys := make([]appsv1.Deployment, len(cells))
	for i := range cells {
		//cells[i].(appsv1.Deployment)就使用到了断言，断言后转换成了appsv1.Deployment类型，然后又转成了appsv1.Deployment类型
		deploys[i] = appsv1.Deployment(cells[i].(deployCell))

	}
	return deploys
}

//创建deployment
func (d *deployment) CreateDeployment(data *DeployCreate) (err error) {
	//将data中的数据组装成appsv1.Deployment对象
	deployment := &appsv1.Deployment{
		//ObjectMeta中定义资源名、命名空间以及标签
		ObjectMeta: metav1.ObjectMeta{
			Name:      data.Name,
			Namespace: data.Namespace,
			Labels:    data.Label,
		},
		//Spec中定义副本数、选择器、以及pod属性
		Spec: appsv1.DeploymentSpec{
			Replicas: &data.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: data.Label,
			},
			Template: corev1.PodTemplateSpec{
				//定义pod名和标签
				ObjectMeta: metav1.ObjectMeta{
					Name:   data.Name,
					Labels: data.Label,
				},
				//定义容器名、镜像和端口
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  data.Name,
							Image: data.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: data.ContainerPort,
								},
							},
						},
					},
				},
			},
		},
		//Status定义资源的运行状态，这里由于是新建，传入空的appsv1.DeploymentStatus{}对象即可
		Status: appsv1.DeploymentStatus{},
	}
	//判断健康检查功能是否打开，若打开，则定义ReadinessProbe和LivenessProbe
	if data.HealthCheck {
		//设置第一个容器的ReadinessProbe，因为我们pod中只有一个容器，所以直接使用index 0即可
		//若pod中有多个容器，则这里需要使用for循环去定义了
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: data.HealthPath,
					//intstr.IntOrString的作用是端口可以定义为整型，也可以定义为字符串
					//Type=0则表示表示该结构体实例内的数据为整型，转json时只使用IntVal的数据
					//Type=1则表示表示该结构体实例内的数据为字符串，转json时只使用StrVal的数据
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: data.ContainerPort,
					},
				},
			},
			//初始化等待时间
			InitialDelaySeconds: 5,
			//超时时间
			TimeoutSeconds: 5,
			//执行间隔
			PeriodSeconds: 5,
		}
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: data.HealthPath,
					Port: intstr.IntOrString{
						Type:   0,
						IntVal: data.ContainerPort,
					},
				},
			},
			InitialDelaySeconds: 15,
			TimeoutSeconds:      5,
			PeriodSeconds:       5,
		}
		//定义容器的limit和request资源
		deployment.Spec.Template.Spec.Containers[0].Resources.Limits = map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse(data.Cpu),
			corev1.ResourceMemory: resource.MustParse(data.Memory),
		}
		deployment.Spec.Template.Spec.Containers[0].Resources.Requests = map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse(data.Cpu),
			corev1.ResourceMemory: resource.MustParse(data.Memory),
		}
	}
	//调用sdk创建deployment
	_, err = K8s.K8sClientSet.AppsV1().Deployments(data.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		logger.Error(errors.New("创建deployment失败，" + err.Error()))
		return errors.New("创建deployment失败," + err.Error())
	}
	return nil
}

//获取deployment列表，支持过滤、排序、分页
func (d *deployment) GetDeployments(filterName, namespace string, limit, page int) (deploymentsResp *DeploymentResp, err error) {
	//获取deploymentList类型的deployment列表
	deploymentList, err :=
		K8s.K8sClientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error(errors.New("获取Deployment列表失败，" + err.Error()))
		return nil, errors.New("获取Deployment列表失败，" + err.Error())
	}
	//将deploymentList中的deployment列表(Items)，放进dataselector对象中，进行排序
	selectableData := &dataSelector{
		GenericDataList: d.toCells(deploymentList.Items),
		dataSelectQuery: &DataSelectQuery{
			FilterQuery: &FilterQuery{Name: filterName},
			PaginateQuery: &paginateQuery{
				Limit: limit,
				Page:  page,
			},
		},
	}
	filtered := selectableData.Filter()
	total := len(filtered.GenericDataList)
	data := filtered.Sort().Paginate()

	//将[]DataCell类型的deployment列表转为appsv1.deployment列表
	deployments := d.formCells(data.GenericDataList)
	return &DeploymentResp{
		Items: deployments,
		Total: total,
	}, nil

}

//获取deployment详情
func (d *deployment) GetDeploymentDetail(deploymentName, namespace string) (deployment *appsv1.Deployment, err error) {
	deployment, err = K8s.K8sClientSet.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		logger.Error(errors.New("获取deployment详情失败," + err.Error()))
		return nil, errors.New("获取deployment详情失败，" + err.Error())
	}
	return deployment, nil
}

//删除deployment
func (d *deployment) DeleteDeployment(deploymentName, namespace string) (err error) {
	err = K8s.K8sClientSet.AppsV1().Deployments(namespace).Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		logger.Error(errors.New("删除deployment失败," + err.Error()))
		return errors.New("删除deployment失败," + err.Error())
	}
	return nil
}

//更新deployment
func (d *deployment) UpdateDeployment(namespace, content string) (err error) {
	var deploy = &appsv1.Deployment{}
	err = json.Unmarshal([]byte(content), deploy)
	if err != nil {
		logger.Error(errors.New("反序列化失败，" + err.Error()))
		return errors.New("反序列化失败," + err.Error())
	}
	_, err = K8s.K8sClientSet.AppsV1().Deployments(namespace).Update(context.TODO(), deploy, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(errors.New("更新deployment失败，" + err.Error()))
		return errors.New("更新deployment失败," + err.Error())
	}
	return nil
}

//获取每个namespace的deployment的数量
func (d *deployment) GetDeployNum() (deploysNps []*DeploysNp, err error) {
	namespaceList, err := K8s.K8sClientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error(errors.New("获取namespace失败，" + err.Error()))
		return nil, errors.New("获取namespace失败，" + err.Error())
	}
	for _, namespace := range namespaceList.Items {
		deploymentList, err :=
			K8s.K8sClientSet.AppsV1().Deployments(namespace.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logger.Error(errors.New("获取deployment数量失败，" + err.Error()))
			return nil, errors.New("获取deployment数量失败，" + err.Error())
		}
		deploysNp := &DeploysNp{
			Namespace: namespace.Name,
			DeployNum: len(deploymentList.Items),
		}
		deploysNps = append(deploysNps, deploysNp)
	}
	return deploysNps, nil
}
