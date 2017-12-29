package main

import (
	"hash/fnv"
)

/*type HsMsg struct {
	version    byte
	numMethods byte
	methodarr  []byte
}*/

type getRequest struct {
	VER	byte
	CMD	byte
	RSV byte
	ATYP byte
	ADDR []byte
	PORT []byte
}

type getReply struct {
	VER	byte
	REP	byte
	RSV byte
	ATYP byte
	ADDR []byte
	PORT []byte
}

type auth struct{
	VER byte
	NMETHODS byte
	METHODS []byte
}

func (zgr *auth) toByteArr() []byte {
	barr := make([]byte, 1)
	barr[0] = zgr.VER
	barr = append(barr,zgr.NMETHODS)
	barr = iappender(barr, zgr.METHODS)
	//barr = iappender(barr, zgr.PORT)
	//tmpcnv := make([]byte, 4)
	/*binary.LittleEndian.PutUint32(tmpcnv, zgr.datalength)
	barr = iappender(barr, tmpcnv)
	if zgr.datalength != 0 {
		barr = iappender(barr, zgr.data)
	}*/
	return barr
}


type authReply struct{
	VER byte
	METHOD byte
}

func iappender(a []byte ,b []byte ) []byte{
	for i:=0;i<len(b);i++{
		a = append(a,b[i])
	}
	return a
}

func addrxport2id(ipaddr []byte,port []byte) uint32{
	tmpbuff := iappender(ipaddr,port)
	h :=fnv.New32()
	h.Write(tmpbuff)
	return h.Sum32()
}


func (zgr *getReply) toByteArr() []byte {
	barr := make([]byte, 1)
	barr[0] = zgr.VER
	barr = append(barr,zgr.REP)
	barr = append(barr,zgr.RSV)
	barr = append(barr,zgr.ATYP)
	barr = iappender(barr, zgr.ADDR)
	barr = iappender(barr, zgr.PORT)
	//tmpcnv := make([]byte, 4)
	/*binary.LittleEndian.PutUint32(tmpcnv, zgr.datalength)
	barr = iappender(barr, tmpcnv)
	if zgr.datalength != 0 {
		barr = iappender(barr, zgr.data)
	}*/
	return barr
}


func (zgr *getRequest) toByteArr() []byte {
	barr := make([]byte, 1)
	barr[0] = zgr.VER
	barr = append(barr,zgr.CMD)
	barr = append(barr,zgr.RSV)
	barr = append(barr,zgr.ATYP)
	barr = iappender(barr, zgr.ADDR)
	barr = iappender(barr, zgr.PORT)
	//tmpcnv := make([]byte, 4)
	/*binary.LittleEndian.PutUint32(tmpcnv, zgr.datalength)
	barr = iappender(barr, tmpcnv)
	if zgr.datalength != 0 {
		barr = iappender(barr, zgr.data)
	}*/
	return barr
}