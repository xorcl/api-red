package common

type CronTask struct {
	Name    string
	Time    string
	Execute func() error
}
