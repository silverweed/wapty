package intercept

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

type mockRT struct {
	//This is filled with the interceptor req
	req *http.Request
	//Mock return values
	err error
	res *http.Response
}

func (mr *mockRT) RoundTrip(req *http.Request) (res *http.Response, err error) {
	defer func() { mr.res = nil }()
	mr.req = req
	res, err = mr.res, mr.err
	return
}

func mockhandleResponse(presp *pendingResponse) {}

var rtTests = []struct {
	//Input data
	in              *http.Request
	subj            mockRT
	interceptStatus bool
	reqModifier     func(*pendingRequest)
	resModifier     func(*pendingResponse)
	//Expected values
	eEditedRequest *http.Request
	eOut           *http.Response
	eError         error
}{
	//TODO
	{
		in:              &http.Request{URL: &url.URL{}, Header: http.Header{}},
		subj:            mockRT{res: &http.Response{ContentLength: 3, Body: ioutil.NopCloser(bytes.NewReader([]byte(`foo`)))}},
		interceptStatus: false,
		eEditedRequest:  &http.Request{URL: &url.URL{}},
		eOut:            &http.Response{ContentLength: 3, Body: ioutil.NopCloser(bytes.NewReader([]byte(`foo`)))},
	},
}

//This is more of an integration test
func TestRoundTrip(t *testing.T) {
	beforeHopTest()
	defer afterHopTest()
	for i, tt := range rtTests {
		ri := &Interceptor{wrappedRT: &tt.subj}
		intercept.SetValue(tt.interceptStatus)
		if intercept.Value() {
			if tt.reqModifier != nil {
				go tt.reqModifier(<-RequestQueue)
			} else {
				go func(p *pendingRequest) {
					p.modifiedRequest <- &mayBeRequest{req: p.originalRequest}
				}(<-RequestQueue)
			}
			if tt.resModifier != nil {
				go tt.resModifier(<-ResponseQueue)
			} else {
				go func(p *pendingResponse) {
					p.modifiedResponse <- &mayBeResponse{res: p.originalResponse}
				}(<-ResponseQueue)
			}
		}
		out, err := ri.RoundTrip(tt.in)
		if err != tt.eError {
			t.Errorf("Test %d failed, errors differ: wanted %v got %v", i, tt.eError, err)
		}
		if pass := reqEqual(tt.subj.req, tt.eEditedRequest); !pass {
			t.Errorf("Test %d failed, requests differ: wanted \n%+v\n got \n%+v", i, tt.subj.req, tt.eEditedRequest)
		}
		if pass, jout, jeout := jsonEqual(out, tt.eOut); !pass {
			t.Errorf("Test %d failed, responses differ: wanted \n%s\n got \n%s", i, jeout, jout)
		}
	}
}

func beforeHopTest() {}
func afterHopTest()  {}
