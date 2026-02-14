package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/render"
	"github.com/stretchr/testify/suite"

	coretesting "github.com/bbsbb/go-edge/core/testing"
)

type RenderSuite struct {
	suite.Suite
}

func (s *RenderSuite) TestNoOpBinder_SatisfiesInterface() {
	var _ render.Binder = NoOpBinder{}
}

func (s *RenderSuite) TestNoOpRenderer_SatisfiesInterface() {
	var _ render.Renderer = &NoOpRenderer{}
}

func (s *RenderSuite) TestNoOpBinder_JSONMarshalNoExtraFields() {
	type req struct {
		NoOpBinder
		Name string `json:"name"`
	}
	b, err := json.Marshal(req{Name: "test"})
	s.Require().NoError(err)
	s.Assert().JSONEq(`{"name":"test"}`, string(b))
}

func (s *RenderSuite) TestNoOpRenderer_JSONMarshalNoExtraFields() {
	type resp struct {
		NoOpRenderer
		Name string `json:"name"`
	}
	b, err := json.Marshal(resp{Name: "test"})
	s.Require().NoError(err)
	s.Assert().JSONEq(`{"name":"test"}`, string(b))
}

type failRenderer struct{}

func (failRenderer) Render(http.ResponseWriter, *http.Request) error {
	return errors.New("render failed")
}

func (s *RenderSuite) TestRenderOrLog_LogsOnError() {
	lc := coretesting.NewLogCapture(&s.Suite)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	RenderOrLog(rr, req, failRenderer{}, lc.Logger)

	output := lc.Output()
	s.Assert().Contains(output, "failed to render response")
	s.Assert().Contains(output, "render failed")
}

func TestRenderSuite(t *testing.T) {
	suite.Run(t, new(RenderSuite))
}
