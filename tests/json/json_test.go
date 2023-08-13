package json

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/fatih/structs"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"testing"
)

func TestJson(t *testing.T) {
	const json = `{"name":{"first":"Janet","last":"Prichard"},"age":47}`
	value := gjson.Get(json, "name.last")
	println(value.String())

	data := []byte(`{
  "person": {
    "name": {
      "first": "Leonid",
      "last": "Bugaev",
      "fullName": "Leonid Bugaev"
    },
    "github": {
      "handle": "buger",
      "followers": 109
    },
    "avatars": [
      { "url": "https://avatars1.githubusercontent.com/u/14009?v=3&s=460", "type": "thumbnail" }
    ]
  },
  "company": {
    "name": "Acme"
  }
}`)

	// You can specify key path by providing arguments to Get function
	jsonparser.Get(data, "person", "name", "fullName")

	// There is `GetInt` and `GetBoolean` helpers if you exactly know key data type
	jsonparser.GetInt(data, "person", "github", "followers")

	// When you try to get object, it will return you []byte slice pointer to data containing it
	// In `company` it will be `{"name": "Acme"}`
	jsonparser.Get(data, "company")

	// If the key doesn't exist it will throw an error
	var size int64
	if value, err := jsonparser.GetInt(data, "company", "size"); err == nil {
		size = value
		fmt.Println(size)
	}

	// You can use `ArrayEach` helper to iterate items [item1, item2 .... itemN]
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		fmt.Println(jsonparser.Get(value, "url"))
	}, "person", "avatars")

	// Or use can access fields by index!
	jsonparser.GetString(data, "person", "avatars", "[0]", "url")

	my, err := jsonparser.Set(data, []byte("honggang.liu"), "person", "name", "fullName")
	fmt.Println(string(my))

	val, err := jsonparser.GetString(data, "person", "name", "fullName")
	fmt.Println(val)
	fmt.Println(err)

	//val, dt, offset, err := jsonparser.Get(data, "person", "avatars")
	//
	//if err != nil {
	//	fmt.Println(val)
	//}
	//fmt.Println(val)
	//fmt.Println(dt)
	//fmt.Println(offset)

	// You can use `ObjectEach` helper to iterate objects { "key1":object1, "key2":object2, .... "keyN":objectN }
	jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		fmt.Printf("Key: '%s'\n Value: '%s'\n Type: %s\n", string(key), string(value), dataType)
		return nil
	}, "person", "name")

	// The most efficient way to extract multiple keys is `EachKey`

	paths := [][]string{
		[]string{"person", "name", "fullName"},
		[]string{"person", "avatars", "[0]", "url"},
		[]string{"company", "url"},
	}
	fmt.Println(paths)
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}
type User struct {
	Names    []string `json:"names"`
	Position Position `json:"position"`
	Card     []int    `json:"card"`
	Profile  string   `json:"profile"`
	Tel      int      `json:"tel"`
	IsHigh   bool     `json:"isHigh"`
	HighArr  []bool   `json:"highArr"`
}

func TestStruct(t *testing.T) {
	var nameArr = `["hello","world"]`
	var cardArr = `[1,2]`
	var user = &User{
		Names: []string{"hello", "world"},
		Card:  []int{1, 2},
		Position: Position{
			X: 1,
			Y: 2,
		},
		Profile: "test",
	}
	var dataMap = structs.Map(user)
	fmt.Println(dataMap)

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	e, err := json.Marshal(user)
	if err == nil {
		fmt.Println(e)
	}

	val, dt, offset, err := jsonparser.Get(e, "names")

	if err != nil {
		fmt.Println(val)
	}
	fmt.Println(val)
	fmt.Println(dt)
	fmt.Println(offset)
	fmt.Println(nameArr == string(val))
	var nameByte = []byte(nameArr)
	if len(val) != len(nameByte) {
		fmt.Println("not Eq")
	}
	var eq = true
	for i, vone := range val {
		if vone != nameByte[i] {
			fmt.Println("not Eq")
			eq = false
			break
		}
	}
	fmt.Println(eq)

	val, dt, offset, err = jsonparser.Get(e, "card")

	var carByte = []byte(cardArr)
	fmt.Println(cardArr == string(val))

	if len(val) != len(carByte) {
		fmt.Println("not Eq")
	}
	for i, vone := range val {
		if vone != carByte[i] {
			fmt.Println("not Eq")
			eq = false
			break
		}
	}
	fmt.Println(eq)
}

func TestStruct2(t *testing.T) {
	//var cardStr = "[1,2,3,4]"
	//var constCardStr = ``
	var user = &User{
		Names: []string{"hello", "world"},
		Card:  []int{1, 2},
		Position: Position{
			X: 1,
			Y: 2,
		},
		Profile: "test",
		Tel:     10,
		IsHigh:  true,
		HighArr: []bool{true, false},
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	data, err := json.Marshal(user)
	if err != nil {
		fmt.Println(data)
	}
	var test = "[1,2,3,4]"
	fmt.Println(test)

	var test2 = "hello"
	fmt.Println(test2)
	var isHigh = true
	fmt.Println(isHigh)
	var highArr = "[false, true]"
	fmt.Println(highArr)
	var keys = "highArr"
	_, dt, _, _ := jsonparser.Get(data, keys)
	var valS string
	if dt == jsonparser.Array || dt == jsonparser.Boolean {
		valS = fmt.Sprint(``, highArr, ``)
	} else {
		valS = fmt.Sprint(`"`, test2, `"`)
	}
	my, err := jsonparser.Set(data, []byte(valS), keys)
	fmt.Println(err)
	fmt.Println(my)
	err = json.Unmarshal(my, &user)
	fmt.Println(err)
	fmt.Println(user)
}
