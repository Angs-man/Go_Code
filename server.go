// pro2 project main.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-pg/pg"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
)

//常量
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "root"
	dbname   = "test1"
)

type pgData struct {
	Id            int64    `json:"id"`
	Alert_type    string   `json:"alert_type"`
	Line_id       int      `json:"line_id"`
	Line_group_id int      `json:"line_group_id"`
	Content       string   `json:"content"`
	Read          int16    `json:"read"`
	Severity      int16    `json:"severity"`
	Alert_time    int      `json:"alert_time"`
	Create_time   int      `json:"create_time"`
	Proto_class   int16    `json:"proto_class"`
	Si_id         int      `json:"si_id"`
	Site_id       int      `json:"site_id"`
	Ips_id        int      `json:"ips_id"`
	tableName     struct{} `sql:"alert_common"`
}

type outData struct {
	Alert_type    string `json:"alert_type"`
	Line_id       int    `json:"line_id"`
	Line_group_id int    `json:"line_group_id"`
	Severity      int16  `json:"severity"`
	Si_id         int    `json:"si_id"`
	Site_id       int    `json:"site_id"`
}
type postOut struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
type output struct {
	postOut
	Data interface{} `json:"data"`
}

//httprouter写登录、写入client函数
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello,%s!\n", ps.ByName("name"))
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)
	router.GET("/test1/alertCommon", getData)
	router.POST("/test1/alertCommon", uploadData)

	log.Fatal(http.ListenAndServe(":8200", router))
}

//connet函数ORM
func connet() *pg.DB {
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "root",
		Database: "test1",
	})
	var n int
	_, err := db.QueryOne(pg.Scan(&n), "SELECT 1")
	if err != nil {
		panic(err)
	}
	return db
}

//request body json format
func My_json(demo interface{}) *bytes.Buffer {
	if bs, err := json.Marshal(demo); err == nil {
		req := bytes.NewBuffer([]byte(bs))
		return req
	} else {
		panic(err)
	}
}

//插入数据ORM框架
func InsertData(db *pg.DB, body []byte) postOut {
	var a pgData
	var out postOut
	err := json.Unmarshal(body, &a)
	if err != nil {
		out.Code = 10001
		out.Msg = "参数错误"
		return out
	}
	err = db.Insert(&a)
	if err != nil {
		fmt.Println(err)
		out.Code = 10002
		out.Msg = "入库成功"
		return out
	} else {
		out.Code = 10000
		out.Msg = "成功"
		return out
	}
}

//uploadData
func uploadData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, _ := ioutil.ReadAll(r.Body)
	//打开数据库
	db := connet()
	if db == nil {
		fmt.Println("fail to open pgsql")
		return
	}
	ret := InsertData(db, body)
	defer db.Close()
	fmt.Fprint(w, My_json(ret))
}

//getData
func getData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var out output
	var demo pgData
	var outdata outData
	//打开数据库
	db := connet()
	if db == nil {
		fmt.Println("open pgsql default")
		return
	}
	var flags interface{}
	flags = nil
	for i := 1; flags == nil; i++ {
		_, flags = db.QueryOne(&demo, `select * from alert_common where id=?`, i)

		fmt.Println(flags)
		if demo.Id != 0 && flags == nil {
			out.Code = 10000
			out.Msg = "成功"
			outdata.Alert_type = demo.Alert_type
			outdata.Line_group_id = demo.Line_group_id
			outdata.Line_id = demo.Line_id
			outdata.Severity = demo.Severity
			outdata.Site_id = demo.Site_id
			outdata.Si_id = demo.Si_id
			out.Data = outdata

			req := My_json(out)
			fmt.Fprint(w, req)
		} else if i == 1 {
			out.Code = 10003
			out.Msg = "没有数据"

			req := My_json(out)
			fmt.Fprint(w, req)

		}

	}
	defer db.Close()
}
