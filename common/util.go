package common

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"go.uber.org/zap"
)

func MakeHttpRequest(method string, url string, body io.Reader, queryParams map[string]string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if len(queryParams) > 0 {
		q := req.URL.Query()

		for key, param := range queryParams {
			q.Add(key, param)
		}

		req.URL.RawQuery = q.Encode()
	}

	if len(headers) > 0 {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 300 {
		return nil, err
		//TODO vielleicht noch die message returnen
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func GetLogger() *zap.SugaredLogger {
	loglevel := os.Getenv("LOG_LEVEL")

	var l *zap.Logger

	if loglevel == "prod" {
		l, _ = zap.NewProduction()
	} else {
		l = zap.NewExample()
	}

	defer l.Sync()
	return l.Sugar()
}
