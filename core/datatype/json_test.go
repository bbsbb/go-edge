package datatype

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type testJSONData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type JSONSuite struct {
	suite.Suite
}

func (s *JSONSuite) TestScanJSON_Bytes() {
	tests := []struct {
		name    string
		input   []byte
		want    testJSONData
		wantErr bool
	}{
		{
			name:    "valid json",
			input:   []byte(`{"name":"test","value":42}`),
			want:    testJSONData{Name: "test", Value: 42},
			wantErr: false,
		},
		{
			name:    "empty object",
			input:   []byte(`{}`),
			want:    testJSONData{},
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   []byte(`{invalid}`),
			want:    testJSONData{},
			wantErr: true,
		},
		{
			name:    "unknown field",
			input:   []byte(`{"name":"test","unknown":"field"}`),
			want:    testJSONData{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var got testJSONData
			err := ScanJSON(&got, tt.input)
			if tt.wantErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Assert().Equal(tt.want, got)
			}
		})
	}
}

func (s *JSONSuite) TestScanJSON_String() {
	tests := []struct {
		name    string
		input   string
		want    testJSONData
		wantErr bool
	}{
		{
			name:    "valid json string",
			input:   `{"name":"test","value":42}`,
			want:    testJSONData{Name: "test", Value: 42},
			wantErr: false,
		},
		{
			name:    "invalid json string",
			input:   `{invalid}`,
			want:    testJSONData{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var got testJSONData
			err := ScanJSON(&got, tt.input)
			if tt.wantErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Assert().Equal(tt.want, got)
			}
		})
	}
}

func (s *JSONSuite) TestScanJSON_InvalidType() {
	var got testJSONData
	err := ScanJSON(&got, 123)
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "invalid value type")
}

func TestJSONSuite(t *testing.T) {
	suite.Run(t, new(JSONSuite))
}
