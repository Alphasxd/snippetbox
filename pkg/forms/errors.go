package forms

// 定义一个 errors 类型，用来保存表单验证错误信息，表单字段名为键
type errors map[string][]string

// 实现一个 Add() 方法，将给定的 field 和 message 添加到 map 中
func (e errors) Add(field, message string) {
    e[field] = append(e[field], message)
}

// 实现一个 Get() 方法，获取给定字段的第一个错误信息
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}