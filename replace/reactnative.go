package replace

import (
	"crypto/md5"
	"encoding/hex"
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
				if attr == "transform" && css[str][attr] != "" { //处理动画的数组合并
					css[str][attr] = strings.Replace(css[str][attr], "[", "", -1)
					css[str][attr] = strings.Replace(css[str][attr], "]", "", -1)
					value = strings.Replace(value, "[", "", -1)
					value = strings.Replace(value, "]", "", -1)
					css[str][attr] = "[" + css[str][attr] + "," + value + "]"
				} else {
					css[str][attr] = value
				}
			}
		}
	}
	return
}

// h-20 =>  "height":"20px"
func singleCssToMap(str string) (css map[string]string, success bool) {
	css = map[string]string{}
	for _, tv := range append(tempAutoStyle, commonRegexp...) {
		for rp, rps := range tv {
			for _, v2 := range rp.FindAllStringSubmatch(str, -1) { //有触发才会使得这个循环有效，也就是必然匹配了
				for i, v3 := range v2 {
					if i >= 1 {
						rps = strings.Replace(rps, fmt.Sprintf("$%d", i), v3, -1)
					}
				}
				items := strings.SplitN(rps, ";", -1) //多个;分开显示
				for _, item := range items {
					cssTemp := strings.SplitN(item, ":", 2)
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
	if strings.Contains(willReturn, "webkit") { //react厂家标识需要大写
		willReturn = strings.Replace(willReturn, "webkit", "Webkit", -1)
	}
	return willReturn
}

var tempAutoStyle = []map[*regexp.Regexp]string{}

// find tempAutoStyle footer{w-100 h-100}
func findTempAutoStyle(path string) {
	tempAutoStyle = []map[*regexp.Regexp]string{}
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
	rp := regexp.MustCompile(fmt.Sprintf(`%s="([^"]+)"`, myConfig.Params.CommonClass))
	tempAutoStyleArry := rp.FindAllStringSubmatch(string(fileBody), -1)
	if len(tempAutoStyleArry) == 0 {
		return
	}
	for _, v := range tempAutoStyleArry {
		if len(v) != 2 { // all contentStyle --arr
			continue
		}
		arrStyleSplits := strings.SplitN(v[1], "}", -1) // f{w-20 d{h-20 --arr
		for _, v1 := range arrStyleSplits {             // f{w-20 --string
			v2 := strings.SplitN(v1, "{", -1) // f w-20 --arr
			if len(v2) != 2 {
				continue
			}
			tempValue := ""                                     // tempAutoStyle --each map[xx]tempValue
			for _, mapKeyValue := range cssStrToMapArr(v2[1]) { // width:20 --map
				for key, value := range mapKeyValue { // key:value --map
					tempValue += key + ":" + value + ";"
				}
			}
			tempAutoStyle = append(tempAutoStyle, map[*regexp.Regexp]string{ // tempAutoStyle --push item
				regexp.MustCompile(fmt.Sprintf(`^%s$`, strings.Replace(v2[0], " ", "", -1))): tempValue, // 去除空格，强制使用开始结束判断
			})
		}
	}
}
func FindPathToString(path string) {
	findTempAutoStyle(path)
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
				css[md5Attr(attr)] = value
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
	if string(cssByte) == "{}" { //没有数据
		return
	}
	newAutoCss := cssToCover(fmt.Sprintf(autoStyleTpl(), string(cssByte)), []string{"", ""})
	if myConfig.Params.React == "reactnative" {
		newAutoCss = cssHandleWithNative(newAutoCss)
	}
	oldAutoStyleStr := findOldAutoStyle(string(fileBody))
	if !isTheSame(oldAutoStyleStr, newAutoCss) {
		bodyStr := string(fileBody)
		//处理数据 需要md5生成key
		reg := regexp.MustCompile(`autoStyleFun\("[^"]*","([^"]+)"`)
		if myConfig.Params.ReactMode == "multiple" {
			for _, v := range reg.FindAllStringSubmatch(bodyStr, -1) {
				if len(v) != 2 {
					continue
				}
				bodyStr = strings.Replace(bodyStr, v[0], `autoStyleFun("`+md5Attr(v[1])+`","`+v[1]+`"`, -1)
			}
		}
		//写入文件
		if oldAutoStyleStr == "" {
			//没有该自动输出的模板
			fileBodyStr := bodyStr + "\r\n" + newAutoCss
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
		fileBodyStr := strings.Replace(bodyStr, oldAutoStyleStr, newAutoCss, -1)
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

//reactnative font-size:"12px"->font-size:12
func cssHandleWithNative(css string) string {
	css = strings.Replace(css, "px", "", -1)
	valueRegexp := regexp.MustCompile(`"(\d{1,6})"`)
	css = valueRegexp.ReplaceAllStringFunc(css, func(b string) string {
		b = strings.Replace(b, `"`, "", -1)
		return b
	})
	//处理''包裹的值为string,fontWeight为特殊一点
	css = strings.Replace(css, `"'`, `"`, -1)
	css = strings.Replace(css, `'"`, `"`, -1)
	//处理数组或者对象
	css = strings.Replace(css, `"{`, `{`, -1)
	css = strings.Replace(css, `}"`, `}`, -1)
	css = strings.Replace(css, `"[`, `[`, -1)
	css = strings.Replace(css, `]"`, `]`, -1)
	//处理bool值
	css = strings.Replace(css, `"true"`, `true`, -1)
	css = strings.Replace(css, `"false"`, `false`, -1)
	return css
}
func findStyleNeedToAutoToArray(str string) []string {
	reg := &regexp.Regexp{}
	if myConfig.Params.ReactMode == "multiple" {
		//需要一个占位的""来储存自动生成的key
		reg = regexp.MustCompile(`autoStyleFun\("[^"]*","([^"]+)"`)
	} else {
		reg = regexp.MustCompile(`autoStyleFun\("([^"]+)"`)
	}
	cssArry := []string{}
	for _, v := range reg.FindAllStringSubmatch(str, -1) {
		if len(v) != 2 {
			continue
		}
		cssArry = append(cssArry, v[1])
	}
	return cssArry
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
	index := "1"
	if myConfig.Params.ReactMode == "multiple" {
		//这个是为了让输入的样式形成一个提示信息而已，one比较短不需要md5生成key
		index = "2"
	}
	modeStr := ""
	if myConfig.Params.ReactMode == "multiple" {
		modeStr = `_style = autoStyle[data[0]]||{};`
	} else {
		modeStr = `data[0].split(/ /g).filter((value,index)=>{
			return value!==''
		}).forEach(value => {
			_style = Object.assign({},_style, autoStyle[value]||{});
		});`
	}

	if myConfig.Params.React == "react" {
		return `/* autoCssStart */
/* eslint-disable */
const autoStyleFun = (...data)=>{
	let _style = {}
	if(data.length!==0){
		` + modeStr + `
		for(let i=` + index + `;i<data.length;i++){
			_style = Object.assign({},_style,data[i]||{});
		}
	}
	return _style
}
const autoStyle=
%s
/* autoCssEnd */`
	} else {
		return `/* autoCssStart */
/* eslint-disable */
const autoStyleFun = (...data)=>{
	let _style = {}
	if(data.length!=0){
		` + modeStr + `
		for(let i=` + index + `;i<data.length;i++){
			_style = Object.assign({},_style,data[i]||{});
		}
	}
	return _style
}
const autoStyle=StyleSheet.create(%s);
/* autoCssEnd */`
	}
}

// 处理multiple key值过长的问题
func md5Attr(before string) string {
	h := md5.New()
	h.Write([]byte(before))
	md5Str := hex.EncodeToString(h.Sum(nil))
	return md5Str[0 : len(md5Str)/2]
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
