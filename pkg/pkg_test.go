package pkg_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/ya-breeze/httpscripter/pkg"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Package", func() {
	var testserver *httptest.Server

	BeforeEach(func() {
		cnt := 3
		testserver = httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "POST" {
					_, _ = w.Write([]byte(fmt.Sprintf(`{"id":%d}`, cnt)))
					return
				}

				if cnt != 0 {
					_, _ = w.Write([]byte(fmt.Sprintf(`{"id":%d, "status":"processing"}`, cnt)))
					cnt--
					return
				}

				_, _ = w.Write([]byte(fmt.Sprintf(`{"id":%d, "status":"complete", "data":{"valid":true}}`, cnt)))
			}))
		pkg.Last.BaseURL = testserver.URL
	})

	It("allows complex flows", func() {
		pkg.POST(
			"/post",
			pkg.JSON(map[string]interface{}{
				"name": "Jo\"hn",
				"o": map[string]interface{}{
					"a": 1,
					"b": 2,
					"c": true,
				},
			}),
		)
		for {
			pkg.GET("/get")
			Expect(pkg.Succeed(pkg.Last.Response.StatusCode)).To(BeTrue())
			Expect(pkg.Failed(pkg.Last.Response.StatusCode)).To(BeFalse())

			if pkg.Value("status").String() == "complete" {
				break
			}
			Expect(pkg.Value("id").Int()).To(BeNumerically(">", 0))
		}
	})

	AfterEach(func() {
		testserver.Close()
	})
})
