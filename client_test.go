package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

////////////////////////////////////////

func TestRespositoryRootReturnsRootWhenThere(t *testing.T) {
	// setup
	pwd, _ := os.Getwd()
	if err := os.MkdirAll("test/artifacts/.amber", 0700); err != nil {
		t.Errorf("Error creating artifacts: ", err)
	}
	if err := os.Chdir("test/artifacts"); err != nil {
		t.Errorf("Error changing directory: ", err)
	}
	defer os.RemoveAll("test/artifacts")
	defer os.Chdir(pwd)

	// test
	actual, err := repositoryRoot(".amber")
	if err != nil {
		t.Error(err)
	}

	// verify
	var expected = filepath.Join(pwd, "test", "artifacts", ".amber")
	if actual != expected {
		t.Errorf("Data mismatch:\n   actual: [%s]\n expected: [%s]\n", actual, expected)
	}
}

func TestRespositoryRootReturnsRootWhenBelow(t *testing.T) {
	// setup
	pwd, _ := os.Getwd()
	if err := os.MkdirAll("test/artifacts/.amber", 0700); err != nil {
		t.Errorf("Error creating artifacts: ", err)
	}
	if err := os.MkdirAll("test/artifacts/foo/bar", 0700); err != nil {
		t.Errorf("Error creating artifacts: ", err)
	}
	if err := os.Chdir("test/artifacts/foo/bar"); err != nil {
		t.Errorf("Error changing directory: ", err)
	}
	defer os.RemoveAll("test/artifacts")
	defer os.Chdir(pwd)

	// test
	actual, err := repositoryRoot(".amber")
	if err != nil {
		t.Error(err)
	}

	// verify
	var expected = filepath.Join(pwd, "test", "artifacts", ".amber")
	if actual != expected {
		t.Errorf("Data mismatch:\n   actual: [%s]\n expected: [%s]\n", actual, expected)
	}
}

func TestRespositoryRootReturnsErrorWhenDotAmberIsFile(t *testing.T) {
	// setup
	pwd, _ := os.Getwd()
	if err := os.MkdirAll("test/artifacts/foo/bar", 0700); err != nil {
		t.Errorf("Error creating artifacts: ", err)
	}
	if err := ioutil.WriteFile("test/artifacts/.amber", []byte{}, 0600); err != nil {
		t.Error(err)
	}
	if err := os.Chdir("test/artifacts/foo/bar"); err != nil {
		t.Errorf("Error changing directory: ", err)
	}
	defer os.RemoveAll("test/artifacts")
	defer os.Chdir(pwd)

	// test
	_, err := repositoryRoot(".amber")

	// verify
	if err == nil {
		t.Errorf("expected error %v", ErrNoRepos)
	}
}

func TestRespositoryRootReturnsErrorWhenNotFound(t *testing.T) {
	// setup
	pwd, _ := os.Getwd()
	if err := os.MkdirAll("test/artifacts/foo/bar", 0700); err != nil {
		t.Errorf("Error creating artifacts: ", err)
	}
	if err := os.Chdir("test/artifacts/foo/bar"); err != nil {
		t.Errorf("Error changing directory: ", err)
	}
	defer os.RemoveAll("test/artifacts")
	defer os.Chdir(pwd)

	// test
	repos, err := repositoryRoot(".amber-should-never-see")

	// verify
	if repos != "" {
		t.Errorf("expected: %v, actual: %v", "", repos)
	}
	if err == nil {
		t.Errorf("expected error %v", ErrNoRepos)
	}
}

////////////////////////////////////////

func TestEncryptReturnsErrorWhenUnknownEncryption(t *testing.T) {
	iv := make([]byte, 10)
	actual, err := encrypt(make([]byte, 10), "non-existant-algorithm", "some key", iv)
	if actual != nil {
		t.Errorf("expected: %v, actual: %v", nil, actual)
	}
	if err == nil {
		t.Errorf("expected: %v, actual: %v", fmt.Errorf("unknown encryption algorithm: non-existant-algorithm"), err)
	}
}

func TestEncryptReturnsEncryptionStringOfBytes(t *testing.T) {
	bytes := []byte("just some blob of data")
	iv := make([]byte, 10)
	actual, err := encrypt(bytes, "-", "some key", iv)
	expected := "just some blob of data"
	if string(actual) != expected {
		t.Errorf("expected: %v, actual: %v", expected, string(actual))
	}
	if err != nil {
		t.Errorf("expected: %v, actual: %v", nil, err)
	}
}

////////////////////////////////////////

func TestDecryptReturnsErrorWhenUnknownDecryption(t *testing.T) {
	iv := make([]byte, 10)
	actual, err := decrypt(make([]byte, 10), "non-existant-decryption", "some key", iv)
	if actual != nil {
		t.Errorf("expected: %v, actual: %v", nil, actual)
	}
	if err == nil {
		t.Errorf("expected: %v, actual: %v", fmt.Errorf("unknown decryption algorithm: non-existant-algorithm"), err)
	}
}

func TestDecryptReturnsDecryptionStringOfBytes(t *testing.T) {
	bytes := []byte("just some blob of data")
	iv := make([]byte, 10)
	actual, err := decrypt(bytes, "-", "some key", iv)
	expected := "just some blob of data"
	if string(actual) != expected {
		t.Errorf("expected: %v, actual: %v", expected, string(actual))
	}
	if err != nil {
		t.Errorf("expected: %v, actual: %v", nil, err)
	}
}

////////////////////////////////////////

func TestSelectIVInvalidHashName(t *testing.T) {
	plaintext := []byte("this is a test")
	_, actual := selectIV("aes128", "sha", plaintext)
	expected := fmt.Sprintf("unknown hash: %s", "sha")
	if actual.Error() != expected {
		t.Errorf("expected: %v, actual: %v", expected, actual.Error())
	}
}

