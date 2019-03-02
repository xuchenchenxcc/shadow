package main

import (
"net"
"log"
"time"
"io"
"strings"
"fmt"
"regexp"
)

//流量牵引--xcc
func main() {
	// log.SetFlags(log.LstdFlags|log.Lshortfile)
	attack_dict := make(map[string]bool)
	crack_time := time.Now()
	l, err := net.Listen("tcp", ":18081")
	if err != nil {
		log.Panic(err)
	}
	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		//fmt.Printf("%s\n",client)
		go handle_request(client, attack_dict, &crack_time)
	}

}

func detect_attack(buffer [1024]byte, attack_dict map[string]bool, crack_time *time.Time) (flag bool, real bool, session string) {
	str := string(buffer[:])
	index := strings.Index(str,"sysauth=")
	crack_index := strings.Index(str, "luci_password")
	if crack_index >= 0 {
		now := time.Now()
		if now.Sub(*crack_time) < 250 * time.Millisecond {
			log.Print("Crack Detected")
			if index >= 0 {
				index += 8
				end := index + 32
				session := str[index:end]
				return true, true, session
			} else {
				return true, true, ""
			}
		}
		*crack_time = now
	}
	if index >= 0 {
		index += 8
		end := index + 32
		session := str[index:end]
		if attack_dict[session] == true {
			log.Println("Session memory OK!")
			return true, true, session
		} else {
			if _re_(str) {
				attack_dict[session] = true
				return true, false, session
			} else {
				return false, false, session
			}
		}
		return false, false, session
	}
	return false,false ,""
}

func _re_(str string)(flag bool) {
	match, _ := regexp.MatchString(`select\b|insert\b|update\b|drop\b|delete\b|dumpfile\b|outfile\b|load_file|rename\b|floor\(|extractvalue|updatexml|name_const|multipoint\(|base64_decode|eval\(|assert\(`, str)
	return match
	index := strings.Index(str, "shell")
	if index >= 0 {
		return true
	} else {
		return false
	}
}



func handle_request(client net.Conn, attack_dict map[string]bool, crack_time *time.Time) {

	// const buffersize int = 1024
	if client == nil {
		return
	}
	defer client.Close()
	var address string
	address = "192.168.1.5:80"
	//for test
	//address = "127.0.0.1:18082"
	server, err := net.Dial("tcp", address)

	if err != nil {
		log.Printf("-----HHHHH-----:%s\n", err)
		return
	}

	var attack_flag bool = false
	attack_flag_point := &attack_flag
	shadow_address := "192.168.1.6:80"
	//for test
	//shadow_address = "127.0.0.1:18083"

	timeout := 2 * time.Second
	server_timeout := 4 * time.Second
	var s_server net.Conn
	var shadow_server *net.Conn = &s_server
	server.SetReadDeadline(time.Now().Add(server_timeout))
	client.SetReadDeadline((time.Now().Add(timeout)))

	go io.Copy(client, server)
	//go s_copy(server, client, attack_dict, attack_flag_point, shadow_address, shadow_server, crack_time)
	c_copy(client, server, attack_dict, attack_flag_point,shadow_address, shadow_server,crack_time)

}
// 参数中的关于attack的没有用到，说不定以后会用到
//func s_copy(server net.Conn, client net.Conn, attack_dict map[string]bool,
//    attack_flag_point *bool,shadow_address string, shadow_server *net.Conn,crack_time *time.Time) { //还得写超时机制
//    var buffer [1024]byte
//    for {
//        if (*attack_flag_point) {
//            server.Close()
//            return
//        } else {
//            n, err := server.Read(buffer[:])
//            if err != nil {
//                //log.Println(err)
//                log.Printf("-----SSSSS1-----:%s\n",err)
//                if n > 0 {
//                    client.Write(buffer[:])
//                }
//                return
//            }
//            _,err = client.Write(buffer[:])
//            if err!= nil {
//                //log.Println(err)
//                log.Printf("-----SSSSS2-----:%s\n",err)
//                return
//            }
//
//        }
//    }
//}

