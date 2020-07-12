package url

import (
	crand "crypto/rand"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"testing"
)

func TestHandleTrojanPort_Default(t *testing.T) {
	port, e := handleTrojanPort("")
	assert.Nil(t, e, "empty port should not error")
	assert.EqualValues(t, 443, port, "empty port should fallback to 443")
}

func TestHandleTrojanPort_NotNumber(t *testing.T) {
	_, e := handleTrojanPort("fuck")
	assert.Error(t, e, "non-numerical port should error")
}

func TestHandleTrojanPort_GoodNumber(t *testing.T) {
	testCases := []string{"443", "8080", "10086", "80", "65535", "1"}
	for _, testCase := range testCases {
		_, e := handleTrojanPort(testCase)
		assert.Nil(t, e, "good port %s should not error", testCase)
	}
}

func TestHandleTrojanPort_InvalidNumber(t *testing.T) {
	testCases := []string{"443.0", "443.000", "8e2", "3.5", "9.99", "-1", "-65535", "65536", "0"}

	for _, testCase := range testCases {
		_, e := handleTrojanPort(testCase)
		assert.Error(t, e, "invalid number port %s should error", testCase)
	}
}

func TestNewShareInfoFromURL_Empty(t *testing.T) {
	_, e := NewShareInfoFromURL("")
	assert.Error(t, e, "empty link should lead to error")
}

func TestNewShareInfoFromURL_RandomCrap(t *testing.T) {
	for i := 0; i < 100; i++ {
		randomCrap, _ := ioutil.ReadAll(io.LimitReader(crand.Reader, 10))
		_, e := NewShareInfoFromURL(string(randomCrap))
		assert.Error(t, e, "random crap %v should lead to error", randomCrap)
	}
}

func TestNewShareInfoFromURL_NotTrojanGo(t *testing.T) {
	testCases := []string{
		"trojan://what.ever@www.twitter.com:443?allowInsecure=1&allowInsecureHostname=1&allowInsecureCertificate=1&sessionTicket=0&tfo=1#some-trojan",
		"ssr://d3d3LnR3aXR0ZXIuY29tOjgwOmF1dGhfc2hhMV92NDpjaGFjaGEyMDpwbGFpbjpZbkpsWVd0M1lXeHMvP29iZnNwYXJhbT0mcmVtYXJrcz02TC1INXB5ZjVwZTI2WmUwNzd5YU1qQXlNQzB3TnkweE9DQXhNam8xTlRveU1RJmdyb3VwPVEzUkRiRzkxWkNCVFUxSQ",
		"vmess://eyJhZGQiOiJtb3RoZXIuZnVja2VyIiwiYWlkIjowLCJpZCI6IjFmYzI0NzVmLThmNDMtM2FlYi05MzUyLTU2MTFhZjg1NmQyOSIsIm5ldCI6InRjcCIsInBvcnQiOjEwMDg2LCJwcyI6Iui/h+acn+aXtumXtO+8mjIwMjAtMDYtMjMiLCJ0bHMiOiJub25lIiwidHlwZSI6Im5vbmUiLCJ2IjoyfQ==",
	}

	for _, testCase := range testCases {
		_, e := NewShareInfoFromURL(testCase)
		assert.Error(t, e, "non trojan-go link %s should not decode", testCase)
	}
}

func TestNewShareInfoFromURL_EmptyTrojanHost(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://fuckyou@:443/")
	assert.Error(t, e, "empty host should not decode")
}

func TestNewShareInfoFromURL_BadPassword(t *testing.T) {
	testCases := []string{
		"trojan-go://we:are:the:champion@114514.go",
		"trojan-go://evilpassword:@1919810.me",
		"trojan-go://evilpassword::@1919810.me",
		"trojan-go://@password.404",
		"trojan-go://mother.fuck#yeah",
	}

	for _, testCase := range testCases {
		_, e := NewShareInfoFromURL(testCase)
		assert.Error(t, e, "bad password link %s should not decode", testCase)
	}
}

func TestNewShareInfoFromURL_GoodPassword(t *testing.T) {
	testCases := []string{
		"trojan-go://we%3Aare%3Athe%3Achampion@114514.go",
		"trojan-go://evilpassword%3A@1919810.me",
		"trojan-go://passw0rd-is-a-must@password.200",
	}

	for _, testCase := range testCases {
		_, e := NewShareInfoFromURL(testCase)
		assert.Nil(t, e, "good password link %s should decode", testCase)
	}
}

func TestNewShareInfoFromURL_BadPort(t *testing.T) {
	testCases := []string{
		"trojan-go://pswd@example.com:114514",
		"trojan-go://pswd@example.com:443.0",
		"trojan-go://pswd@example.com:-1",
		"trojan-go://pswd@example.com:8e2",
		"trojan-go://pswd@example.com:65536",
	}

	for _, testCase := range testCases {
		_, e := NewShareInfoFromURL(testCase)
		assert.Error(t, e, "decode url %s with invalid port should error", testCase)
	}
}

func TestNewShareInfoFromURL_BadQuery(t *testing.T) {
	testCases := []string{
		"trojan-go://cao@ni.ma?NMSL=%CG%GE%CAONIMA",
		"trojan-go://ni@ta.ma:13/?#%2e%fu",
	}

	for _, testCase := range testCases {
		_, e := NewShareInfoFromURL(testCase)
		assert.Error(t, e, "parse bad query should error")
	}

}

