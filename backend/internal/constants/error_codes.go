package constants

// 统一响应码：code=0 表示成功，业务错误码从 10000 起分段。
const (
	CodeOK                 = 0
	CodeBadRequest         = 40000 // 参数错误
	CodeUnauthorized       = 40100 // 未认证
	CodeForbidden          = 40300 // 无权限
	CodeNotFound           = 40400 // 资源不存在
	CodeConflict           = 40900 // 状态冲突
	CodeInternalError      = 50000 // 内部错误
	CodeUserExists         = 10001 // 用户已存在
	CodeUserNotFound       = 10002 // 用户不存在
	CodeInvalidCredentials = 10003 // 用户名或密码错误
	CodeProductNotFound    = 10004 // 商品不存在
	CodeProductSold        = 10005 // 商品已售出
	CodeOrderStateInvalid  = 10006 // 订单状态非法流转
	CodeOrderNotFound      = 10007 // 订单不存在
	CodeNotOrderOwner      = 10008 // 非订单归属人
	CodeCannotSelfReview   = 10009 // 不能评价自己
	CodeReviewExists       = 10010 // 该订单已评价
	CodeAddressNotFound    = 10011 // 收货地址不存在
	CodeCartItemNotFound   = 10012 // 购物车条目不存在
	CodeMessageNotFound    = 10013 // 消息不存在
	CodeProductOffShelf    = 10014 // 商品已下架
	CodeFileTooLarge       = 10015 // 文件过大
	CodeUnsupportedMedia   = 10016 // 不支持的图片格式
	CodeRedisUnavailable   = 10017 // 消息通道不可用
)

// ErrorCodeMessages 错误码对应的默认提示文案（constants/messages.go 中另有接口文案）。
var ErrorCodeMessages = map[int]string{
	CodeOK:                 "ok",
	CodeBadRequest:         "请求参数错误",
	CodeUnauthorized:       "请先登录",
	CodeForbidden:          "无权执行该操作",
	CodeNotFound:           "资源不存在",
	CodeConflict:           "资源状态冲突",
	CodeInternalError:      "服务器内部错误",
	CodeUserExists:         "该用户名已被注册",
	CodeUserNotFound:       "用户不存在",
	CodeInvalidCredentials: "用户名或密码错误",
	CodeProductNotFound:    "商品不存在",
	CodeProductSold:        "商品已售出，无法下单",
	CodeOrderStateInvalid:  "订单状态不允许该操作",
	CodeOrderNotFound:      "订单不存在",
	CodeNotOrderOwner:      "只有订单归属人才可执行该操作",
	CodeCannotSelfReview:   "不能对自己发布的内容进行评价",
	CodeReviewExists:       "该订单已完成评价",
	CodeAddressNotFound:    "收货地址不存在",
	CodeCartItemNotFound:   "购物车条目不存在",
	CodeMessageNotFound:    "消息不存在",
	CodeProductOffShelf:    "商品已下架，无法购买",
	CodeFileTooLarge:       "上传图片不能超过 5MB",
	CodeUnsupportedMedia:   "仅支持 jpg/jpeg/png/webp 图片",
	CodeRedisUnavailable:   "实时消息通道暂不可用",
}
