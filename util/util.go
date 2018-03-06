package util

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"strconv"
)

//PrettyPrint オブジェクトなどを可視性高くprintします
func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}

func FloatToString(value float64) string {
	return strconv.FormatFloat(value, 'f', 8, 64)
}

func StringToFloat(value string) (float64, error) {
	ret, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0, err
	}
	return ret, nil
}

func SaveJsonToFile(data interface{}, fileName string) {
	bdata := EncodingJSON(data)

	content := []byte(bdata)
	ioutil.WriteFile(fileName, content, os.ModePerm)
}

func EncodingJSON(data interface{}) []byte {
	bdata, err := json.Marshal(data)

	if err != nil {
		return nil
	}
	return bdata
}

//小数を指定の位置で四捨五入します
func Round(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Floor(f*shift+.5) / shift
}
