package loadBalance

import "math/rand"

type RandomSelect struct {
}

// 随机选择一个
func (receiver *RandomSelect) SelectOne(arr []string) string {
	//取切片长度范围内的随机数
	n := rand.Intn(len(arr))
	return arr[n]
}
