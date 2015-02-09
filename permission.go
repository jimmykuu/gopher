package gopher

//PerType 是一个权限类型.
type PerType int

const (
	Everyone      PerType = iota //访客
	Authenticated                //已登陆
	Administrator                //管理员
)