func c_copy(client net.Conn, server net.Conn, attack_dict map[string]bool,
	attack_flag_point *bool,shadow_address string, shadow_server *net.Conn,crack_time *time.Time) { //还得写超时机制
	var buffer [1024]byte
	magic_hack := "POST /cgi-bin/luci/ HTTP/1.1\r\nHost: 192.168.123.1\r\nUser-Agent: Mozilla/5.0 (Macintosh;Intel Mac OS X 10.13; rv:57.0) Gecko/20100101 Firefox/57.0\r\nAccept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\nAccept-Language: en-US,en;q=0.5\r\nReferer: http://192.168.123.1/cgi-bin/luci/\r\nContent-Type: application/x-www-form-urlencoded\r\nContent-Length: 42\r\nCookie: sysauth="
	magic_hack_end := "Connection: close\r\nUpgrade-Insecure-Requests: 1\r\n\r\nluci_username=root&luci_password=qazwsxedc"
	//bak_auth := "9bda4679861dd7abee831fbbdc96ddf6"

	for {
		n, err := client.Read(buffer[:])
		if err != nil {
			log.Printf("-----CCCCC0-----:%s\n", err)
			if n > 0 {
				if *attack_flag_point {
					_, err := (*shadow_server).Write(buffer[:])
					if err != nil {
						//log.Println(err)
						log.Printf("-----CCCCC1-----:%s\n", err)
						return
					}
					return
				} else {
					flag, real, session := detect_attack(buffer, attack_dict, crack_time)
					if flag {
						(*shadow_server), err = net.Dial("tcp", shadow_address)
						if err != nil {
							//log.Println(err)
							log.Printf("-----CCCCC2-----:%s\n", err)
							return
						}
						(*shadow_server).SetReadDeadline(time.Now().Add(2 * time.Second))
						*attack_flag_point = true
						if !real {
							if len(session) > 0 {
								fmt.Println(magic_hack + session + "\r\n" + magic_hack_end)
								magic_hack_buf := []byte(magic_hack + session + "\r\n" + magic_hack_end)
								(*shadow_server).Write(magic_hack_buf[:])
								var buf [512]byte
								_, err = (*shadow_server).Read(buf[:])
								if err != nil {
									//log.Println(err)
									log.Printf("-----CCCCC3-----:%s\n", err)
									return
								}
								set_session := string(buf[:])
								index := strings.Index(set_session,"Set-Cookie")
								if index > 0 {
									index = index + 20
									end := index + 32
									session = set_session[index: end]
									attack_dict[session] = true
									log.Println(session)
									client.Write(buf[:])
								}
							}
							_, err := (*shadow_server).Write(buffer[:])
							if err != nil {
								//log.Println(err)
								log.Printf("-----CCCCC3-----:%s\n", err)
								return
							}

							go io.Copy(client, *(shadow_server))
						} else {
							_, err := server.Write(buffer[:])
							if err != nil {
								//log.Println(err)
								log.Printf("-----CCCCC4-----:%s\n", err)
								return
							}
						}

					}
				}

			}
			return
		} else {
			if *attack_flag_point {
				_, err := (*shadow_server).Write(buffer[:])
				if err != nil {
					log.Printf("-----CCCCC5-----:%s\n", err)
					return
				}
			} else {
				flag, real, session := detect_attack(buffer, attack_dict, crack_time)
				if flag {
					(*shadow_server), err = net.Dial("tcp", shadow_address)
					if err != nil {
						//log.Println(err)
						log.Printf("-----CCCCC6-----:%s\n", err)
						return
					}
					*attack_flag_point = true
					if !real {
						if len(session) > 0 {
							magic_hack_buf := []byte(magic_hack + session + "\r\n" + magic_hack_end)
							fmt.Println(magic_hack + session + "\r\n" + magic_hack_end)
							(*shadow_server).Write(magic_hack_buf[:])
						}
					}

					_, err := (*shadow_server).Write(buffer[:])

					if err != nil {
						//log.Println(err)
						log.Printf("-----CCCCC7-----:%s\n", err)
						return
					}
					go io.Copy(client, *(shadow_server))
				} else {

					_, err := server.Write(buffer[:])
					if err != nil {
						//log.Println(err)
						log.Printf("-----CCCCC7-----:%s\n", err)
						return
					}
				}

			}
		}
	}
}
