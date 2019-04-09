package replace

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	myConfig "github.com/jiuker/htmlcss/config"
	myHttp "github.com/jiuker/htmlcss/http"

	"strconv"

	"os"

	"regexp"
	"strings"
)

var commonRegexp = []map[*regexp.Regexp]string{}
var Ctype = ""        //convert的auto,优先
var BaseFloat float64 //convert的数值
var Unit = ""         //convert的里面单位，px,rem,upx,
var NodeName = ""     //node里面的节点名字 ,需要查找的节点名字
var NodeNum = 0       //node里面的顺序第几个
var matchCssRegexp = map[string][]string{
	"default": []string{},
	"common":  []string{},
}

func Init() {
	//初始化正则表达式
	rep := regexp.MustCompile(`/(\^[^$]+\$)/`)
	reps := []string{}
	for _, v := range rep.FindAllStringSubmatch(myHttp.SyncJs, -1) {
		for i1, v1 := range v {
			if i1 == 1 {
				reps = append(reps, v1)
			}
		}
	}
	rp := regexp.MustCompile(`rep: (.+)\n`)
	rps := []string{}
	for _, v := range rp.FindAllStringSubmatch(myHttp.SyncJs, -1) {
		for i1, v1 := range v {
			if i1 == 1 {
				rps = append(rps, strings.Replace(v1, `"`, "", -1))
			}
		}
	}
	if len(reps) != len(rps) {
		fmt.Println("正则匹配错误！")
		os.Exit(0)
	}
	for i, v := range reps {
		commonRegexp = append(commonRegexp, map[*regexp.Regexp]string{
			regexp.MustCompile(v): rps[i],
		})
	}
	if myConfig.Params.React == "reactnative" {
		//react-native单独走一个文件
		lines := 0
		osFile, err := os.Open("regexp-reactnative.ext")
		if err != nil {
			fmt.Println("不存在拓展文件regexp.ext!", err)
			return
		}
		defer osFile.Close()
		reader := bufio.NewReader(osFile)
		for {
			_, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			lines++
		}
		commonRegexp = commonRegexp[0:lines]
	}
	//识别convert配置
	convertArry := strings.SplitN(myConfig.Params.Convert, "[", 2)
	if len(convertArry) != 2 {
		log.Fatalln("convert配置错误!")
	}
	convertArry[1] = strings.Replace(convertArry[1], "]", "", -1)
	ctype := convertArry[0]
	base := strings.Replace(convertArry[1], "upx", "", -1)
	base = strings.Replace(base, "rem", "", -1)
	base = strings.Replace(base, "rpx", "", -1)
	base = strings.Replace(base, "px", "", -1)
	baseFloat, err := strconv.ParseFloat(base, 64)
	if err != nil {
		log.Fatalln("convert基准参数错误!")
	}
	unit := strings.Replace(convertArry[1], base, "", -1)
	Ctype = ctype
	BaseFloat = baseFloat
	Unit = unit
	//识别node节点
	nodes := strings.SplitN(myConfig.Params.Node, "[", 2)
	if len(nodes) != 2 {
		log.Fatalln("node节点配置错误!")
	}
	nodes[1] = strings.Replace(nodes[1], "]", "", -1)
	NodeName = nodes[0]
	NodeNum, err = strconv.Atoi(nodes[1])
	if err != nil {
		log.Fatalln("node节点配置错误!")
	}
	//初始化css抓取点
	for _, v := range strings.SplitN(myConfig.Params.Class, ",", -1) {
		if v != "" {
			matchCssRegexp["default"] = append(matchCssRegexp["default"], fmt.Sprintf(` %s=`, v))
		}
	}
	for _, v := range strings.SplitN(myConfig.Params.CommonClass, ",", -1) {
		if v != "" {
			matchCssRegexp["common"] = append(matchCssRegexp["common"], fmt.Sprintf(` %s=`, v))
		}
	}
}

