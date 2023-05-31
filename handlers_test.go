package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"golang.org/x/net/html"
)

func TestGetIndexHandler(t *testing.T) {
	// Test stuff that doesn't need chi
	getIndexHandle := getIndex(nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	getIndexHandle.ServeHTTP(rec, req)

	resp := rec.Result()
	body, _ := io.ReadAll(resp.Body)

	if rec.Code != http.StatusOK {
		t.Errorf("getIndex() expected %v, got %v", http.StatusOK, rec.Code)
	}
	if !strings.Contains(string(body), "About") {
		t.Errorf("getIndex() appears to have returned the incorrect page")
	}

}

func TestGetIndexHandlerChi(t *testing.T) {

	r := chi.NewRouter()
	r.Get("/", getIndex(nil))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	resp := rec.Result()
	body, _ := io.ReadAll(resp.Body)

	if rec.Code != http.StatusOK {
		t.Errorf("getIndex() expected %v, got %v", http.StatusOK, rec.Code)
	}
	if !strings.Contains(string(body), "About") {
		t.Errorf("getIndex() appears to have returned the incorrect page")
	}
}

func TestGetCodeHandler(t *testing.T) {
	// Test stuff that doesn't need chi
	testDB := setupTestDB(t)
	getCodeHandle := getCode(testDB)
	req := httptest.NewRequest(http.MethodGet, "/testpath", nil)
	rec := httptest.NewRecorder()
	getCodeHandle.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("getCode() expected %v, got %v", http.StatusOK, rec.Code)
	}
}

