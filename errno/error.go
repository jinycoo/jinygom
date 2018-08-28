package errno

var ErrCode map[int]string

func init() {
	ErrCode = make(map[int]string)
	ErrCode[0] = "操作成功"
	ErrCode[1] = "欢迎来到JinyGo！"

	ErrCode[200] = "请求成功"
	ErrCode[404] = "请求地址不存在"

	ErrCode[20000] = "请求参数有误，具体请参考接口文档"
	ErrCode[20001] = "所需参数缺失，具体请参考接口文档"
}