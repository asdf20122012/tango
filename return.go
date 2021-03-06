package tango

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"reflect"
)

const (
	AutoResponse = iota
	JsonResponse
	XmlResponse
)

type ResponseTyper interface {
	ResponseType() int
}

type Json struct {}
func (Json) ResponseType() int {
	return JsonResponse
}

type Xml struct {}
func (Xml) ResponseType() int {
	return XmlResponse
}

func isNil(a interface{}) bool {
	if a == nil {
		return true
	}
	aa := reflect.ValueOf(a)
	return !aa.IsValid() || (aa.Type().Kind() == reflect.Ptr && aa.IsNil())
}

type XmlError struct {
	XMLName   xml.Name `xml:"err"`
	Content string `xml:"content"`
}

func ReturnHandler(ctx *Context) {
	var rt int
	if action := ctx.Action(); action != nil {
		if i, ok := action.(ResponseTyper); ok {
			rt = i.ResponseType()
		}
	}

	ctx.Next()

	// if has been write, then return
	if ctx.Written() {
		return
	}

	if isNil(ctx.Result) {
		if ctx.Action() == nil {
			// if there is no action match
			ctx.Result = NotFound()
		} else {
			// there is an action but return nil, then we return blank page
			ctx.Result = ""
		}
	}

	if rt == JsonResponse {
		encoder := json.NewEncoder(ctx)
		ctx.Header().Set("Content-Type", "application/json")
		if e, ok := ctx.Result.(error); ok {
			encoder.Encode(map[string]string{
				"err": e.Error(),
			})
		} else {
			err := encoder.Encode(ctx.Result)
			if err != nil {
				ctx.Result = err
				encoder.Encode(map[string]string{
					"err": err.Error(),
				})
			}
		}
		return
	} else if rt == XmlResponse {
		encoder := xml.NewEncoder(ctx)
		ctx.Header().Set("Content-Type", "application/xml")
		if e, ok := ctx.Result.(error); ok {
			encoder.Encode(XmlError{
				Content: e.Error(),
			})
		} else {
			err := encoder.Encode(ctx.Result)
			if err != nil {
				ctx.Result = err
				encoder.Encode(XmlError{
					Content: err.Error(),
				})
			}
		}
		return
	}

	switch res := ctx.Result.(type) {
	case AbortError:
		ctx.WriteHeader(res.Code())
		ctx.Write([]byte(res.Error()))
	case error:
		ctx.WriteHeader(http.StatusInternalServerError)
		ctx.Write([]byte(res.Error()))
	case []byte:
		ctx.WriteHeader(http.StatusOK)
		ctx.Write(res)
	case string:
		ctx.WriteHeader(http.StatusOK)
		ctx.Write([]byte(res))
	}
}