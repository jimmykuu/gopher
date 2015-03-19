package gopher

//PerType 是一个权限类型.
type PerType uint

const (
	Everyone          PerType = 1 << 0                            //访客             001
	Authenticated             = 1 << 1                            //已登陆           010
	AdministratorOnly         = 1 << 2                            //管理员           100
	Administrator             = AdministratorOnly | Authenticated //管理员也要已登陆 110
)