func TestGetCodeHandlerChi(t *testing.T) {
	testDB := setupTestDB(t)

	r := chi.NewRouter()
	r.Get("/{id}", getCode(testDB))

	req := httptest.NewRequest(http.MethodGet, "/testpath", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	resp := rec.Result()
	body, _ := io.ReadAll(resp.Body)

	if rec.Code != http.StatusOK {
		t.Errorf("getCode() expected %v, got %v", http.StatusOK, rec.Code)
	}
	if !strings.Contains(string(body), "This page is claimed") {
		t.Errorf("getCode() appears to have returned the incorrect page")
	}

	req = httptest.NewRequest(http.MethodGet, "/doesnotexistyet", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	resp = rec.Result()
	body, _ = io.ReadAll(resp.Body)

	if rec.Code != http.StatusOK {
		t.Errorf("getCode() expected %v, got %v", http.StatusOK, rec.Code)
	}
	if strings.Contains(string(body), "This page is claimed") {
		t.Errorf("getCode() appears to have returned the incorrect page")
	}

}

func TestPostEncodeHandlerChi(t *testing.T) {
	testDB := setupTestDB(t)

	r := chi.NewRouter()
	r.Post("/{id}/encode", postEncode(testDB))

	form := url.Values{}
	form.Add("encInput", "abc")

	req := httptest.NewRequest(http.MethodPost, "/testpath/encode", nil)
	//req := httptest.NewRequest(http.MethodPost, "/testpath/encode", strings.NewReader(form.Encode()))
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Form = form
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	resp := rec.Result()

	if rec.Code != http.StatusOK {
		t.Errorf("postEncode() expected %v, got %v", http.StatusOK, rec.Code)
	}

	htmlResp, err := html.Parse(resp.Body)
	if err != nil {
		//TODO
		t.Errorf("html parse error: %v", err)
	}
	tag := getElementById(htmlResp, "encOutput")
	nodeOutput := renderNode(tag)

	if !strings.Contains(nodeOutput, "Encoded text: zyx") {
		t.Errorf("postEncode() appears to have returned the incorrect page")
	}

}

func TestPostDecodeHandlerChi(t *testing.T) {
	testDB := setupTestDB(t)

	r := chi.NewRouter()
	r.Post("/{id}/decode", postDecode(testDB))

	form := url.Values{}
	form.Add("decInput", "zyx")

	req := httptest.NewRequest(http.MethodPost, "/testpath/decode", nil)
	//req := httptest.NewRequest(http.MethodPost, "/testpath/encode", strings.NewReader(form.Encode()))
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Form = form
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	resp := rec.Result()
	//body, _ := io.ReadAll(resp.Body)
	//fmt.Println(string(body))

	//tkn := html.NewTokenizer(resp.Body)
	htmlResp, err := html.Parse(resp.Body)
	if err != nil {
		//TODO
		t.Errorf("html parse error: %v", err)
	}
	tag := getElementById(htmlResp, "decOutput")
	nodeOutput := renderNode(tag)

	if rec.Code != http.StatusOK {
		t.Errorf("postEncode() expected %v, got %v", http.StatusOK, rec.Code)
	}
	if !strings.Contains(nodeOutput, "Decoded text: abc") {
		t.Errorf("postEncode() appears to have returned the incorrect page")
	}

}

func TestPostSaveMapHandlerChi(t *testing.T) {
	testDB := setupTestDB(t)

	r := chi.NewRouter()
	r.Post("/{id}/save", postSaveMap(testDB))

	form := url.Values{}
	for r := 'a'; r <= 'z'; r++ {
		keyString := fmt.Sprintf("%c", r)
		valueString := fmt.Sprintf("%c", 219-r)
		form.Add(keyString, valueString)
	}

	form.Add("pathPass", "password123")

	req := httptest.NewRequest(http.MethodPost, "/testpath/save", nil)
	req.Form = form
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	resp := rec.Result()

	if rec.Code != http.StatusOK {
		t.Errorf("postEncode() expected %v, got %v", http.StatusOK, rec.Code)
	}

	htmlResp, err := html.Parse(resp.Body)
	if err != nil {
		//TODO
		t.Errorf("html parse error: %v", err)
	}

	errTag := getElementById(htmlResp, "errMsg")
	if errTag != nil {
		nodeOutput := renderNode(errTag)
		t.Errorf("postEncode() should not have returned an error message, but it did: %v", nodeOutput)
	}

	aTag := getElementById(htmlResp, "a")
	if aTag == nil {
		t.Errorf("postEncode() did not return the map correctly")
	} else {
		nodeOutput := renderNode(aTag)
		if !strings.Contains(nodeOutput, "value=\"z\"") {
			t.Errorf("postEncode() should have returned a z, instead returned: %v", nodeOutput)
		}
	}

	// try changing map values
	form2 := url.Values{}
	form2.Add("a", "a")
	form2.Add("z", "z")
	form2.Add("pathPass", "password123")

	req = httptest.NewRequest(http.MethodPost, "/testpath/save", nil)
	req.Form = form2
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	resp2 := rec.Result()
	if rec.Code != http.StatusOK {
		t.Errorf("postEncode() expected %v, got %v", http.StatusOK, rec.Code)
	}

	htmlResp2, err := html.Parse(resp2.Body)
	if err != nil {
		//TODO
		t.Errorf("html parse error: %v", err)
	}

	errTag2 := getElementById(htmlResp2, "errMsg")
	if errTag2 != nil {
		nodeOutput := renderNode(errTag2)
		t.Errorf("postEncode() should not have returned an error message, but it did: %v", nodeOutput)
	}

	aTag2 := getElementById(htmlResp2, "a")
	if aTag2 == nil {
		t.Errorf("postEncode() did not return the map correctly")
	} else {
		nodeOutput := renderNode(aTag2)
		if !strings.Contains(nodeOutput, "value=\"a\"") {
			t.Errorf("postEncode() should have returned a a, instead returned: %v", nodeOutput)
		}
	}

	// try bad secret
	form3 := url.Values{}
	form3.Add("pathPass", "thishouldfail")

	req = httptest.NewRequest(http.MethodPost, "/testpath/save", nil)
	req.Form = form3
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	resp3 := rec.Result()
	if rec.Code != http.StatusOK {
		t.Errorf("postEncode() expected %v, got %v", http.StatusOK, rec.Code)
	}

	htmlResp3, err := html.Parse(resp3.Body)
	if err != nil {
		//TODO
		t.Errorf("html parse error: %v", err)
	}

	errTag3 := getElementById(htmlResp3, "errMsg")
	if errTag3 == nil {
		t.Errorf("postEncode() should have returned an error message, but it didn't")
	}

	// try claiming an unclaimed page
	form4 := url.Values{}
	form4.Add("pathPass", "newpass")

	req = httptest.NewRequest(http.MethodPost, "/unclaimedpath/save", nil)
	req.Form = form4
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	resp4 := rec.Result()
	if rec.Code != http.StatusOK {
		t.Errorf("postEncode() expected %v, got %v", http.StatusOK, rec.Code)
	}
	body4, _ := io.ReadAll(resp4.Body)

	if !strings.Contains(string(body4), "This page is claimed") {
		t.Errorf("postCode() appears to have returned the incorrect page")
	}
}

// HELPERS
func getAttribute(n *html.Node, key string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)

	err := html.Render(w, n)
	if err != nil {
		return ""
	}
	return buf.String()
}

func checkId(n *html.Node, id string) bool {
	if n.Type == html.ElementNode {
		s, ok := getAttribute(n, "id")
		if ok && s == id {
			return true
		}
	}
	return false
}

func traverse(n *html.Node, id string) *html.Node {
	if checkId(n, id) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res := traverse(c, id)
		if res != nil {
			return res
		}
	}
	return nil
}

func getElementById(n *html.Node, id string) *html.Node {
	return traverse(n, id)
}
