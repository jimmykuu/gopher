package gopher

import (
	"testing"
)

func TestParseJsonFile(t *testing.T) {
	var c ConfigStruct
	parseJsonFile("etc/config.json", &c)
	if c.DB == "" {
		t.Fatal("parse json file failed.")
	}
	/*
		analyticsCode = getDefaultCode(Config.AnalyticsFile)
		shareCode = getDefaultCode(Config.ShareCodeFile)
			if analyticsCode == "" || shareCode == "" {
				t.Fatal("get default code failed.")
			}
	*/
}
