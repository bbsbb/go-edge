package datatype

import (
	"database/sql/driver"
	"slices"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testEnum string

const (
	testEnumFoo testEnum = "foo"
	testEnumBar testEnum = "bar"
)

func (e testEnum) IsValid() bool {
	return slices.Contains([]testEnum{testEnumFoo, testEnumBar}, e)
}

func (e *testEnum) Scan(value any) error {
	return ScanStringEnum(e, value)
}

func (e testEnum) Value() (driver.Value, error) {
	return ValueStringEnum(e)
}

type EnumSuite struct {
	suite.Suite
}

func (s *EnumSuite) TestFromString() {
	tests := []struct {
		name    string
		input   string
		want    testEnum
		wantErr bool
	}{
		{
			name:    "valid enum foo",
			input:   "foo",
			want:    testEnumFoo,
			wantErr: false,
		},
		{
			name:    "valid enum bar",
			input:   "bar",
			want:    testEnumBar,
			wantErr: false,
		},
		{
			name:    "invalid enum",
			input:   "invalid",
			want:    testEnum(""),
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    testEnum(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := FromString[testEnum](tt.input)
			if tt.wantErr {
				s.Require().Error(err)
				s.Assert().Contains(err.Error(), "invalid value")
			} else {
				s.Require().NoError(err)
			}
			s.Assert().Equal(tt.want, got)
		})
	}
}

func (s *EnumSuite) TestValueStringEnum() {
	tests := []struct {
		name    string
		input   testEnum
		want    driver.Value
		wantErr bool
	}{
		{
			name:    "valid enum",
			input:   testEnumFoo,
			want:    "foo",
			wantErr: false,
		},
		{
			name:    "invalid enum",
			input:   testEnum("invalid"),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := ValueStringEnum(tt.input)
			if tt.wantErr {
				s.Require().Error(err)
				s.Assert().Contains(err.Error(), "is invalid")
			} else {
				s.Require().NoError(err)
			}
			s.Assert().Equal(tt.want, got)
		})
	}
}

func (s *EnumSuite) TestScanStringEnum() {
	tests := []struct {
		name    string
		input   any
		want    testEnum
		wantErr bool
	}{
		{
			name:    "valid string",
			input:   "foo",
			want:    testEnumFoo,
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "invalid",
			want:    testEnum(""),
			wantErr: true,
		},
		{
			name:    "wrong type int",
			input:   123,
			want:    testEnum(""),
			wantErr: true,
		},
		{
			name:    "wrong type bytes",
			input:   []byte("foo"),
			want:    testEnum(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var got testEnum
			err := ScanStringEnum(&got, tt.input)
			if tt.wantErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
			s.Assert().Equal(tt.want, got)
		})
	}
}

func (s *EnumSuite) TestNullStringEnum_Value() {
	tests := []struct {
		name    string
		input   NullStringEnum[testEnum]
		want    driver.Value
		wantErr bool
	}{
		{
			name:    "valid non-null",
			input:   NullStringEnum[testEnum]{Enum: testEnumFoo, Valid: true},
			want:    "foo",
			wantErr: false,
		},
		{
			name:    "null value",
			input:   NullStringEnum[testEnum]{Enum: testEnumFoo, Valid: false},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "invalid enum when valid",
			input:   NullStringEnum[testEnum]{Enum: testEnum("invalid"), Valid: true},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := tt.input.Value()
			if tt.wantErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
			s.Assert().Equal(tt.want, got)
		})
	}
}

func (s *EnumSuite) TestNullStringEnum_Scan() {
	tests := []struct {
		name      string
		input     any
		wantEnum  testEnum
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "valid string",
			input:     "foo",
			wantEnum:  testEnumFoo,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "null value",
			input:     nil,
			wantEnum:  testEnum(""),
			wantValid: false,
			wantErr:   false,
		},
		{
			name:      "invalid string",
			input:     "invalid",
			wantEnum:  testEnum(""),
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var got NullStringEnum[testEnum]
			err := got.Scan(tt.input)
			if tt.wantErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
			s.Assert().Equal(tt.wantEnum, got.Enum)
			s.Assert().Equal(tt.wantValid, got.Valid)
		})
	}
}

func TestEnumSuite(t *testing.T) {
	suite.Run(t, new(EnumSuite))
}
