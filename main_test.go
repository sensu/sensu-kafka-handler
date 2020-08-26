package main

import (
	"encoding/json"
	"fmt"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

//func GetUnexportedField(field reflect.Value) interface{} {
//	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
//}

func SetUnexportedField(field reflect.Value, value interface{}) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}

func TestTopicAnnotation(t *testing.T) {
	assert := assert.New(t)
	file, _ := ioutil.TempFile(os.TempDir(), "sensu-kafka-handler-")
	defer func() {
		_ = os.Remove(file.Name())
	}()

	event := corev2.FixtureEvent("entity1", "check1")
	val := `alt-topic`
	m := make(map[string]string)
	m["sensu.io/plugins/sensu-kafa-handler/config/topic"] = val
	event.Check.Annotations = m
	eventJSON, _ := json.Marshal(event)
	_, err := file.WriteString(string(eventJSON))
	require.NoError(t, err)
	require.NoError(t, file.Sync())
	_, err = file.Seek(0, 0)
	require.NoError(t, err)
	os.Stdin = file

	// Using docker compose to startup kafka server and prepare sensu-events topic
	host := "localhost:9092"
	oldArgs := os.Args
	os.Args = []string{"kafka-handler", "-H", host, "-t", "undefined-events"}
	defer func() { os.Args = oldArgs }()

	var exitStatus int
	exitStatus = 0
	mockExit := func(i int) {
		exitStatus = i
	}

	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, checkArgs, executeHandler)
	field := reflect.ValueOf(handler).Elem().FieldByName("exitFunction")
	SetUnexportedField(field, mockExit)
	c1 := make(chan string, 1)
	go func() {
		handler.Execute()
		c1 <- "execute is done"
	}()
	select {
	case <-c1:
		assert.Zero(exitStatus)
	case <-time.After(10 * time.Second):
		fmt.Println("Error: timeout reached in test")
		assert.True(false)
	}

}

func TestExecute(t *testing.T) {
	assert := assert.New(t)
	file, _ := ioutil.TempFile(os.TempDir(), "sensu-kafka-handler-")
	defer func() {
		_ = os.Remove(file.Name())
	}()

	event := corev2.FixtureEvent("entity1", "check1")
	event.Check = nil
	event.Metrics = corev2.FixtureMetrics()
	eventJSON, _ := json.Marshal(event)
	_, err := file.WriteString(string(eventJSON))
	require.NoError(t, err)
	require.NoError(t, file.Sync())
	_, err = file.Seek(0, 0)
	require.NoError(t, err)
	os.Stdin = file

	// Using docker compose to startup kafka server and prepare sensu-events topic
	host := "localhost:9092"
	oldArgs := os.Args
	os.Args = []string{"kafka-handler", "-H", host, "-t", "sensu-events"}
	defer func() { os.Args = oldArgs }()

	var exitStatus int
	exitStatus = 0
	mockExit := func(i int) {
		exitStatus = i
	}

	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, checkArgs, executeHandler)
	field := reflect.ValueOf(handler).Elem().FieldByName("exitFunction")
	SetUnexportedField(field, mockExit)
	c1 := make(chan string, 1)
	go func() {
		handler.Execute()
		c1 <- "execute is done"
	}()
	select {
	case <-c1:
		assert.Zero(exitStatus)
	case <-time.After(10 * time.Second):
		fmt.Println("Error: timeout reached in test")
		assert.True(false)
	}

}
