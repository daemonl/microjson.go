package microjson

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var helloObject = "Hello"

func helloWorld(req *http.Request) (interface{}, error) {
	switch req.URL.Path {
	case "/panic":
		panic("Panic As Requested")
	case "/user-error":
		return nil, UserError(400, "RULES", "Test")
	case "/error":
		return nil, fmt.Errorf("Unhandled Error")
	}
	return helloObject, nil
}

func TestGoodWrapper(t *testing.T) {
	h := Wrap(helloWorld)
	req, _ := http.NewRequest("GET", "http://nil/hello", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != 200 {
		t.Errorf("Expected 200, got %d", rw.Code)
	}
	resp := strings.TrimSpace(rw.Body.String())
	if resp != `"Hello"` {
		t.Errorf(`Expected '"Hello"' in JSON, got '%s'`, resp)
	}
}

func TestUserErrorWrapper(t *testing.T) {
	h := Wrap(helloWorld)
	req, _ := http.NewRequest("GET", "http://nil/user-error", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != 400 {
		t.Errorf("Expected 400, got %d", rw.Code)
	}
	resp := map[string]interface{}{}
	json.NewDecoder(rw.Body).Decode(&resp)
	if resp["code"] != "RULES" || resp["error"] != "Test" {
		t.Errorf("Bad error body: %#v", resp)
	}
}

func TestErrorWrapper(t *testing.T) {
	h := Wrap(helloWorld)
	req, _ := http.NewRequest("GET", "http://nil/error", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != 500 {
		t.Errorf("Expected 500, got %d", rw.Code)
	}
	resp := map[string]interface{}{}
	json.NewDecoder(rw.Body).Decode(&resp)
	if resp["code"] != "UNKNOWN" || resp["error"] != "Internal Server Error" {
		t.Errorf("Bad error body: %#v", resp)
	}
}

func TestPanicWrapper(t *testing.T) {
	h := Wrap(helloWorld)
	req, _ := http.NewRequest("GET", "http://nil/panic", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != 500 {
		t.Errorf("Expected 500, got %d", rw.Code)
	}
	resp := map[string]interface{}{}
	json.NewDecoder(rw.Body).Decode(&resp)
	if resp["code"] != "UNKNOWN" || resp["error"] != "Internal Server Error" {
		t.Errorf("Bad error body: %#v", resp)
	}
}

func helloDatabase(req *http.Request, tx *sql.Tx) (interface{}, error) {
	if _, err := tx.Exec("INSERT INTO test (val) VALUES (1)"); err != nil {
		panic("Actual unexpected error in test " + err.Error())
	}
	return helloWorld(req)
}

func TestDBWrap(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer db.Close()

	wrapper := TxWrapper{
		DB: db,
	}

	handler := wrapper.TxHandler(StandardSQLHandler(helloDatabase))

	mock.ExpectBegin()
	mock.ExpectExec(`.*`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectRollback()

	func() {
		req, _ := http.NewRequest("GET", "http://nil/panic?t=panic-test", nil)
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)
		if rw.Code != 500 {
			t.Errorf("Expected 500, got %d", rw.Code)
		}
		resp := map[string]interface{}{}
		json.NewDecoder(rw.Body).Decode(&resp)
		if resp["code"] != "UNKNOWN" || resp["error"] != "Internal Server Error" {
			t.Errorf("Bad error body: %#v", resp)
		}
	}()

	mock.ExpectBegin()
	mock.ExpectExec(`.*`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectRollback()

	func() {
		req, _ := http.NewRequest("GET", "http://nil/error?t=error-test", nil)
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)
		if rw.Code != 500 {
			t.Errorf("Expected 500, got %d", rw.Code)
		}
		resp := map[string]interface{}{}
		json.NewDecoder(rw.Body).Decode(&resp)
		if resp["code"] != "UNKNOWN" || resp["error"] != "Internal Server Error" {
			t.Errorf("Bad error body: %#v", resp)
		}

	}()

	mock.ExpectBegin()
	mock.ExpectExec(`.*`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	func() {
		req, _ := http.NewRequest("GET", "http://nil/hello?t=good-test", nil)
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)
		if rw.Code != 200 {
			t.Errorf("Expected 200, got %d", rw.Code)
		}

	}()
}

func helloQueueDatabase(req *http.Request, tx *Tx) (interface{}, error) {
	tx.DeferPublish("test", "Message")
	return helloDatabase(req, tx.Tx)
}

type mockPublisher struct{}

func (mp *mockPublisher) Publish(message interface{}) error {
	return nil
}

func TestDBQueueWrap(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer db.Close()

	wrapper := TxWrapper{
		DB: db,
		Publishers: map[string]MessagePublisher{
			"test": &mockPublisher{},
		},
	}

	handler := wrapper.TxHandler(helloQueueDatabase)

	mock.ExpectBegin()
	mock.ExpectExec(`.*`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectRollback()

	func() {
		req, _ := http.NewRequest("GET", "http://nil/panic?t=panic-test", nil)
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)
		if rw.Code != 500 {
			t.Errorf("Expected 500, got %d", rw.Code)
		}
		resp := map[string]interface{}{}
		json.NewDecoder(rw.Body).Decode(&resp)
		if resp["code"] != "UNKNOWN" || resp["error"] != "Internal Server Error" {
			t.Errorf("Bad error body: %#v", resp)
		}
	}()

	mock.ExpectBegin()
	mock.ExpectExec(`.*`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectRollback()

	func() {
		req, _ := http.NewRequest("GET", "http://nil/error?t=error-test", nil)
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)
		if rw.Code != 500 {
			t.Errorf("Expected 500, got %d", rw.Code)
		}
		resp := map[string]interface{}{}
		json.NewDecoder(rw.Body).Decode(&resp)
		if resp["code"] != "UNKNOWN" || resp["error"] != "Internal Server Error" {
			t.Errorf("Bad error body: %#v", resp)
		}

	}()

	mock.ExpectBegin()
	mock.ExpectExec(`.*`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	func() {
		req, _ := http.NewRequest("GET", "http://nil/hello?t=good-test", nil)
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)
		if rw.Code != 200 {
			t.Errorf("Expected 200, got %d", rw.Code)
		}

	}()

}
