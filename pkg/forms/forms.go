package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Form 创建一个自定义 Form struct
// 匿名嵌入 url.Values 字段，用来保持表单字段的值
// 以及一个 Errors 字段，用来保存表单验证错误信息
type Form struct {
	url.Values // anonymous field，Form struct 会继承 url.Values 的所有方法，譬如 Get() 和 Add()
	Errors     errors
}

// New 实现一个 New() 函数，用来初始化一个自定义的 Form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Required 实现一个 Required() 方法，用来检测指定的字段是否为空
func (f *Form) Required(fields ...string) {
	// 使用 fields ...string 作为参数类型，可以让函数接受任意数量的字符串参数，而不需要显式地创建一个字符串切片。
	// 这样可以提高函数的灵活性和可用性，使得函数可以接受不同数量的参数。
	// 如果使用 fields []string 作为参数类型，那么在调用函数时，必须显式地创建一个字符串切片，并将其作为参数传递给函数。
	// 这样会增加函数的调用复杂度，降低函数的可用性。
	// 因此，使用 fields ...string 作为参数类型，可以让函数更加灵活和易用，是一种更好的设计选择。
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// MaxLength 实现一个 MaxLength() 方法，用来检测指定的字段的值的长度是否超过了给定的最大长度
func (f *Form) MaxLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) > d {
		f.Errors.Add(field, fmt.Sprintf("This field is too long (maximum is %d characters)", d))
	}
}

// PermittedValues 实现一个 PermittedValues() 方法，用来检测指定的字段的值是否在指定的值列表中
func (f *Form) PermittedValues(field string, opts ...string) {
	value := f.Get(field)
	if value == "" {
		return
	}
	for _, opt := range opts {
		if value == opt {
			return
		}
	}
	f.Errors.Add(field, "This field is invalid")
}

func (f *Form) MinLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) < d {
		f.Errors.Add(field, fmt.Sprintf("This field is too short (minimum is %d characters)", d))
	}
}

func (f *Form) MatchesPattern(field string, pattern *regexp.Regexp) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if !pattern.MatchString(value) {
		f.Errors.Add(field, "This field is invalid")
	}
}

// Valid 实现一个 Valid() 方法，用来检测表单中是否有任何错误
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}
