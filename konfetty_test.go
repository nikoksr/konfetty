package konfetty_test

import (
	"errors"
	"testing"

	"github.com/shoenig/test/must"

	"github.com/nikoksr/konfetty"
)

type TestConfig struct {
	Name    string
	Age     int
	IsAdmin bool
}

func TestFromStruct(t *testing.T) {
	t.Parallel()

	config := &TestConfig{Name: "Alice", Age: 30, IsAdmin: true}
	processor := konfetty.FromStruct(config)
	result, err := processor.Build()

	must.NoError(t, err)
	must.Eq(t, config, result)
}

func TestFromLoaderFunc(t *testing.T) {
	t.Parallel()

	loader := func() (TestConfig, error) {
		return TestConfig{Name: "Bob", Age: 25, IsAdmin: false}, nil
	}
	processor := konfetty.FromLoaderFunc(loader)

	result, err := processor.Build()
	must.NoError(t, err)
	must.Eq(t, &TestConfig{Name: "Bob", Age: 25, IsAdmin: false}, result)
}

type MockProvider struct {
	config TestConfig
	err    error
}

func (m *MockProvider) Load() (TestConfig, error) {
	return m.config, m.err
}

func TestFromProvider(t *testing.T) {
	t.Parallel()

	provider := &MockProvider{config: TestConfig{Name: "Charlie", Age: 35, IsAdmin: true}}
	processor := konfetty.FromProvider(provider)

	result, err := processor.Build()
	must.NoError(t, err)
	must.Eq(t, &TestConfig{Name: "Charlie", Age: 35, IsAdmin: true}, result)
}

func TestWithDefaults(t *testing.T) {
	t.Parallel()

	config := &TestConfig{}
	processor := konfetty.FromStruct(config).WithDefaults(TestConfig{Name: "Default", Age: 18, IsAdmin: false})

	result, err := processor.Build()
	must.NoError(t, err)
	must.Eq(t, &TestConfig{Name: "Default", Age: 18, IsAdmin: false}, result)
}

func TestWithTransformer(t *testing.T) {
	t.Parallel()

	config := &TestConfig{Name: "Dave", Age: 20}
	transformer := func(c *TestConfig) {
		c.Name = "Mr. " + c.Name
		c.Age++
	}
	processor := konfetty.FromStruct(config).WithTransformer(transformer)

	result, err := processor.Build()
	must.NoError(t, err)
	must.Eq(t, &TestConfig{Name: "Mr. Dave", Age: 21}, result)
}

func TestWithValidator(t *testing.T) {
	t.Parallel()

	config := &TestConfig{Name: "Eve", Age: 17}
	validator := func(c *TestConfig) error {
		if c.Age < 18 {
			return errors.New("age must be 18 or older")
		}
		return nil
	}
	processor := konfetty.FromStruct(config).WithValidator(validator)

	_, err := processor.Build()
	must.Error(t, err)
	must.ErrorContains(t, err, "age must be 18 or older")
}

func TestBuildWithAllOptions(t *testing.T) {
	t.Parallel()

	config := &TestConfig{}
	processor := konfetty.FromStruct(config).
		WithDefaults(TestConfig{Name: "Default", Age: 18, IsAdmin: false}).
		WithTransformer(func(c *TestConfig) {
			c.Name = "Mr. " + c.Name
		}).
		WithValidator(func(c *TestConfig) error {
			if c.Age < 18 {
				return errors.New("age must be 18 or older")
			}
			return nil
		})

	result, err := processor.Build()
	must.NoError(t, err)
	must.Eq(t, &TestConfig{Name: "Mr. Default", Age: 18, IsAdmin: false}, result)
}

func TestBuildErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("LoaderFuncError", func(t *testing.T) {
		t.Parallel()

		loader := func() (TestConfig, error) {
			return TestConfig{}, errors.New("loader error")
		}
		processor := konfetty.FromLoaderFunc(loader)

		_, err := processor.Build()
		must.Error(t, err)
		must.ErrorContains(t, err, "loader error")
	})

	t.Run("ProviderError", func(t *testing.T) {
		t.Parallel()

		provider := &MockProvider{err: errors.New("provider error")}
		processor := konfetty.FromProvider(provider)

		_, err := processor.Build()
		must.Error(t, err)
		must.ErrorContains(t, err, "provider error")
	})

	t.Run("ValidatorError", func(t *testing.T) {
		t.Parallel()

		config := &TestConfig{}
		validator := func(_ *TestConfig) error {
			return errors.New("validator error")
		}
		processor := konfetty.FromStruct(config).WithValidator(validator)

		_, err := processor.Build()
		must.Error(t, err)
		must.ErrorContains(t, err, "validator error")
	})
}
