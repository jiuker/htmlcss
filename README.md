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
* config.ini的配置mode为start/runtime，推荐runtime
* dir是需要处理目录。项目的目录，或者装有文件的目录
* suffix是需要处理的文件名。看是.vue还是其他。
* convert为需要转换的比例和值。uniapp里面推荐 auto[2upx],vue推荐 auto[1px]
* replace是处理模式。推荐为node.
* ignoreSplit是在单style的情况下作为区分自动处理和不需要处理的样式。推荐为/*auto*/ 
* 参考页面结构
```<template>
  <view class="d-wb wbo-v h100 w100" c-class="body{w100 h100}">
    
  </view>
</template>
<script>
  
</script px="100">
<style>
body{height:100%;}
/*auto*/
willInsertHere!
</style> or
</script px="100">
<style>
body{height:100%;}
/*auto*/
willInsertHere!
</style>
```
## apicloud/传统页面如何使用
* config.ini的配置,推荐mode为none，
* serverIpPort为ip和端口，最好不要写localhost:80
* replace推荐为style[1]，在页面里面使用两个`<style></style>`节点,表示前一个为公共的，后一个为自动生成的。
* dir需要配置为项目目录，如目录名为  hello，dir的配置就需要到 绝对地址到 hello。
* 只需要引入`http://ip:port/sync.js`即可，推荐在公共的js里面写document.write来写入js路径
* 里面由于自带了apicloud的调试模块，网页可以打开`http://ip:port/debug/html`，使用了这个在调试里面加一个cd({asd:"haha"})也会显示在该页面。
## 小程序如何使用
* 结构和vue类似，但是replace为 write-../%s.wcss这种配置。%s为当前程序处理的文件名
## 参数的简单解释
### 单位转换
* 在配置文件里面convert=auto[2px]，那么 style节点加一个参数为 px="4"，那么实际以4px进行换算。配置里面单位是什么，单个文件里面的属性为这个才会生效。
* 如果是self就会强制使用配置的单位和比例。
* attr为只以页面配置为准，如果页面没有，就会以1p来处理。
### commonCalss公共class
* 作用只是替换里面的class属性，比如 c-class="body{h100 w100}",工具就会识别为body{width:100%;height:100%;}
### dir需要处理的目录
* 当为start/runtime/always时，这个目录就是需要处理文件目录，也就是文件模式。当为none的时候，就是开启的静态资源库，如目录为 c:/asd/asd/test,就可以通过ip:port/test访问该文件下的东西。
### class的体替换规则原理和对比表
* 利用的正则表达式去生成的class，然后通过一系列的操作达到减轻输入css的数量。
* 在http.go里面，有一个syncjs的定义，最下面的正则表达式就是对比表，当您需要什么参数就去看，但是一般是简写，本工具不提倡使用的原有规则来。regexp.ext优先覆盖内置规则。
* 在regexp.ext里面遵守约定就可以实现自己的正则class。
## 如何配置regexp.ext
* 里面有实例，只支持这种类型的单class的形式，单位一定以px来。编辑好您自己的正则class就可以启动运用，开始愉快的撸撸码了。
* 如 test3-(\d{1,3})-(\d{1,3})->test-$1;test-$2;  规则以->区分为两处，前一处为正则class，后一处为输出，{}会自动添加上。
* 注意事项：最后的替换表达式不能换行，最后一处必定有回车，目前没做兼容处理，后期会考虑的。
# next
* 重构部分逻辑，让维护不会那么困难（就是写得丑了）。
* 增加拓展功能，让写拓展也是一种乐趣，比如增加动画联动，一句代码搞定什么的 myTransitton-10-20 -> @-webkit-transframe{0%{top:10px;}}。
* 对react的支持，或者出一个react版本。