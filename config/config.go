package config

import (
	"log"

	"github.com/larspensjo/config"
)

type Conf struct {
	Mode         string
	ServerIpPort string //端口模式下,http 开启js的端口
	Ip           string //端口模式的白名单，最后一定要以逗号结束，过滤是判断是否存在来做的
	Dir          string //文件模式下，需要处理的目录
	Suffix       string //需要处理的文件后缀名,只支持一种模式
	Ignore       string //忽略的文件,可以多个，逗号分开
	Convert      string //单位转换，self强制以本单位为准，attr就是单独的文件配置的为准,auto 是优先文件attr配置。1px就是转换前的单位乘以的值，可以是0.1，px为转换后的单位。
	Replace      string //node为原文件模式,node,ignoreSplit配置生效, write-../%s.css 覆盖模式，append-/xxxx/api.css,追加模式,只能使用固定的文件目录
	Node         string //比较替换的节点，[index],index为第几个这样的节点，如果没有这样的节点就会在控制输出该目录不符合
	IgnoreSplit  string // 单页面的中间切分标志，如果没有则当作全部替换，如果有则指替换标志下面的，none为空
	Class        string //单独的class处理
	CommonClass  string // 公共的class
	React        string //后缀名为js，就是使用这个none为普通js，其他则为对应的应用
	ReactMode    string //生成多个还是单个的区别
}

var Params = Conf{}

func init() {
	cfg, err := config.ReadDefault("config.ini")
	if err != nil {
		log.Fatal(err)
	}
	Params.Mode, err = cfg.String("mode", "mode")
	if err != nil {
		log.Fatalln(err)
	}
	Params.Dir, err = cfg.String("handle", "dir")
	if err != nil {
		log.Fatalln(err)
	}
	Params.Suffix, err = cfg.String("handle", "suffix")
	if err != nil {
		log.Fatalln(err)
	}
	Params.Ignore, err = cfg.String("handle", "ignore")
	if err != nil {
		log.Fatalln(err)
	}
	Params.ServerIpPort, err = cfg.String("handle", "serverIpPort")
	if err != nil {
		log.Fatalln(err)
	}
	Params.Convert, err = cfg.String("handle", "convert")
	if err != nil {
		log.Fatalln(err)
	}
	Params.Node, err = cfg.String("handle", "node")
	if err != nil {
		log.Fatalln(err)
	}
	Params.IgnoreSplit, err = cfg.String("handle", "ignoreSplit")
	if err != nil {
		log.Fatalln(err)
	}
	Params.Replace, err = cfg.String("handle", "replace")
	if err != nil {
		log.Fatalln(err)
	}
	Params.Ip, err = cfg.String("handle", "ip")
	if err != nil {
		log.Fatalln(err)
	}
	Params.Class, err = cfg.String("csshandle", "class")
	if err != nil {
		log.Fatalln(err)
	}
	Params.CommonClass, err = cfg.String("csshandle", "commonCalss")
	if err != nil {
		log.Fatalln(err)
	}
	Params.React, err = cfg.String("handle", "react")
	if err != nil {
		log.Fatalln(err)
	}
	Params.ReactMode, err = cfg.String("handle", "reactmode")
	if err != nil {
		log.Fatalln(err)
	}
}
