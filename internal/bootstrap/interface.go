package bootstrap

type Item interface {
	Check() (bool, error)
	Install() error
	GetName() string
}

//type Runner interface {
//    RunAll([]Item) error
//    RunSingle(Item) error
//}
