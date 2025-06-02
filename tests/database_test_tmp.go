package idm_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvFileExists(t *testing.T) {
	dir, _ := os.Getwd()
	fmt.Println("Current dir:", dir)

	_, err := os.Stat("../.env")
	assert.NoError(t, err, ".env файл должен быть в корне проекта")
}
