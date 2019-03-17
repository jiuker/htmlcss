## 背景
* 在这几年工作中针对css的书写和排查错误感受到通用性的css在开发过程中的变得很重要,less/sass等处理css也不是很方便。
* 在工作写了很多类似class带属性值的class比如 w-20 w20等，于是萌生出是否可以用脚本使用正则去处理，经过两年的摸索，有了些许眉目。
## 支持的处理的样式
* 只支持单class，层级class需要特殊处理才能支持。
## 支持平台
* apicloud/小程序/uniapp/vue/传统的网页，RN目前不支持。
## 项目如何使用
* config.ini
## vue/uniapp如何使用
* config.ini的关键配置mode为start/runtime，dir是需要处理目录，suffix是需要处理的文件名，convert为需要转换的比例和值，replace是处理模式为node，ignoreSplit是在单style的情况下作为区分自动处理和不需要处理的样式
* 参考页面结构
<template>
  <view class="d-wb wbo-v h100 w100" c-class="body{w100 h100}">
    
  </view>
</template>
<script>
  
</script>
<style>
body{height:100%;}
/*auto*/
willInsertHere!
</style>
## apicloud/传统页面如何使用
* config.ini的关键配置为mode为none，然后开其端口，里面由于自带了apicloud的调试模块，网页可以打开http://ip:port/debug/html，使用了这个在调试里面加一个cd({asd:"haha"})也会显示在该页面。replace推荐为style[1]，在页面里面使用两个`<style></style>`节点
* 只需要引入http://ip:port/sync.js即可，推荐在公共的js里面写document.write来写入js路径
## 小程序如何使用
* 结构和vue类似，但是replace为 write-../%s.css这种配置。%s为处理的文件名