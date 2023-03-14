# Gee

version 0.0.1
实现了路由映射表，提供了用户注册静态路由的方法，包装了启动服务的函数

version 0.0.2
设计Context,针对使用场景，封装*http.Request和http.ResponseWriter的方法，
简化相关接口的调用，只是设计 Context 的原因之一。对于框架来说，还需要支撑额外的功能。
例如，将来解析动态路由/hello/:name，参数:name的值放在哪呢？再比如，框架需要支持中间件，
那中间件产生的信息放在哪呢？Context 随着每一个请求的出现而产生，请求的结束而销毁，
和当前请求强相关的信息都应由 Context 承载。因此，设计 Context 结构，扩展性和复杂性留在了内部，
而对外简化了接口。路由的处理函数，以及将要实现的中间件，参数都统一使用 Context 实例， 
Context 就像一次会话的百宝箱，可以找到任何东西。


version 0.0.3
之前，我们用了一个非常简单的map结构存储了路由表，使用map存储键值对，索引非常高效，但是有一个弊端，键值对的存储的方式，
只能用来索引静态路由。那如果我们想支持类似于/hello/:name这样的动态路由怎么办呢？所谓动态路由，
即一条路由规则可以匹配某一类型而非某一条固定的路由。例如/hello/:name，可以匹配/hello/geektutu、hello/jack等。
动态路由有很多种实现方式，支持的规则、性能等有很大的差异。例如开源的路由实现gorouter支持在路由规则中嵌入正则表达式，
例如/p/[0-9A-Za-z]+，即路径中的参数仅匹配数字和字母；另一个开源实现httprouter就不支持正则表达式。
著名的Web开源框架gin 在早期的版本，并没有实现自己的路由，而是直接使用了httprouter，
后来不知道什么原因，放弃了httprouter，自己实现了一个版本。
实现动态路由最常用的数据结构，被称为前缀树(Trie树)。看到名字你大概也能知道前缀树长啥样了：
每一个节点的所有的子节点都拥有相同的前缀。这种结构非常适用于路由匹配，比如我们定义了如下路由规则：


/:lang/doc

/:lang/tutorial

/:lang/intro

/about

/p/blog

/p/related

![img.png](img.png)
HTTP请求的路径恰好是由/分隔的多段构成的，因此，每一段可以作为前缀树的一个节点。我们通过树结构查询，如果中间某一层的节点都不满足条件，
那么就说明没有匹配到的路由，查询结束。
接下来我们实现的动态路由具备以下两个功能。
参数匹配:。例如 /p/:lang/doc，可以匹配 /p/c/doc 和 /p/go/doc。
通配*。例如 /static/*filepath，可以匹配/static/fav.ico，也可以匹配/static/js/jQuery.js，这种模式常用于静态服务器，能够递归地匹配子路径。

version 0.0.4
分组控制(Group Control)是 Web 框架应提供的基础功能之一。所谓分组，是指路由的分组。
如果没有路由分组，我们需要针对每一个路由进行控制。但是真实的业务场景中，往往某一组路由需要相似的处理。例如：

以/post开头的路由匿名可访问。

以/admin开头的路由需要鉴权。

以/api开头的路由是 RESTful 接口，可以对接第三方平台，需要三方平台鉴权。

大部分情况下的路由分组，是以相同的前缀来区分的。因此，我们今天实现的分组控制也是以前缀来区分，
并且支持分组的嵌套。例如/post是一个分组，/post/a和/post/b可以是该分组下的子分组。
作用在/post分组上的中间件(middleware)，也都会作用在子分组，子分组还可以应用自己特有的中间件。


version 0.0.5


如果我们将用户在映射路由时定义的Handler添加到c.handlers列表中，结果会怎么样呢？想必你已经猜到了。</br>
func A(c *Context) {</br>
&emsp;&emsp;part1</br>
&emsp;&emsp;c.Next()</br>
&emsp;&emsp;part2</br>
}</br>
func B(c *Context) {</br>
&emsp;&emsp;part3</br>
&emsp;&emsp;c.Next()</br>
&emsp;&emsp;part4</br>
}</br>
假设我们应用了中间件 A 和 B，和路由映射的 Handler。c.handlers是这样的[A, B, Handler]，c.index初始化为-1。调用c.Next()，接下来的流程是这样的：

c.index++，c.index 变为 0</br>
0 < 3，调用 c.handlers[0]，即 A</br>
执行 part1，调用 c.Next()</br>
c.index++，c.index 变为 1</br>
1 < 3，调用 c.handlers[1]，即 B</br>
执行 part3，调用 c.Next()</br>
c.index++，c.index 变为 2</br>
2 < 3，调用 c.handlers[2]，即Handler</br>
Handler 调用完毕，返回到 B 中的 part4，执行 part4</br>
part4 执行完毕，返回到 A 中的 part2，执行 part2</br>
part2 执行完毕，结束。</br>

一句话说清楚重点，最终的顺序是part1 -> part3 -> Handler -> part 4 -> part2。恰恰满足了我们对中间件的要求，接下来看调用部分的代码，就能全部串起来了。




version 0.0.6
网页的三剑客，JavaScript、CSS 和 HTML。要做到服务端渲染，第一步便是要支持 JS、CSS 等静态文件。还记得我们之前设计动态路由的时候，
支持通配符*匹配多级子路径。比如路由规则/assets/*filepath，可以匹配/assets/开头的所有的地址。例如/assets/js/geektutu.js，
匹配后，参数filepath就赋值为js/geektutu.js。
那如果我么将所有的静态文件放在/usr/web目录下，那么filepath的值即是该目录下文件的相对地址。映射到真实的文件后，
将文件返回，静态服务器就实现了。
找到文件后，如何返回这一步，net/http库已经实现了。因此，gee 框架要做的，仅仅是解析请求的地址，
映射到服务器上文件的真实地址，交给http.FileServer处理就好了。



version 0.0.7

对一个 Web 框架而言，错误处理机制是非常必要的。可能是框架本身没有完备的测试，
导致在某些情况下出现空指针异常等情况。也有可能用户不正确的参数，触发了某些异常，
例如数组越界，空指针等。如果因为这些原因导致系统宕机，必然是不可接受的。