# bilinovel
## 说明
- 这是一个bilinovel小说源，目前只支持自动打包小说。如果要支持更多的源，可以提PR或者issue，然后考虑添加支持。
- 并且这是个中文源，所以就不扯什么了。
- 凑合用吧，不可用时再修补修补。

## 使用方法
- 首先请确保你已经获取了可用novelpackager工具，并且具有该源。
```
root@u24arm:~# ./novelpackager bilinovel -v
bilinovel version v0.1.0
```
- 检索
```
root@u24arm:~# ./novelpackager bilinovel search 你好
[INFO]|bilinovel|<UTC 2025/04/13 09:08:15> Successfully fetched search list for URL: https://www.bilinovel.com/search.html?searchkey=你好 
 INDEX  ID    NAME                  AUTHOR        METAS                 DESCRIPTION                                   
 1      2336  再见龙生你好人生      永岛ひろあき  [转生 校园 魔法 龙傲  最强最古的龙，厌倦了漫长的生命，故意死在来讨  
                                                  天 后宫 人外 其他文   伐自己的勇者们的手上。 龙本以为这下就可以去到 
                                                  库 连载]              冥府......                                    
 2      3939  你好，世界（HELLO WO  野崎惑        [科幻 青春 恋爱 集英  即使世界毁灭，我也想再见你一面！ 夏日的京都， 
              RLD）                               社 完结]              十六岁的直实遇到了人生的初恋，却因为害羞而不  
                                                                        敢表......                                    
 3      3863  你好，我是前世制造杀  优木凛々      [异世界 恋爱 女性视   子爵小姐克洛伊有前世的记忆。 前世是千年前灭亡 
              戮魔道具的子爵小姐                  角 syosetu 完结]      的国家的头号魔道具师，受国家的命令制造了很多  
                                                                        杀戮......                                    
 4      3712  你好？我是手机，有什  早月やたか    [校园 青春 欢乐向 后  叮咚！「你好，这里是日比谷为明的家吗？」 一天 
              么事？                              宫 人外 富士见文库    、性格阴郁的我收到了生日礼物…那便是号称最新型 
                                                  连载]                 ＆......                                      
 5      3135  你好、我是受心上人所  六つ花えいこ  [奇幻 冒险 恋爱 syos  「想拜托妳制作爱情魔药」 『湖之善魔女』在某天 
              托来做恋爱药的魔女                  etu 完结]             ，从暗恋对象那里被拜托了制作爱情魔药而失恋了  
                                                                        。 ......                                     
 6      2168  天才程式少女 ─Hello   仙波ユウスケ  [冒险 青春 恋爱 青梅  圣诞夜的夜里，打工完回到家的少年·池野朋生，发 
              World─                              竹马 讲谈社 连载]     现一名银发少女倒在自家门前。朋生让自称为丽奈  
                                                                        的她......                                    
 7      1448  你好哇，暗杀者        大泽めぐみ    [青春 恋爱 百合 女性  中萱梓，昵称阿梓。 长相和成绩都不是很起眼，却 
                                                  视角 角川文库 完结]   因为「她好像在搞援交还是卖春什么的」这种流言  
                                                                        ，全......                                    
 TOTAL  7                                                                                                             
```
- 查看信息
```
root@u24arm:~# ./novelpackager bilinovel info 2336
[INFO]|bilinovel|<UTC 2025/04/13 09:09:21> Navigator: map[deviceMemory:4 hardwareConcurrency:8 language:en-US languages:[en-US] platform:Linux x86_64 userAgent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 vendor:Google Inc. webdriver:false] 
[INFO]|bilinovel|<UTC 2025/04/13 09:09:25> Successfully fetched book info for URL: https://www.bilinovel.com/novel/2336.html 再见龙生你好人生 
                                                                                                                       
 Id           2336                                                                                                     
 Name         再见龙生你好人生                                                                                         
 Author       永岛ひろあき,市丸きすけ(插画) 著                                                                         
 Metas        [转生 校园 魔法 龙傲天 后宫 人外 其他文库 日本轻小说]                                                    
 Description  最强最古的龙，厌倦了漫长的生命，故意死在来讨伐自己的勇者们的手上。                                       
              龙本以为这下就可以去到冥府永眠了，但回过神来，却转生成了人类的小孩。                                     
              由龙变为人类，重新体会到活着的快乐的龙，决定作为人继续生活下去。                                         
              生为边境农民孩子的龙，隐藏着灵魂中蕴含的伟大力量过着平淡的生活，但他遇见拉米亚族的少女、黑蔷薇的妖精后， 
              因对魔法的力量产生了兴趣而去魔法学院上学。                                                               
              曾经是龙的这个人类，在魔法学院的生活中，与美丽而又强大的同学们，大地母神和吸血鬼女王、龙族女皇们成为了朋 
              友，体会到了活着的喜悦与幸福。                                                                           
 Volumes      total: 21                                                                                                
              1.  第一卷                                                                                               
              2.  第二卷                                                                                               
              3.  第三卷                                                                                               
              4.  第四卷                                                                                               
              5.  第五卷                                                                                               
              6.  第六卷                                                                                               
              7.  第七卷                                                                                               
              8.  第八卷                                                                                               
              9.  第九卷                                                                                               
              10.  第十卷                                                                                              
              11.  第十一卷                                                                                            
              12.  第十二卷                                                                                            
              13.  第十三卷                                                                                            
              14.  第十四卷                                                                                            
              15.  第十五卷                                                                                            
              16.  第十六卷                                                                                            
              17.  第十七卷                                                                                            
              18.  第十八卷                                                                                            
              19.  第十九卷                                                                                            
              20.  第二十卷                                                                                            
              21.  第二十一卷                                                                                                                                                                                          
```
- 下载
```
root@u24arm:~# ./novelpackager bilinovel download 3712
[WARN]|bilinovel|<UTC 2025/04/13 09:42:29> Failed to load record for book 3712: open bn_3712.np: no such file or directory 
[INFO]|bilinovel|<UTC 2025/04/13 09:42:31> Navigator: map[deviceMemory:4 hardwareConcurrency:8 language:en-US languages:[en-US] platform:Linux x86_64 userAgent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 vendor:Google Inc. webdriver:false] 
[INFO]|bilinovel|<UTC 2025/04/13 09:42:35> Successfully fetched book info for URL: https://www.bilinovel.com/novel/3712.html 你好？我是手机，有什么事？ 
[INFO]|bilinovel|<UTC 2025/04/13 09:42:39> Successfully fetched book info for URL: /novel/3712/vol_189045.html 你好？我是手机，有什么事？ 第一卷 Chapters 9 
[INFO]|bilinovel|<UTC 2025/04/13 09:42:41> Fetching chapter : 你好？我是手机，有什么事？ 第一卷 插图 
[INFO]|bilinovel|<UTC 2025/04/13 09:42:45> Successfully fetched chapter for URL: https://www.bilinovel.com/novel/3712/189046.html 插图 
[INFO]|bilinovel|<UTC 2025/04/13 09:42:48> Fetching chapter : 你好？我是手机，有什么事？ 第一卷 序章 初始设定 
[INFO]|bilinovel|<UTC 2025/04/13 09:42:53> Successfully fetched chapter for URL: https://www.bilinovel.com/novel/3712/189047.html 序章 初始设定 
[INFO]|bilinovel|<UTC 2025/04/13 09:42:56> Fetching chapter : 你好？我是手机，有什么事？ 第一卷 第一章 去获得大家的ID吧 
[INFO]|bilinovel|<UTC 2025/04/13 09:43:11> Successfully fetched chapter for URL: https://www.bilinovel.com/novel/3712/189048.html 第一章 去获得大家的ID吧 
[INFO]|bilinovel|<UTC 2025/04/13 09:43:14> Fetching chapter : 你好？我是手机，有什么事？ 第一卷 第二章 通过互发消息来改善印象 
[INFO]|bilinovel|<UTC 2025/04/13 09:43:33> Successfully fetched chapter for URL: https://www.bilinovel.com/novel/3712/189049.html 第二章 通过互发消息来改善印象 
[INFO]|bilinovel|<UTC 2025/04/13 09:43:36> Fetching chapter : 你好？我是手机，有什么事？ 第一卷 第三章 约会辅助也是信手拈来 
[INFO]|bilinovel|<UTC 2025/04/13 09:43:55> Successfully fetched chapter for URL: https://www.bilinovel.com/novel/3712/189050.html 第三章 约会辅助也是信手拈来 
[INFO]|bilinovel|<UTC 2025/04/13 09:43:58> Fetching chapter : 你好？我是手机，有什么事？ 第一卷 第四章 夏日回忆令人略感难忘 
[INFO]|bilinovel|<UTC 2025/04/13 09:44:18> Successfully fetched chapter for URL: https://www.bilinovel.com/novel/3712/189051.html 第四章 夏日回忆令人略感难忘 
[INFO]|bilinovel|<UTC 2025/04/13 09:44:32> Fetching chapter : 你好？我是手机，有什么事？ 第一卷 第五章 就算没有手机，也可以成为朋友 
[INFO]|bilinovel|<UTC 2025/04/13 09:44:50> Successfully fetched chapter for URL: https://www.bilinovel.com/novel/3712/189052.html 第五章 就算没有手机，也可以成为朋友 
[INFO]|bilinovel|<UTC 2025/04/13 09:44:52> Fetching chapter : 你好？我是手机，有什么事？ 第一卷 终章 恋爱就该抛开顾虑 
[INFO]|bilinovel|<UTC 2025/04/13 09:44:53> Successfully fetched chapter for URL: https://www.bilinovel.com/novel/3712/189053.html 终章 恋爱就该抛开顾虑 
[INFO]|bilinovel|<UTC 2025/04/13 09:44:56> Fetching chapter : 你好？我是手机，有什么事？ 第一卷 后记 
[INFO]|bilinovel|<UTC 2025/04/13 09:44:56> Successfully fetched chapter for URL: https://www.bilinovel.com/novel/3712/189054.html 后记 
[INFO]|bilinovel|<UTC 2025/04/13 09:44:56> Download book %s success 3712
root@u24arm:~# ls
novelpackager  你好？我是手机，有什么事？.epub 
```

## 其他
- 基本用法就是这么简单，没有过多的子命令（因为已经满足我的使用了，如果有其他需要可以提issue或者PR，然后考虑添加支持）
- 虽然只有这几个命令，但一些辅助参数也有不少的作用，比如重试次数、打包方式等等，请自行使用-h进行尝试。
- over.