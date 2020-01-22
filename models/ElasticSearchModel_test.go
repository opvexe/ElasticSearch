package models

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewElasticSearchClient(t *testing.T) {
	var str = "http://localhost:9200"
	str = strings.TrimRight(str, "/") + "/"
	fmt.Println(str)
}
