package encrypt

import (
	"fmt"
	"testing"

	"github.com/1Panel-dev/1Panel/init/viper"
)

func TestStringEncrypt(t *testing.T) {
	viper.Init()
	p, err := StringEncrypt("Songliu123++")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(p)
}

func TestStringDecrypt(t *testing.T) {
	viper.Init()
	p, err := StringDecrypt("5WYEZ4XcitdomVvAyimt9WwJwBJJSbTTHncZoqyOraQ=")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(p)
}
