package http

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	myConfig "github.com/jiuker/htmlcss/config"

	"github.com/gorilla/websocket"
)

type Ws struct {
	Conn   *websocket.Conn
	Singal chan int //信号，为1就是退出
}

var CommonWs = Ws{}
var Prhtml = "" //send发送的html保存起来

func ListenAndServe() {
	writeLock := sync.Mutex{}
	http.HandleFunc("/debug/html", func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte(`
<html>
<head>
    <meta charset="utf-8">
    <title>Apicloud控制台</title>
    <meta name="viewport" content="width=device-width,initial-scale=1, maximum-scale=1, minimum-scale=1, user-scalable=no">
    <meta content="telephone=no,email=no" name="format-detection">
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="black">
</head>
<style>
    *{
        margin: 0px;
        padding: 0px;
        font-family: "consolas";
		-webkit-text-size-adjust:none;
    }
    body,html{
        width: 100%;
        background-color: #333;
    }
</style>
<style>
.w100{width:100%;}
.h100{height: 100%;}
.d-wb{display: -webkit-box;display: -webkit-flex;display: flex;}
.wbf-1{-webkit-box-flex: 1;-webkit-flex:1;flex: 1;overflow: auto;-webkit-overflow-scrolling: touch;}
.wbo-v{-webkit-box-orient: vertical;-webkit-flex-flow: column;flex-flow:column;}
.h-30{height: 30px;}
.bc-aaa{background-color: #aaa;}
.wb-c{-webkit-box-align: center;-webkit-align-items: center;align-items: center;-webkit-box-pack: center; -webkit-justify-content: center; justify-content: center;}
.p-1010{padding:10px 10px;}
.fs-14{font-size:14px;}
.c-fff{color:#fff;}
.wbf-3{-webkit-box-flex: 3;-webkit-flex:3;flex: 3;overflow: auto;-webkit-overflow-scrolling: touch;}
.bl-1-eee{border-left:1px solid #eee;}
.p-r{position: relative;}
.p-a{position: absolute;}
.r-0{right:0;}
.t-0{top:0;}
.w-360{width:360px;}
.h-720{height: 720px;}
.c-10ff00{color:#10ff00;}
.c-f00{color:#f00;}

</style>
<body >
    <div class="w100 h100 d-wb " id="bVue">
        <div class="wbf-3 d-wb wbo-v bl-1-eee p-r">
            <div @click="retData1=[]" class="h-30 bc-aaa d-wb wb-c">
                debug
            </div>
			<div class="p-a r-0 t-0">
				<iframe class="w-360 h-720" v-bind:src="urlPath"></iframe>
			</div>
            <div class="wbf-1" id="dom1">
                <pre v-for="ret in retData1" v-html="ret" class="p-1010 fs-14 c-10ff00">
                    
                </pre>
            </div>
        </div>
        <div class="wbf-1 d-wb wbo-v bl-1-eee">
            <div @click="retData2=[]" class="h-30 bc-aaa d-wb wb-c">
                error
            </div>
            <div class="wbf-1" id="dom2">
                <pre v-for="ret in retData2" v-html="ret" class="p-1010 fs-14 c-f00">
                    
                </pre>
            </div>
        </div>
    </div>
</body>
<script src="https://cdn.jsdelivr.net/npm/vue/dist/vue.js"></script>
<script type="text/javascript">
    var v= new Vue({
        el:"#bVue",
        data:{
            retData1:[],//debug
            retData2:[],//error
			urlPath:"",
        },
        watch:{
            retData1:function(nl,ol){
				if(nl.length>20){
					this.retData1 = this.retData1.splice(-20,20);
				}
                this.$nextTick(function(){
                    dom1.scrollTop=dom1.scrollHeight
                });
            },
            retData2:function(nl,ol){
				if(nl.length>20){
					this.retData2 = this.retData2.splice(-20,20);
				}
                this.$nextTick(function(){
                    dom2.scrollTop=dom2.scrollHeight
                });
            }
        }
    });
    window.onload=function(){
        wsF();
    };
	function getYmdTime(time){
		return time.getHours()+":"+time.getMinutes()+":"+time.getSeconds();
	}
    function wsF(){
        var ws = new WebSocket("ws://` + myConfig.Params.ServerIpPort + `/debug/console");
        ws.addEventListener("message",function(evt){
            var data = JSON.parse(evt.data);
            //type:"",data:{}
            if(data.type=="debug"){
                v.retData1.push('[--'+getYmdTime(new Date())+'--] '+data.data)
            }
            if(data.type=="error"){
                v.retData2.push('[--'+getYmdTime(new Date())+'--] '+data.data)
            }
			if(data.type=="urlChageData"){
                v.urlPath="http://` + myConfig.Params.ServerIpPort + `/debug/phoneData?time="+(new Date()).getTime(); 
            }
        });
        ws.addEventListener("error",function(evt){
            setTimeout(function(){
                wsF();
            },1000)
        });
    }
</script>
</html>
        `))
	})
	http.HandleFunc("/debug/phoneData", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
		w.Header().Add("content-type", "text/html")
		var willreturn = Prhtml[31 : len(Prhtml)-2]
		willreturn = fmt.Sprintf(`
		<html style="font-size:10px;">
		<script type="text/javascript">
			document.querySelector("html").innerHTML = decodeURIComponent("%s")
		</script>
		</html>`, willreturn)
		w.Write([]byte(willreturn))
	})
	http.HandleFunc("/debug/console", func(w http.ResponseWriter, r *http.Request) {
		ws, err := websocket.Upgrade(w, r, nil, 1024*1024, 1024*1024)
		if err != nil {
			fmt.Println(err)
			w.Write([]byte(`error`))
			return
		}
		defer ws.Close()
		CommonWs.Conn = ws
		for {
			select {
			case singal := (<-CommonWs.Singal):
				if singal == 1 {
					//退出
					break
				}
			}
		}
		w.Write([]byte(`done`))
	})
	http.HandleFunc("/debug/send", func(w http.ResponseWriter, r *http.Request) {
		writeLock.Lock()
		defer writeLock.Unlock()
		w.Header().Set("Access-Control-Allow-Origin", "*") //允许访问所有域
		bodyContent, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Write([]byte(`error`))
			return
		}
		if strings.Contains(string(bodyContent), "urlChageData") {
			Prhtml = string(bodyContent)
		}
		if len(bodyContent) != 0 && CommonWs.Conn != nil {
			err = CommonWs.Conn.WriteMessage(websocket.TextMessage, bodyContent)
			if err != nil {
				CommonWs.Singal <- 1
				w.Write([]byte(`{status:false}`))
			} else {
				w.Write([]byte(`{status:true}`))
			}
		} else {
			w.Write([]byte(`{status:false}`))
		}
	})
	http.HandleFunc("/sync.js", func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
		resp.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
		resp.Header().Add("content-type", "application/javascript")
		if strings.Contains(myConfig.Params.Ip, strings.SplitN(req.RemoteAddr, ":", -1)[0]+",") {
			resp.Write([]byte(SyncJs))
		} else {
			resp.Write([]byte(`"..."`))
		}

	})
	//为apicloud平台添加文件目录
	h := http.FileServer(http.Dir(myConfig.Params.Dir))
	dirName := filepath.Base(myConfig.Params.Dir)
	http.Handle(fmt.Sprintf("/%s/", dirName), http.StripPrefix(fmt.Sprintf("/%s/", dirName), h))
	err := http.ListenAndServe(myConfig.Params.ServerIpPort, nil)
	if err != nil {
		log.Fatalln(err)
	}
}
func init() {
	osFile, err := os.Open("regexp.ext")
	if err != nil {
		SyncJs = strings.Replace(SyncJs, "insertHere", "", 1)
		fmt.Println("不存在拓展文件regexp.ext!", err)
		return
	}
	defer osFile.Close()
	extendRegexpStr := ""
	tpl := `this.regexps.push({
            rp: new RegExp(/^%s$/),
            rep: "%s"
        });`
	reader := bufio.NewReader(osFile)
	_, err = reader.Peek(1)
	if err != nil {
		SyncJs = strings.Replace(SyncJs, "insertHere", "", 1)
		fmt.Println("不存在拓展文件regexp.ext!", err)
		return
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		ls := strings.SplitN(line, "->", -1)
		if len(ls) != 2 {
			fmt.Println("拓展错误!", ls)
			continue
		}
		extendRegexpStr += fmt.Sprintf(tpl, ls[0], ls[1][0:len(ls[1])-2])
	}
	SyncJs = strings.Replace(SyncJs, "insertHere", extendRegexpStr, 1)
}