func TestSelectIVInvalidEncryptionAlgorithm(t *testing.T) {
	plaintext := []byte("this is a test")

	cases := []map[string]string{
		{
			"eName": "-",
			"msg":   "unknown encryption algorithm: -",
		},
		{
			"eName": "aes",
			"msg":   "unknown encryption algorithm: aes",
		},
	}
	for _, item := range cases {
		_, actual := selectIV(item["eName"], "sha1", plaintext)
		expected := fmt.Sprintf("unknown encryption algorithm: %s", item["eName"])
		if actual.Error() != expected {
			t.Errorf("expected: %v, actual: %v", expected, actual.Error())
		}
	}
}

func TestSelectIV(t *testing.T) {
	plaintext := []byte("this is our blob of plaintext, and its size will be hashed to come up with an iv.")

	cases := []map[string]string{
		{
			"eName":          "aes128", // determines output bit length
			"hName":          "sha1",   // which hash to use
			"ivSizeExpected": "16",
			"ivExpected":     "31643531336330626362653333623265",
		},
		{
			"eName":          "aes192",
			"hName":          "sha1",
			"ivSizeExpected": "24",
			"ivExpected":     "316435313363306263626533336232653734343065356531",
		},
		{
			"eName":          "aes256",
			"hName":          "sha1",
			"ivSizeExpected": "32",
			"ivExpected":     "3164353133633062636265333362326537343430653565313464306232326566",
		},
		{
			"eName":          "aes128",
			"hName":          "sha256",
			"ivSizeExpected": "16",
			"ivExpected":     "35333136636131633564646361386536",
		},
		{
			"eName":          "aes128",
			"hName":          "sha512",
			"ivSizeExpected": "16",
			"ivExpected":     "61346133636436616432376230613539",
		},
	}

	for _, item := range cases {
		iv, err := selectIV(item["eName"], item["hName"], plaintext)
		if err != nil {
			t.Errorf("expected: %v, actual: %v", nil, err)
		}

		if ivSizeActual := fmt.Sprintf("%d", len(iv)); ivSizeActual != item["ivSizeExpected"] {
			t.Errorf("expected: %v, actual: %v", item["ivSizeExpected"], ivSizeActual)
		}

		if ivActual := fmt.Sprintf("%x", iv); ivActual != item["ivExpected"] {
			t.Errorf("expected: %v, actual: %v", item["ivExpected"], ivActual)
		}
	}
}

////////////////////////////////////////

func TestParseUriList(t *testing.T) {
	uriList := "# this is a comment\r\nhttp://example.com/1\r\nhttp://example.com/2"
	expected := []string{
		"http://example.com/1",
		"http://example.com/2",
	}
	actual := parseUriList(uriList)
	if len(expected) != len(actual) {
		t.Errorf("expected: %v, actual: %v", len(expected), len(actual))
	}
	for i, _ := range expected {
		if expected[i] != actual[i] {
			t.Errorf("expected: %v, actual: %v", expected[i], actual[i])
		}
	}
}

type parseUrcCase struct {
	name		string
	blob		[]byte
	output		metadata
	err			error
}

func TestParseUrcCatchesErrors(t *testing.T) {
	cases := []parseUrcCase{
		{
			name:	"three fields",
			blob:	[]byte("one two three"),
			output:	metadata{},
			err:	fmt.Errorf("invalid line format: one two three"),
		},
		{
			name:	"splits on crlf",
			blob:	[]byte("X-Amber-Hash: foo\nX-Amber-Encryption: bar\n"),
			output:	metadata{},
			err:	errors.New("invalid line format: X-Amber-Hash: foo\nX-Amber-Encryption: bar\n"),
		},
	}
	for _, item := range cases {
		output, err := parseUrc(item.blob)
		if item.err.Error() != err.Error() {
			t.Errorf("Case: %v; Expected error: %v; Acutal error: %v\n", item.name, item.err.Error(), err.Error())
		}
		if fmt.Sprintf("%#v", item.output) != fmt.Sprintf("%#v", output) {
			t.Errorf("Case: %v; Expected: %#v; Acutal: %#v\n", item.name, item.output, output)
		}
	}
}

func TestParseUrcExpectedResults(t *testing.T) {
	cases := []parseUrcCase{
		{
			name:	"gets hash name",
			blob:	[]byte("X-Amber-Hash: foo"),
			output:	metadata{hName: "foo"},
			err:	nil,
		},
		{
			name:	"gets encryption name",
			blob:	[]byte("X-Amber-Encryption: bar"),
			output:	metadata{eName: "bar"},
			err:	nil,
		},
		{
			name:	"gets hash and encryption name",
			blob:	[]byte("X-Amber-Hash: foo\r\nX-Amber-Encryption: bar\r\n"),
			output:	metadata{hName: "foo", eName: "bar"},
			err:	nil,
		},
		{
			name:	"stops at empty line",
			blob:	[]byte("X-Amber-Hash: foo\r\n\r\nX-Amber-Encryption: bar\r\n"),
			output:	metadata{hName: "foo"},
			err:	nil,
		},
	}
	for _, item := range cases {
		output, err := parseUrc(item.blob)
		if err != nil {
			t.Errorf("Case: %v; Didn't expect error: %v\n", item.name, err.Error())
		}
		if fmt.Sprintf("%#v", item.output) != fmt.Sprintf("%#v", output) {
			t.Errorf("Case: %v; Expected: %#v; Acutal: %#v\n", item.name, item.output, output)
		}
	}
}
