package ui

import "embed"

//go:embed "html" "static"
var Files embed.FS

// 第5行的 //go:embed "html" "static" 是一个特殊的注释，它告诉 Go 工具链在编译时将 html 和 static 目录嵌入到二进制文件中。
// 特别注意 开头的 // 和 go:embed 之间没有空格，否则注释将不会被识别。