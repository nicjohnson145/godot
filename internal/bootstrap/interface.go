package bootstrap

type Item interface {
	Check() (bool, error)
	Install() error
}

//type Runner interface {
//    RunAll([]Item) error
//    RunSingle(Item) error
//}
