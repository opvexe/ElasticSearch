#### 1.1 ElasticSearch安装

```shell
$ brew install elasticsearch@5.6
$ brew services start elasticsearch@5.6
# 访问http://localhost:9200
```

#### 1.2 ElasticSearch使用

增：

```json
请求地址：http://localhost:9200/testbookindex/books/testbookindex/book/1
请求头：Content-Type:application/json;charset=UTF-8
请求方式：PUT
请求体:
{
 "book_id":3,
 "book_name":"Java",
 "description":"一门垃圾语言"
}
```

删：

```json
请求地址：http://localhost:9200/testbookindex/books/testbookindex/book/1
请求头：Content-Type:application/json;charset=UTF-8
请求方式：DELETE
```

`模糊查询`——单字段 : ==关键字:match==

```json
请求地址：http://localhost:9200/testbookindex/books/testbookindex/book/_search
请求头：Content-Type:application/json;charset=UTF-8
请求方式：POST
请求体:
{
   "query":{
           "match":{
               "book_name":"java"
           }
   }
}
```

`模糊查询`——多字段 : ==关键字:multi_match==

```json
请求地址：http://localhost:9200/testbookindex/books/testbookindex/book/_search
请求头：Content-Type:application/json;charset=UTF-8
请求方式：POST
请求体:
{
   "query":{
           "multi_match":{
               "query":"语言",
               "fields":["book_name","description"]
           }
   }
}
```

`模糊查询` —— 未指定字段：==关键字:query_string==

```json
请求地址：http://localhost:9200/testbookindex/books/testbookindex/book/_search
请求头：Content-Type:application/json;charset=UTF-8
请求方式：POST
请求体:
{
   "query":{
           "query_string":{
               "query":"ios是个垃圾"
           }
   }
}
```

`模糊查询` —— 指定字段：==关键字:query_string==

```json
请求地址：http://localhost:9200/testbookindex/books/testbookindex/book/_search
请求头：Content-Type:application/json;charset=UTF-8
请求方式：POST
请求体:
{
   "query":{
           "query_string":{
               "query":"ios对象",
               "fields":["book_name","description"]
           }
   }
}
```

指定结构体中字段查询：==关键字:term==

```json
请求地址：http://localhost:9200/testbookindex/books/testbookindex/book/_search
请求头：Content-Type:application/json;charset=UTF-8
请求方式：POST
请求体:
{
    "query" : 
        {
            "term" : {"book_name" : "ios"}
        }
}
```

范围查询: ==关键字：range==

```shell
gte 大于等于
lte  小于等于
gt 大于
lt 小于
now 当前时间
```

```json
请求地址：http://localhost:9200/testbookindex/books/testbookindex/book/_search
请求头：Content-Type:application/json;charset=UTF-8
请求方式：POST
请求体:
{
    "query" : 
        {
            "range" : {
                "create_data" : {
                        "gte":"2018-01-01",
                         "lte":"now"
                         }
                    }
        }
}
```

对数据过滤: ==关键字: bool，filter==

```json
请求地址：http://localhost:9200/testbookindex/books/testbookindex/book/_search
请求头：Content-Type:application/json;charset=UTF-8
请求方式：POST
请求体:
{
    "query" : {
            "bool" : {
                "filter" : {
                        "term":{
                        "book_name":"java"                                }
                            }
                        }
              }
}
```

复合查询 --- ==关键字：should关键词：或的关系==

```json
请求地址：http://localhost:9200/testbookindex/books/testbookindex/book/_search
请求头：Content-Type:application/json;charset=UTF-8
请求方式：POST
请求体:
{
    "query" : {
            "bool" : {
                "should" : [
                            {
                                "match":{"book_name":"java"}
                            },
                            {
                                "match":{"description":"ios"}
                            }
                            ]
                        }
              }
}

```

==must :关键字== 都满足

```json
请求地址：http://localhost:9200/testbookindex/books/testbookindex/book/_search
请求头：Content-Type:application/json;charset=UTF-8
请求方式：POST
请求体:
{
    "query" : {
            "bool" : {
                "must" : [
                            {
                                "match":{"book_name":"java"}
                            },
                            {
                                "match":{"description":"ios"}
                            }
                            ]
                        }
              }
}
```

```json
# book_id ==3 且 book_name = java 且 description = 语言
{
    "query" : {
            "bool" : {
                "must" : [
                            {
                                "match":{"book_name":"java"}
                            },
                            {
                                "match":{"description":"语言"}
                            }
                            ],
                   "filter":[
                            {
                                "term":{
                                    "book_id":3
                                }
                            }    
                            ]
                        }
              }
}
```

