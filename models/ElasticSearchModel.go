package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//参考文献：【https://www.cnblogs.com/liuxiaoming123/p/8124969.html】

const (
	ConfigElasticSearchTimeOut = 15
	ConfigElasticSearchHost    = "http://localhost:9200/"
	ConfigElasticSearchIndex   = "docBook"
	ConfigElasticSearchType    = "Book"
)

//全文搜索客户端
type ElasticSearchClient struct {
	Host    string        `json:"host"`  //host的地址:http://localhost:9200/
	Index   string        `json:"index"` //索引-->数据库
	Type    string        `json:"type"`  //数据库表
	Timeout time.Duration //超时时间
}

/*
{
  "count": 3,
  "_shards": {
    "total": 5,
    "successful": 5,
    "skipped": 0,
    "failed": 0
  }
}
*/
//统计信息结构
type ElasticSearchCount struct {
	Shards struct {
		Failed     int `json:"failed"`
		Skipped    int `json:"skipped"`
		Successful int `json:"successful"`
		Total      int `json:"total"`
	} `json:"_shards"`
	Count int `json:"count"`
}

// 统计结果结构
type ElasticSearchResult struct {
	TimedOut bool `json:"timed_out"`
	Took     int  `json:"took"`
	//统计信息
	Shards struct {
		Failed     int `json:"failed"`
		Skipped    int `json:"skipped"`
		Successful int `json:"successful"`
		Total      int `json:"total"`
	} `json:"_shards"`
	//命中结构
	Hits struct {
		Total    int         `json:"total"`
		MaxScore interface{} `json:"max_score"`
		//命中结果数组结果
		Hits []struct {
			ID    string      `json:"id"`
			Index string      `json:"_index"`
			Type  string      `json:"_type"`
			Score interface{} `json:"_score"`
			//数据结果结构
			Source struct {
				BookID      int    `json:"book_id"`
				BookName    string `json:"book_name"`
				Description string `json:"description"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

//搜索数据
type ElasticSearchData struct {
	Id          int    `json:"id"`
	BookName    string `json:"book_name"`
	Description string `json:"description"`
}

//创建全文索引【客户端】
func NewElasticSearchClient(index, typ string) (client *ElasticSearchClient) {
	client = &ElasticSearchClient{
		Host:    ConfigElasticSearchHost,
		Index:   index,
		Type:    typ,
		Timeout: time.Duration(ConfigElasticSearchTimeOut) * time.Second,
	}
	client.Host = strings.TrimRight(client.Host, "/") + "/"
	return
}

//检测ElasticSearch服务能否连通
func (this *ElasticSearchClient) ping() error {
	resp, err := this.get(this.Host).Response()
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusMultipleChoices || resp.StatusCode < http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + "；" + string(body))
	}
	return err
}

//创建索引
//http://localhost:9200/docBook/Book/1
/*
{
	"id":1
	"book_name":"Go",
	"description":"Go是一门更高并发语言。"
}
*/
func (this *ElasticSearchClient) BuildIndex(data ElasticSearchData) error {
	api := this.Host + this.Index + "/" + this.Type + "/" + strconv.Itoa(data.Id)
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp, err := this.put(api).Body(js).Response()
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusMultipleChoices || resp.StatusCode < http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(resp.Status + "；" + string(body))
	}
	return nil
}

//删除索引
func (this *ElasticSearchClient) DeleteIndex(id int) (err error) {
	api := this.Host + this.Index + "/" + this.Type + "/" + strconv.Itoa(id)
	if resp, errResp := this.delete(api).Response(); errResp != nil {
		err = errResp
	} else {
		if resp.StatusCode >= http.StatusMultipleChoices || resp.StatusCode < http.StatusOK {
			b, _ := ioutil.ReadAll(resp.Body)
			err = errors.New("删除索引失败：" + resp.Status + "；" + string(b))
		}
	}
	return
}

//检查索引是否存在
func (this *ElasticSearchClient) isExists() (err error) {
	api := this.Host + this.Index
	var resp *http.Response
	if resp, err = this.get(api).Response(); err == nil {
		if resp.StatusCode >= 300 || resp.StatusCode < 200 {
			b, _ := ioutil.ReadAll(resp.Body)
			err = errors.New(resp.Status + "：" + string(b))
		}
	}
	return
}

//查询数量
//http://localhost:9200/docBook/Book/_count
func (this *ElasticSearchClient) Count() (count int, err error) {
	api := this.Host + this.Index + "/" + this.Type + "/_count"
	if resp, errResp := this.get(api).Response(); errResp != nil {
		err = errResp
	} else {
		buff, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode >= http.StatusMultipleChoices || resp.StatusCode < http.StatusOK {
			err = errors.New(resp.Status + "；" + string(buff))
		} else {
			var obj ElasticSearchCount
			if err = json.Unmarshal(buff, &obj); err == nil {
				count = obj.Count
			}
		}
	}
	return
}

//should 或的意思||
//must 与&&


//根据书名或描述查询有关书籍
/*
	{
   "query":{
           "multi_match":{
               "query":"java",
               "fields":["description","book_name"]
           }
	"_source":["book_id"],  //只要book_id
   }
}
*/
func (this *ElasticSearchClient) SearchBooks(field string, size int, page int) (book_id []string,err error) {
	api := this.Host + this.Index + "/" + this.Type + "/_search"
	if page > 0 {
		page = page - 1
	} else {
		page = 0 // 默认是0
	}
	queryBody := `{
  		 "query":{
           		"multi_match":{
              		 "query":%v,
               		"fields":["description","book_name"]
          		 }
   		},
		"_source":["book_id"],
		"size": %v,
		"from": %v
	}`
	queryBody = fmt.Sprintf(queryBody, field,size,page)
	if resp, errResp := this.post(api).Body(queryBody).Response(); errResp != nil {
		err = errResp
	} else {
		buff, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode >= http.StatusMultipleChoices || resp.StatusCode < http.StatusOK {
			err = errors.New(resp.Status + "；" + string(buff))
		} else {
			var obj ElasticSearchResult
			if err = json.Unmarshal(buff, &obj); err == nil {
				for _,v := range obj.Hits.Hits{
					book_id = append(book_id, v.ID)
				}
			}
		}
	}
	return
}

//未指定字段查询
func (this *ElasticSearchClient) SearchByUnSpecified(field string, size int, page int) (book_id []string,err error) {
	api := this.Host + this.Index + "/" + this.Type + "/_search"
	if page > 0 {
		page = page - 1
	} else {
		page = 0 // 默认是0
	}
	queryBody := `{
  		 "query":{
           		"query_string":{
				"query":%v
			}
   		},
		"_source":["book_id"],
		"size": %v,
		"from": %v
	}`
	queryBody = fmt.Sprintf(queryBody, field,size,page)
	if resp, errResp := this.post(api).Body(queryBody).Response(); errResp != nil {
		err = errResp
	} else {
		buff, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode >= http.StatusMultipleChoices || resp.StatusCode < http.StatusOK {
			err = errors.New(resp.Status + "；" + string(buff))
		} else {
			var obj ElasticSearchResult
			if err = json.Unmarshal(buff, &obj); err == nil {
				for _,v := range obj.Hits.Hits{
					book_id = append(book_id, v.ID)
				}
			}
		}
	}
	return
}
/*
	{
    "query" : {
            "bool" : {
                "should" : [
                            {
                                "match":{"name":"springboot"}
                            },
                            {
                                "match":{"country":"中国"}
                            }
                            ]
                        }
              }
}

　2.must:
{
    "query" : {
            "bool" : {
                "must" : [
                            {
                                "match":{"name":"springboot"}
                            },
                            {
                                "match":{"country":"中国"}
                            }
                            ]
                        }
              }
}

3.must filter:
{
    "query" : {
            "bool" : {
                "must" : [
                            {
                                "match":{"name":"springboot"}
                            },
                            {
                                "match":{"country":"中国"}
                            }
                            ],
                   "filter":[
                            {
                                "term":{
                                    "age":20
                                }
                            }
                            ]
                        }
              }
}

4.must_not:
{
    "query" : {
            "bool" : {
                "must_not" : {
                                "term":{"age":20}
                             }
                      }
               }
}




*/

//put请求
func (this *ElasticSearchClient) put(api string) (req *httplib.BeegoHTTPRequest) {
	return httplib.Put(api).Header("Content-Type", "application/json").SetTimeout(this.Timeout, this.Timeout)
}

//post请求
func (this *ElasticSearchClient) post(api string) (req *httplib.BeegoHTTPRequest) {
	return httplib.Post(api).Header("Content-Type", "application/json").SetTimeout(this.Timeout, this.Timeout)
}

//delete请求
func (this *ElasticSearchClient) delete(api string) (req *httplib.BeegoHTTPRequest) {
	return httplib.Delete(api).Header("Content-Type", "application/json").SetTimeout(this.Timeout, this.Timeout)
}

//get请求
func (this *ElasticSearchClient) get(api string) (req *httplib.BeegoHTTPRequest) {
	return httplib.Get(api).Header("Content-Type", "application/json").SetTimeout(this.Timeout, this.Timeout)
}
