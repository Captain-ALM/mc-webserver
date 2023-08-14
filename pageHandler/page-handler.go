package pageHandler

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.captainalm.com/mc-webserver/conf"
	"golang.captainalm.com/mc-webserver/pageHandler/utils"
)

const indexName = "index.go"

type PageHandler struct {
	PageContentsCache        map[string]*CachedPage
	PageProviders            map[string]PageProvider
	pageContentsCacheRWMutex *sync.RWMutex
	RangeSupported           bool
	CacheSettings            conf.CacheSettingsYaml
}

type CachedPage struct {
	Content     []byte
	ContentType string
	LastMod     time.Time
}

func NewPageHandler(config conf.ServeYaml) *PageHandler {
	var thePCCMap map[string]*CachedPage
	var theMutex *sync.RWMutex
	if config.CacheSettings.EnableContentsCaching {
		thePCCMap = make(map[string]*CachedPage)
		theMutex = &sync.RWMutex{}
	}
	toReturn := &PageHandler{
		PageContentsCache:        thePCCMap,
		pageContentsCacheRWMutex: theMutex,
		RangeSupported:           config.RangeSupported,
		CacheSettings:            config.CacheSettings,
	}
	if config.EnableGoInfoPage {
		toReturn.PageProviders = GetProviders(config.CacheSettings.EnableTemplateCaching, config.DataStorage, toReturn)
	} else {
		toReturn.PageProviders = GetProviders(config.CacheSettings.EnableTemplateCaching, config.DataStorage, nil)
	}
	return toReturn
}

func (ph *PageHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	actualPagePath := ""
	if strings.HasSuffix(request.URL.Path, "/") {
		if strings.HasSuffix(request.URL.Path, ".go/") {
			actualPagePath = strings.TrimRight(request.URL.Path, "/")
		} else {
			actualPagePath = request.URL.Path + indexName
		}
	} else {
		actualPagePath = request.URL.Path
	}

	var currentProvider PageProvider
	canCache := false
	actualQueries := ""
	queryValues := request.URL.Query()
	var pageContent []byte
	pageContentType := ""
	var lastMod time.Time

	if currentProvider = ph.PageProviders[actualPagePath]; currentProvider != nil {
		actualQueries = currentProvider.GetCacheIDExtension(queryValues)

		if ph.CacheSettings.EnableContentsCaching {
			cached := ph.getPageFromCache(request.URL.Path, actualQueries)
			if cached != nil {
				pageContent = cached.Content
				pageContentType = cached.ContentType
				lastMod = cached.LastMod
			}
		}

		if pageContentType == "" {
			pageContentType, pageContent, canCache = currentProvider.GetContents(queryValues)
			lastMod = currentProvider.GetLastModified()
			if pageContentType != "" && canCache && ph.CacheSettings.EnableContentsCaching {
				ph.setPageToCache(request.URL.Path, actualQueries, &CachedPage{
					Content:     pageContent,
					ContentType: pageContentType,
					LastMod:     lastMod,
				})
			}
		}
	}

	allowedMethods := ph.getAllowedMethodsForPath(request.URL.Path)
	allowed := false
	if request.Method != http.MethodOptions {
		for _, method := range allowedMethods {
			if method == request.Method {
				allowed = true
				break
			}
		}
	}

	if allowed {

		if pageContentType == "" {
			utils.WriteResponseHeaderCanWriteBody(request.Method, writer, http.StatusNotFound, "Page Not Found")
		} else {

			switch request.Method {
			case http.MethodGet, http.MethodHead:

				writer.Header().Set("Content-Type", pageContentType)
				writer.Header().Set("Content-Length", strconv.Itoa(len(pageContent)))
				utils.SetLastModifiedHeader(writer.Header(), lastMod)
				utils.SetCacheHeaderWithAge(writer.Header(), ph.CacheSettings.MaxAge, lastMod)
				theETag := utils.GetValueForETagUsingByteArray(pageContent)
				writer.Header().Set("ETag", theETag)

				if utils.ProcessSupportedPreconditionsForNext(writer, request, lastMod, theETag, ph.CacheSettings.NotModifiedResponseUsingLastModified, ph.CacheSettings.NotModifiedResponseUsingETags) {

					httpRangeParts := utils.ProcessRangePreconditions(int64(len(pageContent)), writer, request, lastMod, theETag, ph.RangeSupported)
					if httpRangeParts != nil {
						if len(httpRangeParts) <= 1 {
							var theWriter io.Writer = writer
							if len(httpRangeParts) == 1 {
								theWriter = utils.NewPartialRangeWriter(theWriter, httpRangeParts[0])
							}
							_, _ = theWriter.Write(pageContent)
						} else {
							multWriter := multipart.NewWriter(writer)
							writer.Header().Set("Content-Type", "multipart/byteranges; boundary="+multWriter.Boundary())
							for _, currentPart := range httpRangeParts {
								mimePart, err := multWriter.CreatePart(textproto.MIMEHeader{
									"Content-Range": {currentPart.ToField(int64(len(pageContent)))},
									"Content-Type":  {"text/plain; charset=utf-8"},
								})
								if err != nil {
									break
								}
								_, err = mimePart.Write(pageContent[currentPart.Start : currentPart.Start+currentPart.Length])
								if err != nil {
									break
								}
							}
							_ = multWriter.Close()
						}
					}
				}
			case http.MethodDelete:
				ph.PurgeTemplateCache(actualPagePath, request.URL.Path == "/")
				ph.PurgeContentsCache(request.URL.Path, actualQueries)
				utils.SetNeverCacheHeader(writer.Header())
				utils.WriteResponseHeaderCanWriteBody(request.Method, writer, http.StatusOK, "")
			}
		}
	} else {

		theAllowHeaderContents := ""
		for _, method := range allowedMethods {
			theAllowHeaderContents += method + ", "
		}

		writer.Header().Set("Allow", strings.TrimSuffix(theAllowHeaderContents, ", "))
		if request.Method == http.MethodOptions {
			utils.WriteResponseHeaderCanWriteBody(request.Method, writer, http.StatusOK, "")
		} else {
			utils.WriteResponseHeaderCanWriteBody(request.Method, writer, http.StatusMethodNotAllowed, "")
		}
	}
}

