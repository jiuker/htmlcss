package replace

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	myConfig "github.com/jiuker/htmlcss/config"
)

// w-20 h-20 => "h-20 w-20":{"width":"20px","height":"20px"}
func cssStrToMapArr(str string) (css map[string]map[string]string) {
	css = map[string]map[string]string{}
	cssStrSplite := []string{}
	for _, v := range strings.SplitN(str, " ", -1) {
		if v != "" {
			cssStrSplite = append(cssStrSplite, v)
		}
	}
	for _, v := range cssStrSplite {
		mapCssValue, ok := singleCssToMap(v)
		if ok {
			for attr, value := range mapCssValue {
				if _, exit := css[str]; !exit {
					css[str] = map[string]string{}
				}
				css[str][attr] = value
			}
		}
	}
	return
}

// h-20 =>  "height":"20px"
func singleCssToMap(str string) (css map[string]string, success bool) {
	css = map[string]string{}
	for rp, rps := range commonRegexp {
		for _, v2 := range rp.FindAllStringSubmatch(str, -1) { //有触发才会使得这个循环有效，也就是必然匹配了
			for i, v3 := range v2 {
				if i >= 1 {
					rps = strings.Replace(rps, fmt.Sprintf("$%d", i), v3, -1)
				}
			}
			items := strings.SplitN(rps, ";", -1) //多个;分开显示
			for _, item := range items {
				cssTemp := strings.SplitN(item, ":", -1)
				if len(cssTemp) != 2 {
					continue
				} else {
					css[preToUpper(cssTemp[0])] = strings.Trim(strings.Replace(cssTemp[1], ";", "", -1), " ")
					success = true
				}
			}
			return
		}
	}
	return
}

// line-height ->lineHeight
func preToUpper(before string) string {
	willReturn := ""
	before = strings.Replace(before, " ", "", -1)  //去除空格
	beforeArray := strings.SplitN(before, "-", -1) //[line height]
	for _, item := range beforeArray {
		if len(willReturn) != 0 {
			willReturn += strings.ToUpper(string(item[0])) + string(item[1:])
		} else {
			willReturn += item
		}
	}
	return willReturn
}
func FindPathToString(path string) {
	file, err := os.OpenFile(path, os.O_RDWR, 0x666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fileBody, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	css := map[string]map[string]string{}
	for _, lineStyle := range findStyleNeedToAutoToArray(string(fileBody)) {
		if myConfig.Params.ReactMode == "multiple" {
			for attr, value := range cssStrToMapArr(lineStyle) {
				css[attr] = value
			}
		} else {
			lineStyleOneS := strings.SplitN(lineStyle, " ", -1) // "w-20 h-20" => ["w-20" "h-20"]
			for _, v := range lineStyleOneS {
				if v != "" {
					for attr, value := range cssStrToMapArr(v) {
						css[attr] = value
					}
				}
			}
		}
	}
	cssByte, err := json.MarshalIndent(&css, "", "    ")
	if err != nil {
		fmt.Println(err)
	}
	newAutoCss := cssToCover(fmt.Sprintf(autoStyleTpl(), string(cssByte)), []string{"", ""})
	oldAutoStyleStr := findOldAutoStyle(string(fileBody))
	if !isTheSame(oldAutoStyleStr, newAutoCss) {
		//写入文件
		if oldAutoStyleStr == "" {
			//没有该自动输出的模板
			fileBodyStr := string(fileBody) + "\r\n" + newAutoCss
			file1, err := os.OpenFile(path, os.O_TRUNC|os.O_RDWR, 0x666)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer file1.Close()
			_, err = file1.WriteString(fileBodyStr)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("findReact   ", path, "   changed!")
			return
		}
		fileBodyStr := strings.Replace(string(fileBody), oldAutoStyleStr, newAutoCss, -1)
		file1, err := os.OpenFile(path, os.O_TRUNC|os.O_RDWR, 0x666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file1.Close()
		_, err = file1.WriteString(fileBodyStr)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("findReact   ", path, "   changed!")
	}
}
func findStyleNeedToAutoToArray(str string) []string {
	reg, err := regexp.Compile(`autoStyleFun\("([^"]+)"`)
	if err != nil {
		fmt.Println(err)
	}
	cssArry := []string{}
	for _, v := range reg.FindAllStringSubmatch(str, -1) {
		if len(v) != 2 {
			continue
		}
		cssArry = append(cssArry, v[1])
	}
	return cssArry
	//todo 去重
}
func findOldAutoStyle(str string) string {
	strs := strings.Split(str, `/* autoCssStart */`)
	if len(strs) != 2 {
		fmt.Println("没有自动生成段，插入")
		return ""
	} else {
		strs := strings.Split(`/* autoCssStart */`+strs[1], `/* autoCssEnd */`)
		if len(strs) != 2 {
			fmt.Println("不符合自动生成的闭合条件")
		}
		return strs[0] + `/* autoCssEnd */`
	}
	return ""
}
func autoStyleTpl() string {
	modeStr := ""
	if myConfig.Params.ReactMode == "multiple" {
		modeStr = `_style = autoStyle[data[0]]||{};`
	} else {
		modeStr = `data[0].split(/ /g).filter((value,index)=>{
			return value!=''
		}).forEach(value => {
			_style = Object.assign(_style, autoStyle[value]||{});
		});`
	}

	if myConfig.Params.React == "react" {
		return `/* autoCssStart */
const autoStyleFun = (...data)=>{
	let _style = {}
	if(data.length!=0){
		` + modeStr + `
		for(let i=1;i<data.length;i++){
			_style = Object.assign(_style, data[i]||{});
		}
	}
	return _style
}
const autoStyle=
%s
/* autoCssEnd */`
	} else {
		return `/* autoCssStart */
const autoStyleFun = (...data)=>{
	let _style = {}
	if(data.length!=0){
		` + modeStr + `
		for(let i=1;i<data.length;i++){
			_style = Object.assign(_style, data[i]);
		}
	}
	return _style
}
const autoStyle=StyleSheet.create(%s);
/* autoCssEnd */`
	}
}

func isTheSame(str, str1 string) (same bool) {
	compareStr := `1234567890-=qwertyuiop[]\asdfghjkl;'zxcvbnm,./!*@`
	for _, v := range compareStr {
		if strings.Count(str, string(v)) == strings.Count(str1, string(v)) {

		} else {
			same = false
			return
		}
	}
	same = true
	return
}
