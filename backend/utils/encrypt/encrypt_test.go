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
	p, err := StringDecrypt("Jmg4EUACGznt3dEQTJ+0ZRxwLaVNsNg7R5RcZ0V7ElQ=")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(p)
}
