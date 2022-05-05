package DomainTool

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type pretty struct{}

var Pretty pretty

func (Pretty pretty) Prepare(data interface{}) string {
	prd, _ := json.MarshalIndent(data, "", "  ")
	return string(prd)
}

func (Pretty pretty) Println(data ...interface{}) {
	var str []interface{}
	if len(data) > 0 {
		for _, d := range data {
			str = append(str, Pretty.Prepare(d))
		}
	}
	fmt.Println(str...)
}

func (Pretty pretty) Save(data interface{}, filename string) {
	file, _ := json.MarshalIndent(data, "", " ")
	_ = ioutil.WriteFile(filename, file, 0644)
}

func (Pretty pretty) Fatalln(data ...interface{}) {
	Pretty.Println(data...)
	os.Exit(1)
}

func GenerateUUIDFromInt(number int) string {
	b := []byte(fmt.Sprintf("%016d", number))
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