var SyncJs = `if(true) { //debug js
    //server address
    window.DebugServer = 'http://` + myConfig.Params.ServerIpPort + `/` + filepath.Base(myConfig.Params.Dir) + `/';
    function c() {
        var willConsole = "";
        for(var i = 0; i < arguments.length; i++) {
            if(typeof(arguments[i]) == "object") {
                willConsole += JSON.stringify(arguments[i], null, 4);
            } else {
                willConsole += (arguments[i] + "");
            }
        }
        console.log(willConsole)
        try{
            throw new Error()
        }catch(e){
            console.log(e.stack)
        }
    }
     function cw() {
        var willConsole = "";
        for(var i = 0; i < arguments.length; i++) {
            if(typeof(arguments[i]) == "object") {
                willConsole += JSON.stringify(arguments[i], null, 4);
            } else {
                willConsole += (arguments[i] + "");
            }
        }
        console.warn(willConsole)
    }
    function ce() {
        var willConsole = "";
        for(var i = 0; i < arguments.length; i++) {
            if(typeof(arguments[i]) == "object") {
                willConsole += JSON.stringify(arguments[i], null, 4);
            } else {
                willConsole += (arguments[i] + "");
            }
        }
		ajaxDebugData('error',willConsole)
    }
	function cd(){
		var willConsole = "";
        for(var i = 0; i < arguments.length; i++) {
            if(typeof(arguments[i]) == "object") {
                willConsole += JSON.stringify(arguments[i], null, 4);
            } else {
                willConsole += (arguments[i] + "");
            }
        }
		ajaxDebugData('debug',willConsole)
	}
    function ajaxDebugData(type_d,data){
		var obj=new XMLHttpRequest();
        obj.open("POST","http://` + myConfig.Params.ServerIpPort + `/debug/send",true);
        obj.onreadystatechange = function() {
            if (obj.readyState == 4 && obj.status == 200 || obj.status == 304) { // readyState == 4说明请求已完成
               
            }
        };
        obj.send(JSON.stringify({type:type_d,data:data}));
	}
    window.alert = c;
    window.__defineSetter__("apiready", function(v) {
        window._apiready = v;
    });
    var apireadyTime = 0;
    function l(target,key,tname){
        var showName = ''
        if(typeof tname =='string'){
            showName = tname
        }else{
            throw new Error("没有对象名");
        }
        for(var i = 0;i<key.length;i++){
            var x = key[i];
            target[x] = (function(_Func,value){
                if(typeof _Func == 'function') {
                    return function() {
                        var _params = arguments;
                        var haveCallBack = false;
                        for(var i = 0; i < _params.length; i++) {
                            if(typeof _params[i] == 'function') {
                                haveCallBack = true;
                                var _ai = _params[i];
                                _params[i] = function() {
                                    cw(showName+'.'+value+'----------\r\n',_params, arguments);
                                    try{
                                        throw new Error()
                                    }catch(e){
                                        cw(e.stack)
                                    }
                                    _ai.apply(this, arguments);
                                };
                            }
                        }
                        if(!haveCallBack) {
                            cw(showName+'.'+value+'----------\r\n',_params);
                            try{
                                throw new Error()
                            }catch(e){
                                cw(e.stack)
                            }
                        }
                        return _Func.apply(this, _params);
                    }
                } else {
                    cw(showName+'.'+value+'----------\r\n',_Func);
                    try{
                        throw new Error()
                    }catch(e){
                        cw(e.stack)
                    }
                    return _Func
                }
            })(target[x],x);
        }   
    }
    window.__defineGetter__("apiready", function() {
        if(apireadyTime != 0) {
			
        } else {
            l(api,['ajax', 'openWin','openFrame','openFrameGroup','execScript','addEventListener','sendEvent'],'api');
            apireadyTime++;
        };
		api.addEventListener({
            name: 'longpress'
        }, function(ret, err) {
			document.querySelector("head").insertAdjacentHTML("afterBegin",'<base href="'+window.location.href+'" />')
            ajaxDebugData('urlChageData',encodeURIComponent(document.querySelector("html").innerHTML))
        });
        if(api.winName == "root") {
            if(api.frameName==''){
                api.addEventListener({
                    name: 'longpress'
                }, function(ret, err) {
                    api.rebootApp();
                });
            }
        }
        if(location.href.indexOf('file://') != -1) {
            return function() {
                location.href = location.href.replace(/.+A\d+\//, DebugServer)
            }
        } else {
            return window._apiready
        }
    });
	setTimeout(function(){
		if(typeof api == "undefined"){
			(vueReady||function(){})()
		}
	},500)
};
/*
 * 页面上只要出现 class=""就会被该js抓取，这个抓取的原理是获取网页实现的
 * (空格)class="" c-class="asd{asd{p-0010 p-10}}"类似的都会被争取处理
 * */
var styleFirst = document.querySelector("style");
if(styleFirst != null) {
    styleFirst.innerHTML = styleFirst.innerHTML + "@media screen and (max-width: 640px){::-webkit-scrollbar{width: 0px;height:0px;}}"
};
function HTMLCSS(){
    var _this=this;
    _this._html="";
    _this.classList=[];
    _this.commonClassList=[];
    _this.regexps=[];
    _this.output=[];
    _this.init=function(url){
        var styles=document.querySelectorAll("style");
        if(styles.length>=2){

        }else{
            alert("没有符合插入css的条件,个数不足！");
            return
        }
        _this.initHtml(url,function(){
            _this.initRegExp();
            _this.initClassArray();
            _this.classArrayToOutput();
            _this.initCommonClassArray();
            _this.commonClassArrayToOutput();
            switch (_this.findStyleElementToInsert(_this.output.join("").replace(/}/g,"}\r\n").replace(/\n /g,""))){
                case 0:
                    alert("没有符合插入css的条件");
                    break;
                case 1:
					window.addEventListener("keypress",function(e){ //c复制
					    if(e.keyCode==99){
							if(typeof HTMLCOPY !='function'){
								var content = "没有内容";
					            var aux = document.createElement("textarea");
					            // 获取复制内容
					            aux.value = content;
					            // 将元素插入页面进行调用
					            document.body.appendChild(aux);
					            // 复制内容
					            aux.select();
					            document.execCommand("selectAll");
					            // 将内容复制到剪贴板
					            document.execCommand("cut");
					            //    // 删除创建元素
					            document.body.removeChild(aux);
							}else{
								HTMLCOPY();
							}
					    }
					});
                    window.HTMLCOPY = function(){
                        var content = document.querySelectorAll("style")[1].innerHTML;
                        var aux = document.createElement("textarea");
                        // 获取复制内容
                        aux.value = content;
                        // 将元素插入页面进行调用
                        document.body.appendChild(aux);
                        // 复制内容
                        aux.select();
                        document.execCommand("selectAll");
                        // 将内容复制到剪贴板
                        document.execCommand("cut");
                        //    // 删除创建元素
                        document.body.removeChild(aux);
                    }
                    document.querySelector('body').insertAdjacentHTML('afterBegin','<div" style="width: 40px;height: 40px;background-color: rgba(255,0,0,0.5);position: fixed;left: 0;top: 0;z-index:111111111;"></div>')
                    console.log(window.location.href+"->>>>>>"+"need fixed!")
                    break;
                case 2:
                    console.log(window.location.href+"->>>>>>"+"dont need fixed!")
                    break;
                default:
                    break;
            }
        });
    },
    /*
     *根据url来初始化需要显示的页面
     */
    _this.initHtml=function(url,cb){
		var obj=new XMLHttpRequest();
        obj.open("GET",url,true);
        obj.onreadystatechange = function() {
            if (obj.readyState == 4 && obj.status == 200 || obj.status == 304) { // readyState == 4说明请求已完成
               _this._html+=obj.responseText+" ";
               cb();
            }
        };
        obj.send();
    };
    /*
     *初始化cssClass数组
     *第一种 class="any class",第二种 c-class="any class"
     */
    _this.initClassArray=function(){
        /*
         *handle class="any class"
         */
        //match class="any class",
        // map 为es5映射，对map里面的进行修改拷贝数组也会修改。
        //forEach 为es5遍历，对里面的进行修改不会对元素进行修改。
        //filter 为es5过滤筛选
        var classMatch = _this._html.match(/ class="([^"]+)"/g)||[];
        classMatch = classMatch.map(function(v,i){
          return v.replace(' class="',"").replace('"',"")
        });
        classMatch.forEach(function(v,i){
            var splitBySingalClass = v.split(/ /g).filter(function(v1,i1){
                if (v1==""){
                    return false;
                }else{
                    return true;
                }
            });
            splitBySingalClass.forEach(function(v1,i1,arr){
                if (_this.classList.indexOf(v1)==-1){
                    _this.classList.push(v1)
                }
            })
        });
    };
    /*
     * 正常的class会过滤掉不存在的属性。
     * */
    _this.classArrayToOutput=function(){
        _this.output=_this.classList.map(function(v,i){
            try{
                _this.regexps.forEach(function(v1,i1){
                    if(v.match(v1.rp) != null) {
                        throw "." + v + "{" + v.replace(v1.rp, v1.rep) + "}"
                    }
                })
            }catch(e){
                return e
            }
            return v
        }).filter(function(v,i){
            return v.indexOf("{")!=-1;
        });
    };
    /*
     * common性质的class
     * */
    _this.initCommonClassArray=function(){
        var classMatch = _this._html.match(/c-class="([^"]+)"/g)||[];
        classMatch = classMatch.map(function(v,i){
          return v.replace('c-class="',"").replace('"',"")
        });
        _this.commonClassList=classMatch;
    };
    /*
     把common css输入到ouput
     * */
    _this.commonClassArrayToOutput=function(){
        var data=_this.commonClassList.map(function(v,i){
            /*
             to find base{} string incloud '{}'
             * */
            var baseCssClass="";
            var start=false;
            for(var i=0;i<v.length;i++){
                if(v[i]=="{"){
                    baseCssClass=""
                    start=true;
                }
                if(start){
                    baseCssClass+=v[i]
                }
                if(v[i]=="}"){
                    start=false;
                    var base=baseCssClass.replace(/{|}/g,"").split(/ /g).filter(function(v1,i1){
                        return v1!=""
                    }).map(function(v2p,i2p){
                        //去除掉原来的table
                        return v2p.replace(/\t/g,"")
                    }).map(function(v2,i2){
//                      console.log(v2)
                        try{
                            _this.regexps.forEach(function(v3,i3){
                                if(v2.match(v3.rp) != null) {
//                                  console.log("{"+v2.replace(v3.rp, v3.rep)+"}")
                                    throw  "{"+v2.replace(v3.rp, v3.rep)+"}" 
                                }
                            })
                        }catch(e){
                            return e
                        }
                        return v2
                    }).filter(function(v4,i4){
                        //把无用数组过滤掉
                        return v4.indexOf("{")!=-1;
                    }).map(function(v5,i5){
                        //把数组改成可渲染的数组
                        return v5.replace(/{|}/g,"")
                    }).join("");
                    v=v.replace(baseCssClass,"{"+base+"}")
                }
            }
            return v
        });
        this.output=_this.output.concat(data);
    }
    _this.findStyleElementToInsert=function(css){
        try{
            _this._html.match(/<style[^<]+<\/style>/g).forEach(function(v,i){
                    if (v.match(/<style[^>]*>/)[0].match(/px="/)!=null){
                        //这是px
                        var baseNum=parseFloat(v.match(/<style[^>]*>/)[0].match(/px="([^"]+)"/g)[0].replace(/px="|"]/g,""));
                        css=css.replace(/\d{1,10}px/g,function(a){
                            var d=parseInt(a.replace("px",""));
                                if (d==0){
                                    return '0';
                                }
                                if(d<=2){
                                    return d+"px";
                                }
                                return d/baseNum+"px"
                            });
                        if (!_this.findStyleElementIsTheSame(v,css)){
                            document.querySelectorAll("style")[i].innerHTML=css;
                            throw 1
                        }else{
                            throw 2
                        }
                    }
                    if (v.match(/<style[^>]*>/)[0].match(/rem="/)!=null){
                        //这是rem
                            var baseNum=parseFloat(v.match(/<style[^>]*>/)[0].match(/rem="([^"]+)"/g)[0].replace(/[rem="|"]/g,""));
                            css=css.replace(/\d{1,10}px/g,function(a){
                            var d=parseFloat(a.replace("px",""));
                                if (d==0){
                                    return '0';
                                }
                                //这个小于等于1就默认不转,防止1px的时候会产生意外
								 if(d<=1){
									 return d+"px";
								 }
                                return d/baseNum+"rem"
                            });
                        if (!_this.findStyleElementIsTheSame(v,css)){
                            document.querySelectorAll("style")[i].innerHTML=css;
                            throw 1
                        }else{
                            throw 2
                        }
                    }
            })
        }catch(e){
            return e
        }
        //没有找到可以插入的style，就默认插入第二个
        if (_this._html.match(/<style[^<]+<\/style>/g).length==2){
            var v=_this._html.match(/<style[^<]+<\/style>/g)[1];
            var baseNum=1;
            css=css.replace(/\d{1,10}px/g,function(a){
                var d=parseInt(a.replace("px",""));
                    if (d==0){
                        return '0';
                    }
                    if(d<=2){
                        return d+"px";
                    }
                    return d/baseNum+"px"
                });
            if (!_this.findStyleElementIsTheSame(v,css)){
                document.querySelectorAll("style")[1].innerHTML=css;
                return 1
            }else{
                return 2
            }
        }
        return 0
    }
    _this.findStyleElementIsTheSame=function(old,css){
        old=old.replace(/<style[^>]*/,"").replace(/<\/style>/,"")
        var data='0123456789-qwertyuiopasdfghjkl:;zxcvbnm,#"'
        for(var i = 0; i < data.length; i++) {
            var r = eval("/" + data[i] + "/g")
            if((old.match(r) || []).length != (css.match(r) || []).length) {
                return false
            }
        }
        return true;
    }
}
setTimeout(function(){
    var htmlcssObj=new HTMLCSS();
    htmlcssObj.init(window.location.href);
});
HTMLCSS.prototype.initRegExp=function(){
		insertHere
        this.regexps.push({
            rp: new RegExp(/^o-(\d{3,3})(\d{3,3})(\d{3,3})(\d{2,2})-(\d{1,2})$/),
            rep: "outline:rgba($1,$2,$3,0.$4) solid  $5px;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbs-(\d{2,2})(\d{2,2})(\d{2,2})(\d{2,2})-([^-]{3,6})$/),
            rep: "-webkit-box-shadow:$1px $2px $3px $4px #$5;box-shadow:$1px $2px $3px $4px #$5;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbs-(\d{2,2})(\d{2,2})(\d{2,2})(\d{2,2})-(\d{3,3})(\d{3,3})(\d{3,3})(\d{2,2})$/),
            rep: "-webkit-box-shadow:$1px $2px $3px $4px rgba($5,$6,$7,0.$8);box-shadow:$1px $2px $3px $4px rgba($5,$6,$7,0.$8);"
        })
        this.regexps.push({
            rp: new RegExp(/^b-([^-]{1,30})-([^-]{3,6})-(\d{3,3})-([^-]{3,6})-(\d{3,3})-([^-]{3,6})-(\d{3,3})$/),
            rep: "background: -webkit-linear-gradient($1,#$2 $3%,#$4 $5%,#$6 $7%);background: linear-gradient($1,#$2 $3%,#$4 $5%,#$6 $7%);"
        })
        this.regexps.push({
            rp: new RegExp(/^b-([^-]{1,30})-([^-]{3,6})-([^-]{3,6})$/),
            rep: "background: -webkit-linear-gradient($1,#$2,#$3);background: linear-gradient($1,#$2,#$3);"
        })
        this.regexps.push({
            rp: new RegExp(/^h-(\d{0,3})$/),
            rep: "height: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^h-(\d{0,3})-v$/),
            rep: "height: $1vmin;"
        })
        this.regexps.push({
            rp: new RegExp(/^h(\d{0,3})$/),
            rep: "height: $1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^bc-([^-]{3,6})$/),
            rep: "background-color: #$1;"
        })
        this.regexps.push({
            rp: new RegExp(/^boc-([^-]{3,6})$/),
            rep: "border-color: #$1;"
        })
        this.regexps.push({
            rp: new RegExp(/^bc-(\d{3,3})(\d{3,3})(\d{3,3})(\d{2,2})$/),
            rep: "background-color: rgba($1,$2,$3,0.$4);"
        })
        this.regexps.push({
            rp: new RegExp(/^d-wb$/),
            rep: "display: -webkit-box;display: -webkit-flex;display: flex;"
        })
        this.regexps.push({
            rp: new RegExp(/^d-wbox$/),
            rep: "display: -webkit-box;display: box;"
        })
        this.regexps.push({
            rp: new RegExp(/^d-ib$/),
            rep: "display: inline-block;"
        })
        this.regexps.push({
            rp: new RegExp(/^w(\d{0,3})$/),
            rep: "width:$1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^w-(\d{0,3})$/),
            rep: "width:$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^w-(\d{0,3})-v$/),
            rep: "width:$1vmin;"
        })
        this.regexps.push({
            rp: new RegExp(/^fs-(\d{0,3})$/),
            rep: "font-size:$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^fw-([^-]{3,6})$/),
            rep: "font-weight:$1;"
        })
        this.regexps.push({
            rp: new RegExp(/^c-([^-]{3,6})$/),
            rep: "color:#$1;"
        })
        this.regexps.push({
            rp: new RegExp(/^c-(\d{3,3})(\d{3,3})(\d{3,3})(\d{2,2})$/),
            rep: "color:rgba($1,$2,$3,0.$4);"
        })
        this.regexps.push({
            rp: new RegExp(/^br-(\d{1,3})$/),
            rep: "border-radius:$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^br-(\d{1,2})(\d{1,2})(\d{1,2})(\d{1,2})$/),
            rep: "border-radius:$1px $2px $3px $4px;"
        })
        this.regexps.push({
            rp: new RegExp(/^br(\d{1,3})$/),
            rep: "border-radius:$1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^wa-n$/),
            rep: "-webkit-appearance: none;"
        })
        this.regexps.push({
            rp: new RegExp(/^ml-(.{1,3})$/),
            rep: "margin-left:$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^mr-(.{1,3})$/),
            rep: "margin-right:$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^mt-(.{1,3})$/),
            rep: "margin-top:$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^mb-(.{1,3})$/),
            rep: "margin-bottom:$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^mb([^-]{1,3})$/),
            rep: "margin-bottom:$1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^ml([^-]{1,3})$/),
            rep: "margin-left:$1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^mr([^-]{1,3})$/),
            rep: "margin-right:$1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^mt([^-]{1,3})$/),
            rep: "margin-top:$1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^m-(\d{2,2})(\d{2,2})$/),
            rep: "margin:$1px $2px;"
        })
        this.regexps.push({
            rp: new RegExp(/^m-(\d{2,2})(\d{2,2})(\d{2,2})(\d{2,2})$/),
            rep: "margin:$1px $2px $3px $4px;"
        })
        this.regexps.push({
            rp: new RegExp(/^m(\d{2,2})(\d{2,2})$/),
            rep: "margin: $1% $2%;"
        })
        this.regexps.push({
            rp: new RegExp(/^m(\d{2,2})(\d{2,2})(\d{2,2})(\d{2,2})$/),
            rep: "margin: $1% $2% %3 %4;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbo-v$/),
            rep: "-webkit-box-orient: vertical;-webkit-flex-flow: column;flex-flow:column;"
        })
        this.regexps.push({
            rp: new RegExp(/^d-n$/),
            rep: "display: none;"
        })
        this.regexps.push({
            rp: new RegExp(/^va-m$/),
            rep: "vertical-align: middle;"
        })
        this.regexps.push({
            rp: new RegExp(/^va-t$/),
            rep: "vertical-align: top;"
        })
        this.regexps.push({
            rp: new RegExp(/^wb-c$/),
            rep: "-webkit-box-align: center;-webkit-align-items: center;align-items: center;-webkit-box-pack: center; -webkit-justify-content: center; justify-content: center;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbf-(\d{1,2})$/),
            rep: "-webkit-box-flex: $1;-webkit-flex:$1;flex: $1;overflow: auto;-webkit-overflow-scrolling: touch;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbf-(\d{1,2})n$/),
            rep: "-webkit-box-flex: $1;-webkit-flex:$1;flex: $1;"
        })
        this.regexps.push({
            rp: new RegExp(/^o-e$/),
            rep: "overflow: hidden;white-space: nowrap;text-overflow:ellipsis;"
        })
        this.regexps.push({
            rp: new RegExp(/^wo-(\d{2,2})$/),
            rep: "-webkit-opacity:0.$1;opacity:0.$1;"
        })
        this.regexps.push({
            rp: new RegExp(/^ws-n$/),
            rep: "white-space: nowrap;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbp-e$/),
            rep: "-webkit-box-pack: end;-webkit-justify-content: flex-end;justify-content: flex-end;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbp-c$/),
            rep: "-webkit-box-pack: center;-webkit-justify-content: center;justify-content: center;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbp-s$/),
            rep: "-webkit-box-pack: start;-webkit-justify-content: flex-start;justify-content: flex-start;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbp-sb$/),
            rep: "-webkit-box-pack: justify;-webkit-justify-content: space-between;justify-content: space-between;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbp-sa$/),
            rep: "-webkit-box-pack: justify;-webkit-justify-content: space-around;justify-content: space-around;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbp-j$/),
            rep: "-webkit-box-pack: -webkit-justify;-webkit-box-pack:justify;"
        })
        this.regexps.push({
            rp: new RegExp(/^wfw-w$/),
            rep: "-webkit-flex-wrap:wrap;flex-wrap: wrap;"
        })
        this.regexps.push({
            rp: new RegExp(/^wfw-n$/),
            rep: "-webkit-flex-wrap:nowrap;flex-wrap: nowrap;"
        })
        this.regexps.push({
            rp: new RegExp(/^ta-s$/),
            rep: "text-align:start;"
        })
        this.regexps.push({
            rp: new RegExp(/^ta-c$/),
            rep: "text-align:center;"
        })
        this.regexps.push({
            rp: new RegExp(/^ta-e$/),
            rep: "text-align:end;"
        })
        this.regexps.push({
            rp: new RegExp(/^ta-j$/),
            rep: "text-align: justify;"
        })
        this.regexps.push({
            rp: new RegExp(/^o-a$/),
            rep: "overflow: auto;-webkit-overflow-scrolling: touch;"
        })
        this.regexps.push({
            rp: new RegExp(/^o-an$/),
            rep: "overflow: auto;"
        })
        this.regexps.push({
            rp: new RegExp(/^ox-a$/),
            rep: "overflow-x: auto;"
        })
        this.regexps.push({
            rp: new RegExp(/^oy-a$/),
            rep: "overflow-y: auto;"
        })
        this.regexps.push({
            rp: new RegExp(/^o-s$/),
            rep: "overflow: scroll;-webkit-overflow-scrolling: touch;"
        })
        this.regexps.push({
            rp: new RegExp(/^o-sn$/),
            rep: "overflow: scroll;"
        })
        this.regexps.push({
            rp: new RegExp(/^ox-s$/),
            rep: "overflow-x: scroll;"
        })
        this.regexps.push({
            rp: new RegExp(/^oy-s$/),
            rep: "overflow-y: scroll;"
        })
        this.regexps.push({
            rp: new RegExp(/^bt-(\d{1,2})-([^-]{3,6})$/),
            rep: "border-top:$1px solid #$2;"
        })
        this.regexps.push({
            rp: new RegExp(/^bb-(\d{1,2})-([^-]{3,6})$/),
            rep: "border-bottom:$1px solid #$2;"
        })
        this.regexps.push({
            rp: new RegExp(/^br-(\d{1,2})-([^-]{3,6})$/),
            rep: "border-right:$1px solid #$2;"
        })
        this.regexps.push({
            rp: new RegExp(/^bl-(\d{1,2})-([^-]{3,6})$/),
            rep: "border-left:$1px solid #$2;"
        })
        this.regexps.push({
            rp: new RegExp(/^b-(\d{1,2})-([^-]{3,6})$/),
            rep: "border:$1px solid #$2;"
        })
        
        this.regexps.push({
            rp: new RegExp(/^bt-(\d{1,2})-([^-]{3,6})-d$/),
            rep: "border-top:$1px dashed #$2;"
        })
        this.regexps.push({
            rp: new RegExp(/^bb-(\d{1,2})-([^-]{3,6})-d$/),
            rep: "border-bottom:$1px dashed #$2;"
        })
        this.regexps.push({
            rp: new RegExp(/^br-(\d{1,2})-([^-]{3,6})-d$/),
            rep: "border-right:$1px dashed #$2;"
        })
        this.regexps.push({
            rp: new RegExp(/^bl-(\d{1,2})-([^-]{3,6})-d$/),
            rep: "border-left:$1px dashed #$2;"
        })
        this.regexps.push({
            rp: new RegExp(/^b-(\d{1,2})-([^-]{3,6})-d$/),
            rep: "border:$1px dashed #$2;"
        })
        this.regexps.push({
            rp: new RegExp(/^d-b$/),
            rep: "display:block;"
        })
        this.regexps.push({
            rp: new RegExp(/^d-i$/),
            rep: "display:inline;"
        })
        this.regexps.push({
            rp: new RegExp(/^lh-(\d{1,3})$/),
            rep: "line-height: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^lh-n$/),
            rep: "line-height: normal;"
        })
        this.regexps.push({
            rp: new RegExp(/^wba-s$/),
            rep: "-webkit-box-align: start;-webkit-align-items: flex-start;align-items: flex-start;"
        })
        this.regexps.push({
            rp: new RegExp(/^wba-c$/),
            rep: "-webkit-box-align: center;-webkit-align-items: center;align-items: center;"
        })
        this.regexps.push({
            rp: new RegExp(/^wba-e$/),
            rep: "-webkit-box-align: end;-webkit-align-items: flex-end;align-items: flex-end;"
        })
        this.regexps.push({
            rp: new RegExp(/^wfd-rr$/),
            rep: "-webkit-flex-direction:row-reverse;;flex-direction:row-reverse;"
        })
        this.regexps.push({
            rp: new RegExp(/^wba-b$/),
            rep: "-webkit-box-align: baseline ;box-align: baseline ;"
        })
        this.regexps.push({
            rp: new RegExp(/^wba-sc$/),
            rep: "-webkit-box-align: stretch;box-align: stretch;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbl-s$/),
            rep: "-webkit-box-lines: single;box-lines: single;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbog-(\d{1,2})$/),
            rep: "-webkit-box-ordinal-group: $1;box-ordinal-group: $1;"
        })
        this.regexps.push({
            rp: new RegExp(/^wbl-m$/),
            rep: "-webkit-box-lines: multiple;box-lines: multiple;"
        })
        this.regexps.push({
            rp: new RegExp(/^o-h$/),
            rep: "overflow: hidden;"
        })
        this.regexps.push({
            rp: new RegExp(/^ox-h$/),
            rep: "overflow-x: hidden;"
        })
        this.regexps.push({
            rp: new RegExp(/^oy-h$/),
            rep: "overflow-y: hidden;"
        })
        this.regexps.push({
            rp: new RegExp(/^wb-ba$/),
            rep: "word-break: break-all;"
        })
        this.regexps.push({
            rp: new RegExp(/^o-n$/),
            rep: "outline: none;"
        })
        this.regexps.push({
            rp: new RegExp(/^bw-(\d{1,2})$/),
            rep: "border-width: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^p-a$/),
            rep: "position: absolute;"
        })
        this.regexps.push({
            rp: new RegExp(/^p-r$/),
            rep: "position: relative;"
        })
        this.regexps.push({
            rp: new RegExp(/^p-(\d{2,2})(\d{2,2})$/),
            rep: "padding:$1px $2px;"
        })
        this.regexps.push({
            rp: new RegExp(/^p-(\d{2,2})(\d{2,2})(\d{2,2})(\d{2,2})$/),
            rep: "padding:$1px $2px $3px $4px;"
        })
        this.regexps.push({
            rp: new RegExp(/^p(\d{2,2})(\d{2,2})(\d{2,2})(\d{2,2})$/),
            rep: "padding:$1% $2% $3% $4%;"
        })
        this.regexps.push({
            rp: new RegExp(/^pt-(\d{1,3})$/),
            rep: "padding-top: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^pl-(\d{1,3})$/),
            rep: "padding-left: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^pr-(\d{1,3})$/),
            rep: "padding-right: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^pb-(\d{1,3})$/),
            rep: "padding-bottom: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^pt(\d{1,3})$/),
            rep: "padding-top: $1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^pl(\d{1,3})$/),
            rep: "padding-left: $1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^pr(\d{1,3})$/),
            rep: "padding-right: $1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^pb(\d{1,3})$/),
            rep: "padding-bottom: $1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^p(\d{2,2})(\d{2,2})$/),
            rep: "padding: $1% $2%;"
        })
        this.regexps.push({
            rp: new RegExp(/^mw-(\d{1,4})$/),
            rep: "max-width: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^mw(\d{1,4})$/),
            rep: "max-width: $1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^miw-(\d{1,4})$/),
            rep: "min-width: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^miw(\d{1,4})$/),
            rep: "min-width: $1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^mih-(\d{1,4})$/),
            rep: "min-height: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^mih(\d{1,4})$/),
            rep: "min-height: $1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^zi-(\d{1,4})$/),
            rep: "z-index:$1;"
        })
        this.regexps.push({
            rp: new RegExp(/^f-l$/),
            rep: "float:left;"
        })
        this.regexps.push({
            rp: new RegExp(/^f-r$/),
            rep: "float:right;"
        })
        this.regexps.push({
            rp: new RegExp(/^l-(\d{1,3})$/),
            rep: "left:$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^l--(\d{1,3})$/),
            rep: "left:-$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^t-(\d{1,3})$/),
            rep: "top:$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^t--(\d{1,3})$/),
            rep: "top:-$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^r-(\d{1,3})$/),
            rep: "right:$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^r--(\d{1,3})$/),
            rep: "right:-$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^b-(\d{1,3})$/),
            rep: "bottom:$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^b--(\d{1,3})$/),
            rep: "bottom:-$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^l(\d{1,3})$/),
            rep: "left:$1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^t(\d{1,3})$/),
            rep: "top:$1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^r(\d{1,3})$/),
            rep: "right:$1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^b(\d{1,3})$/),
            rep: "bottom:$1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^ws-(\d{1,3})$/),
            rep: "word-spacing:$1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^td-u$/),
            rep: "text-decoration: underline;"
        })
        this.regexps.push({
            rp: new RegExp(/^td-lt$/),
            rep: "text-decoration:line-through;"
        })
        this.regexps.push({
            rp: new RegExp(/^ti-(\d{1,3})$/),
            rep: "text-indent: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^mh-(\d{1,3})$/),
            rep: "max-height: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^mh(\d{1,3})$/),
            rep: "max-height: $1%;"
        })
        this.regexps.push({
            rp: new RegExp(/^p-f$/),
            rep: "position: fixed;"
        })
        this.regexps.push({
            rp: new RegExp(/^p-s$/),
            rep: "position:-webkit-sticky;position:sticky"
        })
        this.regexps.push({
            rp: new RegExp(/^wlc-(\d{1,3})$/),
            rep: "display: -webkit-box;overflow: hidden;text-overflow: ellipsis;word-wrap: break-word;white-space: normal !important;-webkit-line-clamp: $1;-webkit-box-orient: vertical;"
        })
        this.regexps.push({
            rp: new RegExp(/^bs-bb$/),
            rep: "box-sizing:border-box;-moz-box-sizing:border-box;-webkit-box-sizing:border-box;"
        })
        this.regexps.push({
            rp: new RegExp(/^ts-(.{3,3})(.{3,3})(.{3,3})-([^-]{3,6})$/),
            rep: "webkit-text-shadow: $1px $2px $3px #$4;text-shadow: $1px $2px $3px #$4;"
        })
        this.regexps.push({
            rp: new RegExp(/^m-a0$/),
            rep: "margin:auto 0;"
        })
        this.regexps.push({
            rp: new RegExp(/^m-0a$/),
            rep: "margin: 0 auto;"
        })
        this.regexps.push({
            rp: new RegExp(/^m-a$/),
            rep: "margin:auto;"
        })
        this.regexps.push({
            rp: new RegExp(/^c-b$/),
            rep: "clear: both;"
        })
        this.regexps.push({
            rp: new RegExp(/^ls-(.{1,4})$/),
            rep: "letter-spacing: $1px;"
        })
        this.regexps.push({
            rp: new RegExp(/^r-b$/),
            rep: "resize:both;"
        })
        this.regexps.push({
            rp: new RegExp(/^wt-s(\d{2,2})(\d{2,2})$/),
            rep: "transform: scale($1,$2);-webkit-transform: scale($1,$2);"
        })
        this.regexps.push({
            rp: new RegExp(/^wt-s(\d{2,2})(\d{2,2})-0$/),
            rep: "transform: scale(0.$1,0.$2);-webkit-transform: scale(0.$1,0.$2);"
        })
        this.regexps.push({
            rp: new RegExp(/^wt-sX(\d{2,2})-0$/),
            rep: "transform: scaleX(0.$1);-webkit-transform: scaleX(0.$1);"
        })
        this.regexps.push({
            rp: new RegExp(/^wt-sY(\d{2,2})-0$/),
            rep: "transform: scaleY(0.$1);-webkit-transform: scaleY(0.$1);"
        })
        this.regexps.push({
            rp: new RegExp(/^wt-sZ(\d{2,2})-0$/),
            rep: "transform: scaleZ(0.$1);-webkit-transform: scaleZ(0.$1);"
        })
        this.regexps.push({
            rp: new RegExp(/^wt-rZ(\d{3,3})$/),
            rep: "transform: rotateZ($1deg);-webkit-transform: rotateZ($1deg);"
        })
        this.regexps.push({
            rp: new RegExp(/^wt-rY(\d{3,3})$/),
            rep: "transform: rotateY($1deg);-webkit-transform: rotateY($1deg);"
        })
        this.regexps.push({
            rp: new RegExp(/^wt-rX(\d{3,3})$/),
            rep: "transform: rotateX($1deg);-webkit-transform: rotateX($1deg);"
        })
        this.regexps.push({
            rp: new RegExp(/^wt-r(.{1,4})$/),
            rep: "-webkit-transform: rotate($1deg);transform: rotate($1deg);"
        })
        this.regexps.push({
            rp: new RegExp(/^wos-t$/),
            rep: "-webkit-overflow-scrolling: touch;"
        })
        this.regexps.push({
            rp: new RegExp(/^wos-n$/),
            rep: "-webkit-overflow-scrolling: none;"
        })
        this.regexps.push({
            rp: new RegExp(/^bs-c$/),
            rep: "background-size:cover;background-repeat:no-repeat;background-position:center;background-repeat:none;"
        })
        //com 这个是组件，代表特性。com - 组件名-...属性
        this.regexps.push({
            rp: new RegExp(/^com-td-(\d{1,2})-(\d{1,2})-([^-]{3,6})$/),  //宽，高，颜色
            rep: "width:0;height:0;border-width:$2px $1px 0 $1px ;border-style:solid;border-color:#$3 transparent transparent transparent;"
        });
        this.regexps.push({
            rp: new RegExp(/^com-td-(\d{1,2})-(\d{1,2})-(\d{3,3})(\d{3,3})(\d{3,3})(\d{2,2})$/), //宽，高，颜色
            rep: "width:0;height:0;border-width:$2px $1px 0 $1px ;border-style:solid;border-color:rbga($3,$4,$5,$6) transparent transparent transparent;"
        })
        this.regexps.push({
            rp: new RegExp(/^com-tr-(\d{1,2})-(\d{1,2})-([^-]{3,6})$/),  //宽，高，颜色
            rep: "width:0;height:0;border-width:$1px 0 $1px $2px ;border-style:solid;border-color: transparent transparent transparent #$3;"
        });
        this.regexps.push({
            rp: new RegExp(/^com-tr-(\d{1,2})-(\d{1,2})-(\d{3,3})(\d{3,3})(\d{3,3})(\d{2,2})$/), //宽，高，颜色
            rep: "width:0;height:0;border-width:$1px 0 $1px $2px ;border-style:solid;border-color: transparent transparent transparent rbga($3,$4,$5,$6);"
        })
        this.regexps.push({
            rp: new RegExp(/^com-tl-(\d{1,2})-(\d{1,2})-([^-]{3,6})$/),  //宽，高，颜色
            rep: "width:0;height:0;border-width:$1px $2px $1px 0 ;border-style:solid;border-color: transparent #$3 transparent transparent;"
        });
        this.regexps.push({
            rp: new RegExp(/^com-tl-(\d{1,2})-(\d{1,2})-(\d{3,3})(\d{3,3})(\d{3,3})(\d{2,2})$/), //宽，高，颜色
            rep: "width:0;height:0;border-width:$1px $2px $1px 0 ;border-style:solid;border-color: transparent rbga($3,$4,$5,$6) transparent transparent;"
        })
        this.regexps.push({
            rp: new RegExp(/^com-tt-(\d{1,2})-(\d{1,2})-([^-]{3,6})$/),  //宽，高，颜色
            rep: "width:0;height:0;border-width: 0 $1px $2px $1px ;border-style:solid;border-color: transparent  transparent #$3 transparent;"
        });
        this.regexps.push({
            rp: new RegExp(/^com-tt-(\d{1,2})-(\d{1,2})-(\d{3,3})(\d{3,3})(\d{3,3})(\d{2,2})$/), //宽，高，颜色
            rep: "width:0;height:0;border-width:0 $1px $2px $1px  ;border-style:solid;border-color: transparent  transparent rbga($3,$4,$5,$6) transparent;"
        })
}

`
