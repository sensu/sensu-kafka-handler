package main

import (
	"context"
	"encoding/json"
	"fmt"
	kafka "github.com/segmentio/kafka-go"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"strings"
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
	m["sensu.io/plugins/sensu-kafka-handler/config/topic"] = val
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
	entityName := fmt.Sprintf("entity-%v", rand.Int())
	checkName := fmt.Sprintf("check-%v", rand.Int())
	event := corev2.FixtureEvent(entityName, checkName)
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
	// to consume messages
	topic := "sensu-events"
	partition := 0

	conn, _ := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max
	b := make([]byte, 10e3)            // 10KB max per message
	found := false
	for {
		_, err := batch.Read(b)
		if err != nil {
			break
		}
		bstring := strings.TrimRight(string(b), "\x00")
		e := &types.Event{}
		err = json.Unmarshal([]byte(bstring), e)
		if err != nil {
			fmt.Println(bstring)
			fmt.Println(err)
			continue
		}
		if e.Entity != nil {
			if e.Entity.Name == event.Entity.Name {
				if e.Check != nil {
					if e.Check.Name == event.Check.Name {
						fmt.Printf("Found: %v %v\n", e.Entity.Name, e.Check.Name)
						found = true
					}
				}
			}
		}
	}
	batch.Close()
	conn.Close()
	fmt.Println(found)
	assert.True(found)

}
