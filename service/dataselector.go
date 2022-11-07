package service

import (
	"sort"
	"strings"
	"time"
)

//第一步：定义数据结构
//dateSelect：用于封装排序，过滤，分页的数据类型
type dataSelector struct {
	GenericDataList []DataCell       //资源List类型转换之后的数组
	dataSelectQuery *DataSelectQuery //定义过滤和分页的属性
}

//DataCell接口，用于各种资源list的类型转换，转换后可以使用dataSelector的自定义排序方法
type DataCell interface {
	GetCreation() time.Time
	GetName() string
}

//DataSelectQuery: 定义过滤和分页的属性，过滤：Name,分页：Limit和Page
//Limit是单页的数据条数
//Page是第几页
type DataSelectQuery struct {
	FilterQuery   *FilterQuery
	PaginateQuery *paginateQuery
}

type FilterQuery struct {
	Name string
}
type paginateQuery struct {
	Limit int
	Page  int
}

//第二步：排序
//实现自定义结构的排序，需要重写Len、Swap、Less方法
//Len方法用于获取数组长度
func (d *dataSelector) Len() int {
	return len(d.GenericDataList) //计算出资源转换完成之后的资源数组的长度，直接返回
}

//Swap方法用于数组中的元素在比较大小后的位置交换，可定义升序或降序
func (d *dataSelector) Swap(i, j int) { //直接将资源数组的前一位和后一位调换
	d.GenericDataList[i], d.GenericDataList[j] = d.GenericDataList[j], d.GenericDataList[i]
}

//Less方法用于定义数组中元素排序的“大小”的比较方式
func (d *dataSelector) Less(i, j int) bool {
	a := d.GenericDataList[i].GetCreation() //通过时间进行比较
	b := d.GenericDataList[j].GetCreation()
	return b.Before(a) //判断b时间之前于a时间是否为真，为真返回true，触发Swap方法进行位置调换，反之false
}

//根据重写的以上3个方法用使用sort.Sort进行排序
func (d *dataSelector) Sort() *dataSelector {
	sort.Sort(d) //将实例化的dataSelector数据传入，会根据自定义的len,Swap,Less方法根据时间去排序
	return d
}

//第三步：过滤
//Filter方法用于过滤元素，比较元素的Name属性，若包含，再返回
func (d *dataSelector) Filter() *dataSelector {
	//若Name的传参为空，则返回所有元素
	if d.dataSelectQuery.FilterQuery.Name == " " {
		return d
	}
	//若Name的传参不为空，则返回元素中包含Name的所有元素
	filteredList := []DataCell{} //声明一个新数组，若Name包含，则把数据放进新数组，返回出去
	for _, value := range d.GenericDataList {
		objName := value.GetName()
		if !strings.Contains(objName, d.dataSelectQuery.FilterQuery.Name) {
			continue
		}
		filteredList = append(filteredList, value)
	}
	d.GenericDataList = filteredList
	return d
}

//第四步：分页
//Paginate方法用于数组的分页，根据Limit和Page的传参，取一定范围内的数据，返回
func (d *dataSelector) Paginate() *dataSelector {
	//根据limit和page的入参，定义快捷变量
	limit := d.dataSelectQuery.PaginateQuery.Limit
	page := d.dataSelectQuery.PaginateQuery.Page
	//检验参数的合法性
	if limit <= 0 || page <= 0 {
		return d
	}
	//定义取数范举例：25个元素的数组，limit是10，page是3，startIndex是20，endIndex是30（实际上endIndex是25）
	//limit是每页的个数，page是页
	startIndex := limit * (page - 1)
	endIndex := limit * page
	//处理最后一页，这时候就把endIndex由30改为25了
	if len(d.GenericDataList) < endIndex {
		endIndex = len(d.GenericDataList)
	}
	d.GenericDataList = d.GenericDataList[startIndex:endIndex]
	return d
}
