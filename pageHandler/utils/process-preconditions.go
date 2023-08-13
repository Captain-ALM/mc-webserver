package utils

import (
	"golang.captainalm.com/mc-webserver/utils/io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

func ProcessSupportedPreconditionsForNext(rw http.ResponseWriter, req *http.Request, modT time.Time, etag string, noBypassModify bool, noBypassMatch bool) bool {
	theStrippedETag := GetETagValue(etag)
	if noBypassMatch && theStrippedETag != "" && req.Header.Get("If-None-Match") != "" {
		etagVals := GetETagValues(req.Header.Get("If-None-Match"))
		conditionSuccess := false
		for _, s := range etagVals {
			if s == theStrippedETag {
				conditionSuccess = true
				break
			}
		}
		if conditionSuccess {
			WriteResponseHeaderCanWriteBody(req.Method, rw, http.StatusNotModified, "")
			return false
		}
	}

	if noBypassMatch && theStrippedETag != "" && req.Header.Get("If-Match") != "" {
		etagVals := GetETagValues(req.Header.Get("If-Match"))
		conditionFailed := true
		for _, s := range etagVals {
			if s == theStrippedETag {
				conditionFailed = false
				break
			}
		}
		if conditionFailed {
			SwitchToNonCachingHeaders(rw.Header())
			rw.Header().Del("Content-Type")
			rw.Header().Del("Content-Length")
			WriteResponseHeaderCanWriteBody(req.Method, rw, http.StatusPreconditionFailed, "")
			return false
		}
	}

	if noBypassModify && !modT.IsZero() && req.Header.Get("If-Modified-Since") != "" {
		parse, err := time.Parse(http.TimeFormat, req.Header.Get("If-Modified-Since"))
		if err == nil && modT.Before(parse) || strings.EqualFold(modT.Format(http.TimeFormat), req.Header.Get("If-Modified-Since")) {
			WriteResponseHeaderCanWriteBody(req.Method, rw, http.StatusNotModified, "")
			return false
		}
	}

	if noBypassModify && !modT.IsZero() && req.Header.Get("If-Unmodified-Since") != "" {
		parse, err := time.Parse(http.TimeFormat, req.Header.Get("If-Unmodified-Since"))
		if err == nil && modT.After(parse) {
			SwitchToNonCachingHeaders(rw.Header())
			rw.Header().Del("Content-Type")
			rw.Header().Del("Content-Length")
			WriteResponseHeaderCanWriteBody(req.Method, rw, http.StatusPreconditionFailed, "")
			return false
		}
	}

	return true
}

func ProcessRangePreconditions(maxLength int64, rw http.ResponseWriter, req *http.Request, modT time.Time, etag string, supported bool) []ContentRangeValue {
	canDoRange := supported
	theStrippedETag := GetETagValue(etag)
	modTStr := modT.Format(http.TimeFormat)

	if canDoRange {
		rw.Header().Set("Accept-Ranges", "bytes")
	}

	if canDoRange && !modT.IsZero() && strings.HasSuffix(req.Header.Get("If-Range"), "GMT") {
		newModT, err := time.Parse(http.TimeFormat, modTStr)
		parse, err := time.Parse(http.TimeFormat, req.Header.Get("If-Range"))
		if err == nil && !newModT.Equal(parse) {
			canDoRange = false
		}
	} else if canDoRange && theStrippedETag != "" && req.Header.Get("If-Range") != "" {
		if GetETagValue(req.Header.Get("If-Range")) != theStrippedETag {
			canDoRange = false
		}
	}

	if canDoRange && strings.HasPrefix(req.Header.Get("Range"), "bytes=") {
		if theRanges := GetRanges(req.Header.Get("Range"), maxLength); len(theRanges) != 0 {
			if len(theRanges) == 1 {
				rw.Header().Set("Content-Length", strconv.FormatInt(theRanges[0].Length, 10))
				rw.Header().Set("Content-Range", theRanges[0].ToField(maxLength))
			} else {
				theSize := GetMultipartLength(theRanges, rw.Header().Get("Content-Type"), maxLength)
				rw.Header().Set("Content-Length", strconv.FormatInt(theSize, 10))
			}
			if WriteResponseHeaderCanWriteBody(req.Method, rw, http.StatusPartialContent, "") {
				return theRanges
			} else {
				return nil
			}
		} else {
			SwitchToNonCachingHeaders(rw.Header())
			rw.Header().Del("Content-Type")
			rw.Header().Del("Content-Length")
			rw.Header().Set("Content-Range", "bytes */"+strconv.FormatInt(maxLength, 10))
			WriteResponseHeaderCanWriteBody(req.Method, rw, http.StatusRequestedRangeNotSatisfiable, "")
			return nil
		}
	}
	if WriteResponseHeaderCanWriteBody(req.Method, rw, http.StatusOK, "") {
		return make([]ContentRangeValue, 0)
	}
	return nil
}

func GetMultipartLength(parts []ContentRangeValue, contentType string, maxLength int64) int64 {
	cWriter := &io.CountingWriter{Length: 0}
	var returnLength int64 = 0
	multWriter := multipart.NewWriter(cWriter)
	for _, currentPart := range parts {
		_, _ = multWriter.CreatePart(textproto.MIMEHeader{
			"Content-Range": {currentPart.ToField(maxLength)},
			"Content-Type":  {contentType},
		})
		returnLength += currentPart.Length
	}
	_ = multWriter.Close()
	returnLength += cWriter.Length
	return returnLength
}

func WriteResponseHeaderCanWriteBody(method string, rw http.ResponseWriter, statusCode int, message string) bool {
	hasBody := method != http.MethodHead && method != http.MethodOptions
	if hasBody && message != "" {
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.Header().Set("X-Content-Type-Options", "nosniff")
		rw.Header().Set("Content-Length", strconv.Itoa(len(message)+2))
		SetNeverCacheHeader(rw.Header())
	}
	rw.WriteHeader(statusCode)
	if hasBody {
		if message != "" {
			_, _ = rw.Write([]byte(message + "\r\n"))
			return false
		}
		return true
	}
	return false
}
