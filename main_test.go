package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/peddlrph/apiserver"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var a main.App

//var c main.Conf

func TestMain(m *testing.M) {
	a = main.App{}
	a.Initialize()

	ensureMessagesTableExists()

	code := m.Run()

	clearMessagesTable()

	os.Exit(code)
}

func getBearer() string {
	conf := main.Config{}

	file, err := os.Open("./config.json")
	if err != nil {
		fmt.Println(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		fmt.Println(err)
	}
	return conf.Token
}

func ensureMessagesTableExists() {
	if _, err := a.DB.Exec(createMessagesTable); err != nil {
		log.Fatal(err)
	}
}

func clearMessagesTable() {
	a.DB.Exec("TRUNCATE TABLE messages")
	//a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const createMessagesTable = `CREATE TABLE IF NOT EXISTS messages
(
id int(10) unsigned NOT NULL,
body varchar(200),
msg_box varchar(20) NOT NULL,
address varchar(20) NOT NULL,
datetime varchar(30),
created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

func TestEmptyMessagesTable(t *testing.T) {
	clearMessagesTable()

	req, _ := http.NewRequest("GET", "/messages", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetMessage(t *testing.T) {
	clearMessagesTable()
	addMessages(1)

	req, _ := http.NewRequest("GET", "/message/8888", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestGetLastMessage(t *testing.T) {
	clearMessagesTable()
	addMessages(1)

	req, _ := http.NewRequest("GET", "/message/last", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestCreateMessage(t *testing.T) {
	//	clearTable()

	payload := []byte(`{"id":1111,"body":"Hello","msg_box":"outbox","address":"0123456789","datetime":"11/15/2017"}`)

	req, _ := http.NewRequest("POST", "/message", bytes.NewBuffer(payload))
	//req.Header.Set("Authorization", Bearer)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != 1111.0 {
		t.Errorf("Expected message id to be '1111.0'. Got '%v'", m["id"])
	}

	if m["body"] != "Hello" {
		t.Errorf("Expected message body to be 'Hello'. Got '%v'", m["body"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["address"] != "0123456789" {
		t.Errorf("Expected address to be '0123456789'. Got '%v'", m["address"])
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()

	Bearer := getBearer()
	req.Header.Set("Authorization", Bearer)
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func addMessages(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO messages(id, msg_box,address,body,datetime) VALUES(?,?,?,?,?)", 8888, "inbox", "09898902885", "Hello", "10/15/2018")
	}
}