func TestNewShareInfoFromURL_SNI_Empty(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?sni=")
	assert.Error(t, e, "empty SNI should not be allowed")
}

func TestNewShareInfoFromURL_SNI_Default(t *testing.T) {
	info, e := NewShareInfoFromURL("trojan-go://a@b.c")
	assert.Nil(t, e)
	assert.Equal(t, info.TrojanHost, info.SNI, "default sni should be trojan hostname")
}

func TestNewShareInfoFromURL_SNI_Multiple(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?sni=a&sni=b&sni=c")
	assert.Error(t, e, "multiple SNIs should not be allowed")
}

func TestNewShareInfoFromURL_Type_Empty(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?type=")
	assert.Error(t, e, "empty type should not be allowed")
}

func TestNewShareInfoFromURL_Type_Default(t *testing.T) {
	info, e := NewShareInfoFromURL("trojan-go://a@b.c")
	assert.Nil(t, e)
	assert.Equal(t, ShareInfoTypePlain, info.Type, "default type should be plain")
}

func TestNewShareInfoFromURL_Type_Invalid(t *testing.T) {
	invalidTypes := []string{"nmsl", "dio"}
	for _, invalidType := range invalidTypes {
		_, e := NewShareInfoFromURL(fmt.Sprintf("trojan-go://a@b.c?type=%s", invalidType))
		assert.Error(t, e, "%s should not be a valid type", invalidType)
	}
}

func TestNewShareInfoFromURL_Type_Multiple(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?type=a&type=b&type=c")
	assert.Error(t, e, "multiple types should not be allowed")
}

func TestNewShareInfoFromURL_Host_Empty(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?host=")
	assert.Error(t, e, "empty host should not be allowed")
}

func TestNewShareInfoFromURL_Host_Default(t *testing.T) {
	info, e := NewShareInfoFromURL("trojan-go://a@b.c")
	assert.Nil(t, e)
	assert.Equal(t, info.TrojanHost, info.Host, "default host should be trojan hostname")
}

func TestNewShareInfoFromURL_Host_Multiple(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?host=a&host=b&host=c")
	assert.Error(t, e, "multiple hosts should not be allowed")
}

func TestNewShareInfoFromURL_Type_WS_Multiple(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?type=ws&path=a&path=b&path=c")
	assert.Error(t, e, "multiple paths should not be allowed in wss")
}

func TestNewShareInfoFromURL_Path_WS_None(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?type=ws")
	assert.Error(t, e, "ws should require path")
}

func TestNewShareInfoFromURL_Path_WS_Empty(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?type=ws&path=")
	assert.Error(t, e, "empty path should not be allowed in ws")
}

func TestNewShareInfoFromURL_Path_WS_Invalid(t *testing.T) {
	invalidPaths := []string{"../", ".+!", " "}
	for _, invalidPath := range invalidPaths {
		_, e := NewShareInfoFromURL(fmt.Sprintf("trojan-go://a@b.c?type=ws&path=%s", invalidPath))
		assert.Error(t, e, "%s should not be a valid path in ws", invalidPath)
	}
}

func TestNewShareInfoFromURL_Path_Plain_Empty(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?type=plain&path=")
	assert.Nil(t, e, "empty path should be ignored in plain")
}

func TestNewShareInfoFromURL_Encryption_Empty(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?encryption=")
	assert.Error(t, e, "encryption should not be empty")
}

func TestNewShareInfoFromURL_Encryption_Unknown(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?encryption=motherfucker")
	assert.Error(t, e, "unknown encryption should not be supported")
}

func TestNewShareInfoFromURL_Encryption_SS_NotSupportedMethods(t *testing.T) {
	invalidMethods := []string{"rc4-md5", "rc4", "des-cfb", "table", "salsa20-ctr"}
	for _, invalidMethod := range invalidMethods {
		_, e := NewShareInfoFromURL(fmt.Sprintf("trojan-go://a@b.c?encryption=ss%%3B%s%%3Bshabi", invalidMethod))
		assert.Error(t, e, "encryption %s should not be supported by ss", invalidMethod)
	}
}

func TestNewShareInfoFromURL_Encryption_SS_NoPassword(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?encryption=ss%3Baes-256-gcm%3B")
	assert.Error(t, e, "empty ss password should not be allowed")
}

func TestNewShareInfoFromURL_Encryption_SS_BadParams(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?encryption=ss%3Ba")
	assert.Error(t, e, "broken ss param should not be allowed")
}

func TestNewShareInfoFromURL_Encryption_Multiple(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?encryption=a&encryption=b&encryption=c")
	assert.Error(t, e, "multiple encryption should not be allowed")
}

func TestNewShareInfoFromURL_Plugin_Empty(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?plugin=")
	assert.Error(t, e, "plugin should not be empty")
}

func TestNewShareInfoFromURL_Plugin_Multiple(t *testing.T) {
	_, e := NewShareInfoFromURL("trojan-go://a@b.c?plugin=a&plugin=b&plugin=c")
	assert.Error(t, e, "multiple plugin should not be allowed")
}
