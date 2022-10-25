package service

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	_ "github.com/go-sql-driver/mysql"
)

func TestMysql(t *testing.T) {
	cmd := exec.Command("docker", "exec", "-i", "1Panel-mysql5.7-RnzE", "mysql", "-uroot", "-pCalong@2016", "-e", "show global variables;")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	kk := strings.Split(string(stdout), "\n")
	testMap := make(map[string]interface{})
	for _, v := range kk {
		itemRow := strings.Split(v, "\t")
		if len(itemRow) == 2 {
			testMap[itemRow[0]] = itemRow[1]
		}
	}
	var info dto.MysqlVariables
	arr, err := json.Marshal(testMap)
	if err != nil {
		fmt.Println(err)
	}
	_ = json.Unmarshal(arr, &info)
	fmt.Print(info)
	// fmt.Println(string(stdout))
	// for {
	// 	str, err := hr.Reader.ReadString('\n')
	// 	if err == nil {
	// 		testMap := make(map[string]interface{})
	// 		err = json.Unmarshal([]byte(str), &testMap)
	// 		fmt.Println(err)
	// 		for k, v := range testMap {
	// 			fmt.Println(k, v)
	// 		}
	// 		// fmt.Print(str)
	// 	} else if err == io.EOF {
	// 		// ReadString最后会同EOF和最后的数据一起返回
	// 		fmt.Println(str)
	// 		break
	// 	} else {
	// 		fmt.Println("出错！！")
	// 		return
	// 	}
	// }
	// input, err := hr.Reader.ReadString('\n')
	// if err == nil {
	// 	fmt.Printf("The input was: %s\n", input)
	// }

	// _, err = hr.Conn.Write([]byte("show global variables; \n"))
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// time.Sleep(3 * time.Second)
	// buf1 := make([]byte, 1024)
	// _, err = hr.Reader.Read(buf1)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(string(buf1))
}
