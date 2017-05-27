package module

type IModule interface {
	OnInit() bool
	OnDestroy()
	Run()
}
