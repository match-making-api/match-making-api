package rest_api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/cmd/rest-api/routing"
	"github.com/leet-gaming/match-making-api/pkg/infra/ioc"
)

type Tester struct {
	Container      container.Container
	RequestHandler http.Handler
}

func NewTester() *Tester {
	os.Setenv("DEV_ENV", "test")
	os.Setenv("MONGO_URI", "mongodb://127.0.0.1:37019/matchmaking_test")
	os.Setenv("MONGO_DB_NAME", "matchmaking_test")
	os.Setenv("STEAM_VHASH_SOURCE", "82DA0F0D0135FEA0F5DDF6F96528B48A")

	b := ioc.NewContainerBuilder().WithEnvFile().With(ioc.InjectMongoDB).WithInboundPorts()
	c := b.Build()
	return &Tester{
		Container:      c,
		RequestHandler: routing.NewRouter(context.Background(), c),
	}
}

func (t *Tester) Exec(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()

	// if len(req.Header.Get(string(common.ResourceOwnerIDParamKey))) > 0 {

	// }

	t.RequestHandler.ServeHTTP(rec, req)

	return rec

	// authReq := http.NewRequest()

}

func expectStatus(t *testing.T, expected int, r *httptest.ResponseRecorder) {
	if expected != r.Code {
		t.Errorf("Expected response code %d. Got %d\n Body=%v", expected, r.Code, r.Body)
	}
}

func expectUUIDHeader(t *testing.T, key string, r *httptest.ResponseRecorder) {
	if len(r.Header().Get(key)) == 0 {
		t.Errorf("Expected response header %s to be a valid UUID. Got %s", key, r.Header().Get(key))
	}

	uuidRegEx := `^[\w\d]{8}-[\w\d]{4}-[\w\d]{4}-[\w\d]{4}-[\w\d]{12}$`

	if !regexp.MustCompile(uuidRegEx).MatchString(r.Header().Get(key)) {
		t.Errorf("Expected response header %s to be a UUID. Got %s", key, r.Header().Get(key))
	}
}
