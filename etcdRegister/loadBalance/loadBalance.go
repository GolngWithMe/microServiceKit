package loadBalance

type LoadBalance interface {
	SelectOne([]string) string
}
