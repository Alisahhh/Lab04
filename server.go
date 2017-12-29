package main

import (
	"fmt"
	"net"
	"os"
	"encoding/binary"
	"time"
)

type IOpair struct{
	in chan []byte
	out chan []byte
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func main() {
	service := ":1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	checkError(err)
	l, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	conns :=make(map[uint32]IOpair)
	conn,err:=l.Accept()
	//auth
	var hsm auth
	buf:=make([]byte,1)
	_,err=conn.Read(buf)
	if err != nil {
		fmt.Printf("Network Error.Terminating..\n")
		return
	}
	if buf[0] != 0x05 {
		fmt.Printf("Protocol implementation error.Terminating...\n")
		return
	}
	_,err=conn.Read(buf)
	if err != nil {
		fmt.Printf("Network Error.Terminating..\n")
		return
	}
	hsm.VER=0x05
	hsm.NMETHODS=buf[0]
	//buf=make([]byte,buf[0])
	hsm.METHODS=make([]byte,buf[0])
	_,err=conn.Read(hsm.METHODS)
	if err != nil {
		fmt.Printf("Network Error.Terminating..\n")
		return
	}
	hsm.NMETHODS=0x01
	//buf=make([]byte,buf[0])
	hsm.METHODS[0]=0x00

	var authRep authReply
	authRep.VER = 0x05
	authRep.METHOD=0x00
	tmpbuf:=make([]byte,2)
	tmpbuf[0]=0x05
	tmpbuf[1]=0x00
	_,err=conn.Write(tmpbuf)
	if err!=nil{
		fmt.Printf("Network Error.Terminating..\n")
		return
	}
	//auth success
	for{
		//cnt++
		//fmt.Println(cnt)
		if err!=nil{
			continue
		}
		buf:=make([]byte,4)
		var zgr getRequest
		_,err=conn.Read(buf)
		if err != nil {
			fmt.Printf("Network Error %v.Terminating..\n", err.Error())
			os.Exit(-1)
		}
		zgr.VER=buf[0]
		zgr.CMD=buf[1]
		zgr.RSV=buf[2]
		zgr.ATYP=buf[3]
		if zgr.ATYP==0x01{
			zgr.ADDR=make([]byte,4)
			_,err:=conn.Read(zgr.ADDR)
			if err != nil {
				fmt.Printf("Network Error %v.Terminating..\n", err.Error())
				os.Exit(-1)
			}
		}else if zgr.ATYP==0x04{
			zgr.ADDR=make([]byte,16)
			_,err:=conn.Read(zgr.ADDR)
			if err != nil {
				fmt.Printf("Network Error %v.Terminating..\n", err.Error())
				os.Exit(-1)
			}
		}else if zgr.ATYP==0x03{//change
			tmpbuf=make([]byte,1)
			_,err=conn.Read(tmpbuf)
			if err != nil {
				fmt.Printf("Network Error %v.Terminating..\n", err.Error())
				os.Exit(-1)
			}
			buf=make([]byte,int(tmpbuf[0]))
			_,err=conn.Read(buf)
			if err != nil {
				fmt.Printf("Network Error %v.Terminating..\n", err.Error())
				os.Exit(-1)
			}
			zgr.ADDR=make([]byte,int(tmpbuf[0])+1)
			zgr.ADDR[0]=tmpbuf[0]
			for i:=0;i<int(tmpbuf[0]);i++{
				zgr.ADDR=append(zgr.ADDR,buf[i])
			}

		}

		zgr.PORT=make([]byte,2)
		_,err=conn.Read(zgr.PORT)
		if err != nil {
			fmt.Printf("Network Error %v.Terminating..\n", err.Error())
			os.Exit(-1)
		}

		/*tmpLen:=make([]byte,4)
		_,err=conn.Read(tmpLen)
		if err != nil {
			fmt.Printf("Network Error %v.Terminating..\n", err.Error())
			os.Exit(-1)
		}
		zgr.datalength=binary.LittleEndian.Uint32(tmpLen)

		if zgr.datalength!=0{
			zgr.data=make([]byte,zgr.datalength)
			_,err:=conn.Read(zgr.data)
			if err != nil {
				fmt.Printf("Network Error %v.Terminating..\n", err.Error())
				os.Exit(-1)
			}
		}*/
		go func (zgr getRequest){
			//zgr.data= AESDecrypt(zgr.data)
			//pair, isok := conns[addrxport2id(zgr.ADDR, zgr.PORT)]
			//if !isok{
				var tmpiop,pair IOpair
				tmpiop.in = make(chan []byte, 5)
				tmpiop.out = make(chan []byte, 5)
				go func(){
					var finalTcpAddr net.TCPAddr
					finalTcpAddr.IP=zgr.ADDR
					finalTcpAddr.Port=int(binary.BigEndian.Uint16(zgr.PORT))

					sconn, err := net.DialTCP("tcp", nil, &finalTcpAddr)
					if err!=nil{
						fmt.Printf("NetworkError.Cancelling...\n")
						fmt.Println(err.Error())
						delete(conns, addrxport2id(zgr.ADDR, zgr.PORT))
						return
					}
					defer func(){
						delete(conns, addrxport2id(zgr.ADDR, zgr.PORT))
						sconn.Close()
					}()
					/*go func(){
						for{
							tmpbuf=make([]byte,1024)
							n,err:=sconn.Read(tmpbuf)
							if err != nil {
								fmt.Printf("Prox to Dest Conn error %v. Terminating..\n", err.Error())
								close(tmpiop.out)
								delete(conns, addrxport2id(zgr.ADDR, zgr.PORT))
								sconn.Close()
								return
							}
							tmpbuf=tmpbuf[:n]
							tmpiop.out<-tmpbuf
						}
					}()*/
					for{
						select{
						case data :=<-tmpiop.in:
							sconn.Write(hsm.toByteArr())
							tmpbuf:=make([]byte,2)
							_,err=sconn.Read(tmpbuf)
							if err!=nil||tmpbuf[0]!=0x05||tmpbuf[1]!=0x00{
								fmt.Printf("Network Error %v.Terminating..\n", err.Error())
								os.Exit(-1)
							}
							_,err:=sconn.Write(data)
							if err!=nil{
								fmt.Printf("Network Error.Retrying(In case http closure)..\n")
								sconn, err := net.DialTCP("tcp", nil, &finalTcpAddr)
								if err!=nil{
									fmt.Printf("Network fail retry failed.Terminating..\n")
									return
								}
								defer sconn.Close()
								_,err=sconn.Write(data)
								if err!=nil{
									fmt.Printf("Network fail retry failed.Terminating..\n")
									return
								}
							}
							tmpbuf=make([]byte,1024)
							n,err:=sconn.Read(tmpbuf)
							if err != nil {
								fmt.Printf("Prox to Dest Conn error %v. Terminating..\n", err.Error())
								close(tmpiop.out)
								delete(conns, addrxport2id(zgr.ADDR, zgr.PORT))
								sconn.Close()
								return
							}
							tmpbuf=tmpbuf[:n]
							tmpiop.out<-tmpbuf
						case <-time.After(20 * time.Second):
							fmt.Printf("Timed out.Terminating..\n")
							return
						}
					}
				}()
				//conns[addrxport2id(zgr.ADDR, zgr.PORT)] = tmpiop
				//pair=tmpiop
				go func(){
					for{
						tmpbuff, ok := <-pair.out
						if(!ok){
							return
							var zgrr getReply
							zgrr.VER = zgr.VER
							zgrr.REP = zgr.CMD
							zgrr.PORT = zgr.PORT
							zgrr.ATYP = zgr.ATYP

							//zgrr.data = make([]byte, 1)
							_, err := conn.Write(zgrr.toByteArr())
							if err != nil {
								fmt.Printf("Interserver side Conn error %v.Terminating..\n", err.Error())
								conn.Close()
								os.Exit(-1)
							}
							return
						}

						fmt.Printf("Received str:\n%v\n", string(tmpbuff))
						//conn.Write(tmpbuff)
						/*var zgrr getReply
						zgrr.VER = zgr.VER
						zgrr.REP = zgr.CMD
						zgrr.PORT = zgr.PORT
						zgrr.ATYP = zgr.ATYP
						zgrr.datalength = uint32(len(tmpbuff))
						zgrr.data=tmpbuff
						//zgrr.data = AESEncrypt(tmpbuff)*/

						go func() {
							_, err := conn.Write(tmpbuff)
							if err != nil {
								fmt.Printf("Network error.Terminating..\n")
								conn.Close()
								os.Exit(-1)
							}
						}()

					}
				}()
			//}
			pair.in <- zgr.toByteArr()
		}(zgr)
	}


}

