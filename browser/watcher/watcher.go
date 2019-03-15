package watcher

import (
	"fmt"
	myConfig "htmlcss/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var fileModTime = map[string]time.Time{}
var suffix *regexp.Regexp

func Handle(cb func(path string)) {
	suffix = regexp.MustCompile(fmt.Sprintf(`%s$`, myConfig.Params.Suffix))
	switch myConfig.Params.Mode {
	case "start": //如果是start，不用延时结束,就不用go
		rangeDirs(myConfig.Params.Dir, cb)
	case "always":
		fallthrough
	case "runtime":
		go func() {
			//初始化文件的修改时间
			rangeDirs(myConfig.Params.Dir, func(path string) {
				osFile, err := os.Open(path)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer osFile.Close()
				fileInfor, err := osFile.Stat()
				if err != nil {
					fmt.Println(err)
					return
				}
				fileModTime[path] = fileInfor.ModTime()
				if myConfig.Params.Mode == "always" { //always需要返回第一次的数值
					cb(path)
				}
			})
			fmt.Println("监听初始化成功!")
			//监听文件的修改并返回地址
			for {
				rangeDirs(myConfig.Params.Dir, func(path string) {
					osFile, err := os.Open(path)
					if err != nil {
						fmt.Println(err)
						return
					}
					defer osFile.Close()
					fileInfor, err := osFile.Stat()
					if err != nil {
						fmt.Println(err)
						return
					}
					if fileModTime[path] != fileInfor.ModTime() {
						cb(path)
						fileModTime[path] = fileInfor.ModTime()
					}
				})
				time.Sleep(time.Millisecond * 500)
			}
		}()
	case "none":
	}

}
func rangeDirs(dir string, cb func(path string)) {
	osFiles, err := ioutil.ReadDir(`` + dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, o := range osFiles {
		if o.IsDir() {
			rangeDirs(filepath.Join(dir, o.Name()), cb)
		} else {
			if suffix.MatchString(o.Name()) && !needIgnore(o.Name()) {
				cb(filepath.Join(dir, o.Name()))
			}
		}
	}
}
func needIgnore(name string) (ok bool) {
	var ignoreArr = strings.SplitN(myConfig.Params.Ignore, ",", -1)
	for _, v := range ignoreArr {
		if strings.Contains(name, v) {
			ok = true
			return
		}
	}
	return
}
