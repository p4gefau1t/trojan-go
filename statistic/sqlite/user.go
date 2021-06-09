package sqlite

import "encoding/binary"

type User struct {
	Hash string `gorm:"primary_key"`
	// uint64 = 8 byte binary
	Sent      []byte
	Recv      []byte
	MaxIPNum  int
	SendLimit int
	RecvLimit int
}

func (u *User) setSent(sent uint64) {
	binary.BigEndian.PutUint64(u.Sent, sent)
}
func (u *User) getSent() uint64 {
	return binary.BigEndian.Uint64(u.Sent)
}

func (u *User) setRecv(recv uint64) {
	binary.BigEndian.PutUint64(u.Recv, recv)
}
func (u *User) getRecv() uint64 {
	return binary.BigEndian.Uint64(u.Recv)
}

func (u *User) GetHash() string {
	return u.Hash
}

func (u *User) GetTraffic() (sent, recv uint64) {
	return u.getSent(), u.getRecv()
}

func (u *User) GetSpeedLimit() (sent, recv int) {
	return u.SendLimit, u.RecvLimit
}

func (u *User) GetIPLimit() int {
	return u.MaxIPNum
}