func (ph *PageHandler) PurgeContentsCache(path string, query string) {
	if ph.CacheSettings.EnableContentsCaching && ph.CacheSettings.EnableContentsCachePurge {
		if path == "/" {
			ph.pageContentsCacheRWMutex.Lock()
			ph.PageContentsCache = make(map[string]*CachedPage)
			ph.pageContentsCacheRWMutex.Unlock()
		} else {
			if strings.HasSuffix(path, ".go/") {
				ph.pageContentsCacheRWMutex.RLock()
				toDelete := make([]string, len(ph.PageContentsCache))
				theSize := 0
				for cPath := range ph.PageContentsCache {
					dPath := strings.Split(cPath, "?")[0]
					if dPath == path || dPath == path[:len(path)-1] {
						toDelete[theSize] = cPath
						theSize++
					}
				}
				ph.pageContentsCacheRWMutex.RUnlock()
				ph.pageContentsCacheRWMutex.Lock()
				for i := 0; i < theSize; i++ {
					delete(ph.PageContentsCache, toDelete[i])
				}
				ph.pageContentsCacheRWMutex.Unlock()
				return
			} else if strings.HasSuffix(path, "/") {
				path += indexName
			}
			ph.pageContentsCacheRWMutex.Lock()
			if query == "" {
				delete(ph.PageContentsCache, path)
			} else {
				delete(ph.PageContentsCache, path+"?"+query)
			}
			ph.pageContentsCacheRWMutex.Unlock()
		}
	}
}

func (ph *PageHandler) PurgeTemplateCache(path string, all bool) {
	if ph.CacheSettings.EnableTemplateCaching && ph.CacheSettings.EnableTemplateCachePurge {
		if all {
			for _, pageProvider := range ph.PageProviders {
				pageProvider.PurgeTemplate()
			}
		} else {
			if pageProvider, ok := ph.PageProviders[path]; ok {
				pageProvider.PurgeTemplate()
			}
		}
	}
}
func (ph *PageHandler) getPageFromCache(pathIn string, cleanedQueries string) *CachedPage {
	ph.pageContentsCacheRWMutex.RLock()
	defer ph.pageContentsCacheRWMutex.RUnlock()
	if strings.HasSuffix(pathIn, ".go/") {
		return ph.PageContentsCache[strings.TrimRight(pathIn, "/")]
	} else if strings.HasSuffix(pathIn, "/") {
		pathIn += indexName
	}
	if cleanedQueries == "" {
		return ph.PageContentsCache[pathIn]
	} else {
		return ph.PageContentsCache[pathIn+"?"+cleanedQueries]
	}
}

func (ph *PageHandler) setPageToCache(pathIn string, cleanedQueries string, newPage *CachedPage) {
	ph.pageContentsCacheRWMutex.Lock()
	defer ph.pageContentsCacheRWMutex.Unlock()
	if strings.HasSuffix(pathIn, ".go/") {
		ph.PageContentsCache[strings.TrimRight(pathIn, "/")] = newPage
		return
	} else if strings.HasSuffix(pathIn, "/") {
		pathIn += indexName
	}
	if cleanedQueries == "" {
		ph.PageContentsCache[pathIn] = newPage
	} else {
		ph.PageContentsCache[pathIn+"?"+cleanedQueries] = newPage
	}
}

func (ph *PageHandler) getAllowedMethodsForPath(pathIn string) []string {
	if pathIn == "/" || strings.HasSuffix(pathIn, ".go/") {
		if (ph.CacheSettings.EnableTemplateCaching && ph.CacheSettings.EnableTemplateCachePurge) ||
			(ph.CacheSettings.EnableContentsCaching && ph.CacheSettings.EnableContentsCachePurge) {
			return []string{http.MethodHead, http.MethodGet, http.MethodOptions, http.MethodDelete}
		} else {
			return []string{http.MethodHead, http.MethodGet, http.MethodOptions}
		}
	} else {
		if ph.CacheSettings.EnableContentsCaching && ph.CacheSettings.EnableContentsCachePurge {
			return []string{http.MethodHead, http.MethodGet, http.MethodOptions, http.MethodDelete}
		} else {
			return []string{http.MethodHead, http.MethodGet, http.MethodOptions}
		}
	}
}

func (ph *PageHandler) GetRegisteredPages() []string {
	pages := make([]string, len(ph.PageProviders))
	index := 0
	for s := range ph.PageProviders {
		pages[index] = s
		index++
	}
	return pages
}

func (ph *PageHandler) GetCachedPages() []string {
	if ph.pageContentsCacheRWMutex == nil {
		return make([]string, 0)
	}
	ph.pageContentsCacheRWMutex.RLock()
	defer ph.pageContentsCacheRWMutex.RUnlock()
	pages := make([]string, len(ph.PageContentsCache))
	index := 0
	for s := range ph.PageContentsCache {
		pages[index] = s
		index++
	}
	return pages
}

func (ph *PageHandler) GetNumberOfCachedPages() int {
	if ph.pageContentsCacheRWMutex == nil {
		return 0
	}
	ph.pageContentsCacheRWMutex.RLock()
	defer ph.pageContentsCacheRWMutex.RUnlock()
	return len(ph.PageContentsCache)
}
