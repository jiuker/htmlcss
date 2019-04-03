package main

import (
	"strings"

	myConfig "github.com/jiuker/htmlcss/config"
	myHttp "github.com/jiuker/htmlcss/http"
	"github.com/jiuker/htmlcss/replace"
	"github.com/jiuker/htmlcss/watcher"
)

func main() {
	//初始化正则
	replace.Init()
	//watcher模块开始工作
	if myConfig.Params.Mode != "none" {
		if strings.Contains(myConfig.Params.Replace, "append-") {
			var allFileByte = []byte{}
			watcher.Handle(func(path string) {
				allFileByte = append(allFileByte, replace.FindFileToGetByte(path)...)
			})
			replace.BtyeToCss(allFileByte, "公共路径!")
		} else {
			watcher.Handle(func(path string) {
				if strings.Contains(path, ".js") && myConfig.Params.React != "none" {
					replace.FindPathToString(path)
				} else {
					replace.FindFileToGetCss(path)
				}
			})
		}
	} else { //只是监听js
		go myHttp.ListenAndServe()
	}
	if myConfig.Params.Mode != "start" { //模式为start不会使用协程，就不用延时结束
		select {}
	}
}
