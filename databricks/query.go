package databricks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/google/go-querystring/query"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (o *DBClientOption) craftRequest(ctx context.Context, requestEndpoint, method string, requestData interface{}) (*http.Request, error) {
	// Default info determined by *DBClientOption
	requestURL, err := o.getRequestURI(requestEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create requestURL")
	}

	requestHeaders := o.getDefaultHeaders()

	var requestBody []byte

	// Crafting the right URL and requestBody
	if method == "GET" {
		params, err := query.Values(requestData)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to craft query for GET request")
		}

		requestURL += "?" + params.Encode()
	} else {
		bodyBytes, err := json.Marshal(requestData)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to craft request body for %s request", method)
		}
		requestBody = bodyBytes
	}

	request, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create a HTTP request with given inputs")
	}

	// Setting databricks-specific headers
	for k, v := range requestHeaders {
		request.Header.Set(k, v)
	}

	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(request, true)
	if err != nil {
		fmt.Println(err)
	}

	logrus.Debug(string(requestDump))

	return request, nil
}
