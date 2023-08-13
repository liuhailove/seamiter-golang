package file

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

const (
	TestSystemRules = `[
    {
        "id": "0",
        "metricType": 0,
        "adaptiveStrategy": 0
    },
    {
        "id": "1",
        "metricType": 0,
        "adaptiveStrategy": 0
    },
    {
        "id": "2",
        "metricType": 0,
        "adaptiveStrategy": 0
    }
]`
)

var (
	TestSystemRulesDir  = "./"
	TestSystemRulesFile = TestSystemRulesDir + "SystemRules.json"
)

func prepareSystemRulesTestFile() error {
	content := []byte(TestSystemRules)
	return ioutil.WriteFile(TestSystemRulesFile, content, os.ModePerm)
}

func deleteSystemRulesTestFile() error {
	return os.Remove(TestSystemRulesFile)
}

func TestRefreshableFileDataSource_ReadSource(t *testing.T) {
	t.Run("RefreshableFIleDataSource_ReadSource_Nil", func(t *testing.T) {
		err := prepareSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to prepare test file, err: %+v", err)
		}
		s := &RefreshableFileDataSource{sourceFilePath: TestSystemRulesFile + "NotExisted"}
		got, err := s.ReadSource()
		assert.True(t, got == nil && err != nil && strings.Contains(err.Error(), "RefreshableFileDataSource fail to open the property file"))

		err = deleteSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to delete test file, err: %+v", err)
		}
	})

	t.Run("RefreshableFileDataSource_ReadSource_Normal", func(t *testing.T) {
		err := prepareSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to prepare test file, err: %+v", err)
		}

		s := &RefreshableFileDataSource{
			sourceFilePath: TestSystemRulesFile,
		}
		got, err := s.ReadSource()
		if err != nil {
			t.Errorf("Fail to execute ReadSource, err: %+v", err)
		}
		assert.True(t, reflect.DeepEqual(got, []byte(TestSystemRules)))

		err = deleteSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to delete test file, err: %+v", err)
		}
	})
}
