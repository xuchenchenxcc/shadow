package main

import (
"net"
"log"
"fmt"
"strconv"
"encoding/base64"
//"bytes"
//"encoding/binary"
)
type Cache struct{
	clientnet net.Conn
	num int
	content string
}
type map_cache map[string]*Cache

var number int=0
//流量牵引--xcc
func main() {
	// log.SetFlags(log.LstdFlags|log.Lshortfile)
	//监听端口
	var cache = map_cache{}

	l, err := net.Listen("tcp", ":18081")

	//var num *int
	

	if err != nil {
		log.Panic(err)
	}
	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		//fmt.Printf("%s\n",client)
		go handle_request(client, cache)
	}
}
//delete(map1,key1)

func handle_request(client net.Conn, mpcache map_cache) {
	defer client.Close()

	number=number + 1

	//所有连接全部转发到deep learning服务器上
	
	deep_address := "172.16.183.1:8098"
	deep_server, err := net.Dial("tcp", deep_address)
	if err != nil {
		log.Printf("-----HHHHH---handlerequest1--:%s\n", err)
		return
	}
	buffer := write_to_cache(client,mpcache,number)

	deep_copy(deep_server,client,number,buffer)

	//接受deep learning服务器发来的flag，若为true，可疑的，转发的shadowserver上，若为假，转发到路由器本身
	deep_flag:=get_deep_flag(number,deep_server)

	
	if deep_flag==true {
		shadow_address := "172.16.183.5:80"
		shadow_server, err := net.Dial("tcp", shadow_address)
		if err != nil {
			log.Printf("-----HHHHH---handlerequest2---:%s\n", err)
			return
		}
		//是可疑流量
		shadow_copy(mpcache,number,shadow_server)
		
	} else{
		//不是可疑流量
		router_address:="127.0.0.1:80"
		router_server, err := net.Dial("tcp", router_address)
		if err != nil {
			log.Printf("-----HHHHH---handlerequest3----:%s\n", err)
			return
		}
		router_copy(mpcache,number,router_server)
	}
	/*
	var buffer [1024]byte
	for { 
		n, err := client.Read(buffer[:])
		if err != nil {
			log.Printf("-----HHHHH-----:%s\n", err)
			return
		}
		write_to_cache(client)
	}
	*/
	//go io.Copy(client, server)
	//io.Copy(server, client)
}

func write_to_cache(client net.Conn,mpcache map_cache,cache_num int)[]byte{
	//var buffer []byte
	var buffer = make([]byte, 1024)
	_, err := client.Read(buffer[:])
	log.Printf("-----readread-write to cache---")
	if err != nil {
		log.Printf("-----CCCCC0-writetocache----:%s\n", err)
	}

	encodeString := base64.StdEncoding.EncodeToString(buffer)
	num2str:="cache"+strconv.Itoa(cache_num)
	fmt.Println("write to chache num2str is ",num2str)
	mpcache[num2str]=&Cache{clientnet:client,num:cache_num,content:encodeString}
	return buffer
}
func get_deep_flag( num int,deep_flag_server net.Conn )bool{
	//根据number向deepserver获取flag
	/*
	deep_flag_address := "192.168.1.1:80"
	deep_flag_server, err := net.Dial("tcp", deep_address)
	if err != nil {
		log.Printf("-----HHHHH-----:%s\n", err)
		return
	}
*/	
	//num->byte
	//**********上传number到deep server
	/*
	var i1 int64 = int64(num)// [00000000 00000000 ... 00000000 11111111] = [0 0 0 0 0 0 0 255]

    s1 := make([]byte, 0)
    buf := bytes.NewBuffer(s1)

    // 数字转 []byte, 网络字节序为大端字节序
    binary.Write(buf, binary.BigEndian, i1)
    //fmt.Println(buf.Bytes())
    byte_num:=buf.Bytes()

    //将num上传到deepserver
	_, err := deep_flag_server.Write(byte_num[:])
	if err != nil {
		//log.Println(err)
		log.Printf("-----CCCCC1---getdeepflag--:%s\n", err)

	}
	*/
	//deepserver根据num返回flag
	var buffer = make([]byte, 1024)
	_, err := deep_flag_server.Read(buffer[:])
	if err != nil {
		log.Printf("-----CCCCC0----getdeepflag-:%s\n", err)
	}
	//s2 := buffer // [0 0 0 0 0 0 1 255]
	fmt.Println("buffer flag is :",buffer[0])
    /*
    buf:= bytes.NewBuffer(s2)
    var i2 int64
    binary.Read(buf, binary.BigEndian, &i2)
    //打印
    fmt.Println("flag is :",i2)
    */
    //deepflag若为0则为正常流量，其他是可疑流量

    if buffer[0]==48{
    	fmt.Println("flag is 48")
    	return false
    } else{
    	return true
    }
}
//将client的内容copy到deep_server上
func deep_copy(deep_server net.Conn, client net.Conn,num int,buffer []byte) {
	//var buffer [1024]byte
	//把num上传到deep服务器,num->string->byte
	num_str := strconv.Itoa(num)
	num_str="number"+num_str
	data2 := []byte(num_str)

	_, err := deep_server.Write(buffer[:])
	//test
	//fmt.Println(buffer)
	if err != nil {
		//log.Println(err)
		log.Printf("-----CCCCC1---deepcopy--:%s\n", err)
		return
	}
	_, err = deep_server.Write(data2[:])
	//test
	//fmt.Println(data2)
	if err != nil {
		//log.Println(err)
		log.Printf("-----CCCCC2--deepcopy---:%s\n", err)
		return
	}

}
func router_copy(cache map_cache,num int,router_server net.Conn) {
	num2str:="cache"+strconv.Itoa(num)
	encoding_content:=cache[num2str].content
	decodeBytes, err := base64.StdEncoding.DecodeString(encoding_content)
    if err != nil {
        log.Fatalln(err)
    }
	_, err = router_server.Write(decodeBytes[:])
	if err != nil {
		//log.Println(err)
		log.Printf("-----CCCCC1--routercopy---:%s\n", err)
		return
	}	
	var buffer = make([]byte, 1024)
	_, err = router_server.Read(buffer[:])
	if err != nil {
		log.Printf("-----CCCCC0----getdeepflag-:%s\n", err)
	}
	con := cache[num2str].clientnet
	_, err = con.Write(decodeBytes[:])
	if err != nil {
		//log.Println(err)
		log.Printf("-----CCCCC1--routercopy---:%s\n", err)
		return
	}	
	//read
} 
//把number对应的cache中content内容先解码再发送到shadowserver上
func shadow_copy(cache map_cache,num int,shadow_server net.Conn) {
	num2str:="cache"+strconv.Itoa(num)
	encoding_content:=cache[num2str].content
	decodeBytes, err := base64.StdEncoding.DecodeString(encoding_content)

	fmt.Println(decodeBytes[:])

    if err != nil {
        log.Fatalln(err)
    }
	_, err = shadow_server.Write(decodeBytes[:])
	if err != nil {
		//log.Println(err)
		log.Printf("-----CCCCC1---shadowcopy--:%s\n", err)
		return
	}

	var buffer = make([]byte, 1024)
	_, err = shadow_server.Read(buffer[:])
	if err != nil {
		log.Printf("-----CCCCC0----getdeepflag-:%s\n", err)
	}
	con := cache[num2str].clientnet
	_, err = con.Write(decodeBytes[:])
	if err != nil {
		//log.Println(err)
		log.Printf("-----CCCCC1--routercopy---:%s\n", err)
		return
	}



}