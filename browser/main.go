package main

import (
	myConfig "htmlcss/config"
	myHttp "htmlcss/http"
	"htmlcss/replace"
	myReplace "htmlcss/replace"
	"htmlcss/watcher"
	"strings"
)

func main() {
	//初始化正则
	myReplace.Init()
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
				replace.FindFileToGetCss(path)
			})
		}
	} else { //只是监听js
		go myHttp.ListenAndServe()
	}
	if myConfig.Params.Mode != "start" { //模式为start不会使用协程，就不用延时结束
		select {}
	}
}