//查询文件打开文件返回byte
func FindFileToGetByte(path string) (fileBody []byte) {
	file, err := os.OpenFile(path, os.O_RDWR, 0x666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fileBody, err = ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

//查询文件打开文件并处理css
func FindFileToGetCss(path string) {
	file, err := os.OpenFile(path, os.O_RDWR, 0x666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fileBody, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	BtyeToCss(fileBody, path)
}
func BtyeToCss(fileBody []byte, path string) {
	var css = ""

	var defaultCss = map[string]int{}
	var countIndex = 1
	fileBodyStr := string(fileBody)
	for _, v := range matchCssRegexp["default"] {
		re := regexp.MustCompile(fmt.Sprintf(`%s['|"]{1,1}([^'|^"]+)['|"]{1,1}`, v))
		for _, v1 := range re.FindAllStringSubmatch(fileBodyStr, -1) {
			for i, v2 := range v1 {
				if i == 1 {
					v2splits := strings.SplitN(v2, " ", -1)
					for _, v3 := range v2splits {
						if v3 != "" && defaultCss[v3] == 0 {
							defaultCss[v3] = countIndex
							countIndex++
						}
					}
				}
			}
		}
	}
	defaultCssToArray := []string{}
	for i := 1; i < countIndex; i++ {
		for k, v := range defaultCss {
			if v == i {
				defaultCssToArray = append(defaultCssToArray, k)
			}
		}
	}
	defaultCssOutToArray := []string{}
	for i, v := range defaultCssToArray {
		for _, tv := range commonRegexp {
			for k1, v1 := range tv {
				ss1 := k1.FindAllStringSubmatch(v, -1)
				if len(ss1) != 0 {
					var ss2 = v1
					for i2, v2 := range ss1[0] {
						if i2 >= 1 {
							//数值替换
							ss2 = strings.Replace(ss2, fmt.Sprintf("$%d", i2), v2, -1)
						}
					}
					defaultCssOutToArray = append(defaultCssOutToArray, "."+defaultCssToArray[i]+"{"+ss2+"}")
					break
				}
			}
		}
	}
	var commonCssStr = ""
	for _, v := range matchCssRegexp["common"] {
		re := regexp.MustCompile(fmt.Sprintf(`%s['|"]{1,1}([^'|^"]+)['|"]{1,1}`, v))
		for _, v1 := range re.FindAllStringSubmatch(fileBodyStr, -1) {
			for i, v2 := range v1 {
				if i == 1 {
					commonCssStr += v2 + "\n"
				}
			}
		}
	}
	commonCssStr = strings.Replace(commonCssStr, "{", " { ", -1)
	commonCssStr = strings.Replace(commonCssStr, "}", " } ", -1)
	for _, tv := range commonRegexp {
		for k, v := range tv {
			re := regexp.MustCompile(strings.Replace(strings.Replace(k.String(), "^", " ", -1), "$", " ", -1))
			ss1 := re.FindAllStringSubmatch(commonCssStr, -1)
			if len(ss1) != 0 {
				if len(ss1) != 0 {
					var ss2 = v
					for i2, v2 := range ss1[0] {
						if i2 >= 1 {
							//数值替换
							ss2 = strings.Replace(ss2, fmt.Sprintf("$%d", i2), v2, -1)
						}
					}
					commonCssStr = strings.Replace(commonCssStr, ss1[0][0], " "+ss2+" ", -1)
				}
			}
		}
	}
	css = strings.Join(defaultCssOutToArray, "\n") + "\n" + commonCssStr
	css = strings.Replace(css, " { ", "{", -1)
	css = strings.Replace(css, " } ", "}", -1)
	//获取需要比较的css：willCompareCss
	styles := regexp.MustCompile("<style([^>]*)>([^<]*)</style>").FindAllStringSubmatch(fileBodyStr, -1)
	if len(styles)-1 < NodeNum {
		fmt.Println("找不到替换节点: ", path)
		return
	}
	compare := styles[NodeNum]
	// compare   all,attr,css
	if len(compare) != 3 {
		log.Println("节点数量异常: ", path)
	}
	willCompareCss := ""
	cssSplitBefore := ""
	cssSplit := []string{}
	if myConfig.Params.IgnoreSplit == "none" {
		cssSplitBefore = ""
		willCompareCss = compare[2]
	} else {
		cssSplit = strings.SplitN(compare[2], myConfig.Params.IgnoreSplit, -1)
		if len(cssSplit) == 1 {
			cssSplitBefore = ""
			willCompareCss = compare[2]
		} else if len(cssSplit) == 2 {
			cssSplitBefore = cssSplit[0]
			willCompareCss = cssSplit[1]
		}
	}
	css = cssToCover(css, compare)
	switch myConfig.Params.Replace {
	case "node":
		if !isTheSame(css, willCompareCss) { //生成的css和比较的css存在差异
			styleTmp := ""
			if myConfig.Params.IgnoreSplit == "none" {
				styleTmp = fmt.Sprintf(`<style%s>
		%s</style>`, compare[1], css) //插入本身节点需要的模板
			} else {
				styleTmp = fmt.Sprintf(`<style%s>%s%s
%s</style>`, compare[1], cssSplitBefore, myConfig.Params.IgnoreSplit, css) //插入本身节点需要的模板
			}
			fileBodyStr = strings.Replace(fileBodyStr, compare[0], styleTmp, -1)
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
			fmt.Println("find   ", path, "   changed!")
		}
	default:
		splteFilePath := []string{}
		if strings.Contains(path, "/") {
			splteFilePath = strings.SplitN(path, "/", -1)
		} else {
			splteFilePath = strings.SplitN(path, `\`, -1)
		}
		fileName := strings.Replace(splteFilePath[len(splteFilePath)-1], myConfig.Params.Suffix, "", -1)
		if strings.Contains(myConfig.Params.Replace, "write-") {
			willWritePath := filepath.Join(path, fmt.Sprintf(strings.Replace(myConfig.Params.Replace, "write-", "", -1), fileName))
			writeFile, err := os.OpenFile(willWritePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0x666)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer writeFile.Close()
			_, err = writeFile.WriteString(css)
			if err != nil {
				fmt.Println(err)
				return
			}
		} else if strings.Contains(myConfig.Params.Replace, "append-") { //追加只能追加固定文件
			willAppendPath := strings.Replace(myConfig.Params.Replace, "append-", "", -1)
			appendFile, err := os.OpenFile(willAppendPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0x666)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer appendFile.Close()
			_, err = appendFile.WriteString(css)
			if err != nil {
				fmt.Println(err)
				return
			}
		} else {
			log.Fatalln("替代节点报错！")
		}
	}
}

// @color = > #111
func cssGlobalValueToCss(css string) string {
	for key, value := range myConfig.Values {
		css = strings.Replace(css, key, fmt.Sprintf("%v", value), -1)
	}
	return css
}

// width:10px => width:0.1rem
func cssToCover(css string, compare []string) string {
	css = cssGlobalValueToCss(css)
	//转换成程序可识别的单位zzzz
	//转换 已经识别的css：css
	css = strings.Replace(css, "px", "zzzz", -1)
	var unit = ""
	var baseFloat float64 = 0.0 //这个基数是哪里获取
	var err = errors.New("")
	switch Ctype {
	case "auto": //优先attr，再self
		attrUnit := regexp.MustCompile(fmt.Sprintf(`%s="([^"]+)"`, Unit)).FindAllStringSubmatch(compare[1], -1)
		if len(attrUnit) == 0 { //不存在配置，就使用self
			unit = Unit
			baseFloat = BaseFloat
			break
		}
		unit = Unit
		baseFloat, err = strconv.ParseFloat(attrUnit[0][1], 64)
		if err != nil {
			baseFloat = 1.0
			break
		}
	case "self": //强制使用配置文件的单位
		unit = Unit
		baseFloat = BaseFloat
	case "attr": //使用参数节点参数配置
		attrUnit := regexp.MustCompile(fmt.Sprintf(`%s="([^"]+)"`, Unit)).FindAllStringSubmatch(compare[1], -1)
		if len(attrUnit) == 0 { //不存在配置，就是用px，1.0为默认
			unit = "px"
			baseFloat = 1.0
			break
		}
		unit = Unit
		baseFloat, err = strconv.ParseFloat(attrUnit[0][1], 64)
		if err != nil {
			baseFloat = 1.0
			break
		}
	}
	css = regexp.MustCompile(`(\d+)zzzz`).ReplaceAllStringFunc(css, func(s string) string {
		var ts = strings.Replace(s, "zzzz", "", -1)
		ti, err := strconv.Atoi(ts)
		if err != nil {
			return s
		}
		return fmt.Sprintf("%v%s", float64(ti)*baseFloat, unit)
	})
	return css
}
