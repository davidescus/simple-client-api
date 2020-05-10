package nearearth

import (
	"io/ioutil"
	"net/http"
)

type client struct{}

type response struct {
	statusCode int
	body       []byte
}

func (c client) getData(url string) (response, error) {
	// create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response{}, err
	}

	// create client and execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return response{}, err
	}

	// get body
	// TODO if body is too large RealAll will fail (think about memory)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response{}, err
	}

	response := response{
		statusCode: resp.StatusCode,
		body:       body,
	}

	return response, nil
}
