package konfetty

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	Name     string        `konfetty:"name"`
	Age      int           `konfetty:"age"`
	Timeout  time.Duration `konfetty:"timeout"`
	Features []string      `konfetty:"features"`
}

func (c TestConfig) Defaults() any {
	return TestConfig{
		Name:    "Default",
		Age:     30,
		Timeout: 5 * time.Second,
	}
}

func TestNewLoader(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		loader := NewLoader[TestConfig]()
		assert.NotNil(t, loader)
	})

	t.Run("Custom configuration", func(t *testing.T) {
		loader := NewLoader[TestConfig](
			WithKoanfDelimiter[TestConfig]("_"),
			WithEnvPrefix[TestConfig]("TEST_"),
			WithFileFormat[TestConfig](JSON),
			WithStructTag[TestConfig]("json"),
			WithMaxDepth[TestConfig](50),
		)
		assert.NotNil(t, loader)
	})

	t.Run("With custom koanf setup", func(t *testing.T) {
		customSetup := func(k *koanf.Koanf) error {
			k.Set("custom", "value")
			return nil
		}
		loader := NewLoader[TestConfig](WithKoanfSetup[TestConfig](customSetup))
		assert.NotNil(t, loader)
	})
}

func TestLoad(t *testing.T) {
	t.Run("Load from YAML file", func(t *testing.T) {
		yamlContent := `
name: John Doe
age: 35
timeout: 10s
features:
  - feature1
  - feature2
`
		tmpfile, err := os.CreateTemp("", "test*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.Write([]byte(yamlContent))
		require.NoError(t, err)
		tmpfile.Close()

		loader := NewLoader[TestConfig]()
		cfg, err := loader.Load(tmpfile.Name())
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "John Doe", cfg.Name)
		assert.Equal(t, 35, cfg.Age)
		assert.Equal(t, 10*time.Second, cfg.Timeout)
		assert.Equal(t, []string{"feature1", "feature2"}, cfg.Features)
	})

	t.Run("Load from JSON file", func(t *testing.T) {
		jsonContent := `{
"name": "Jane Doe",
"age": 28,
"timeout": "15s",
"features": ["feature3", "feature4"]
}`
		tmpfile, err := os.CreateTemp("", "test*.json")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.Write([]byte(jsonContent))
		require.NoError(t, err)
		tmpfile.Close()

		loader := NewLoader[TestConfig](WithFileFormat[TestConfig](JSON))
		cfg, err := loader.Load(tmpfile.Name())
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "Jane Doe", cfg.Name)
		assert.Equal(t, 28, cfg.Age)
		assert.Equal(t, 15*time.Second, cfg.Timeout)
		assert.Equal(t, []string{"feature3", "feature4"}, cfg.Features)
	})

	t.Run("Load from TOML file", func(t *testing.T) {
		tomlContent := `
name = "Bob Smith"
age = 40
timeout = "20s"
features = ["feature5", "feature6"]
`
		tmpfile, err := os.CreateTemp("", "test*.toml")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.Write([]byte(tomlContent))
		require.NoError(t, err)
		tmpfile.Close()

		loader := NewLoader[TestConfig](WithFileFormat[TestConfig](TOML))
		cfg, err := loader.Load(tmpfile.Name())
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "Bob Smith", cfg.Name)
		assert.Equal(t, 40, cfg.Age)
		assert.Equal(t, 20*time.Second, cfg.Timeout)
		assert.Equal(t, []string{"feature5", "feature6"}, cfg.Features)
	})

	t.Run("Load multiple files", func(t *testing.T) {
		yamlContent1 := `
name: John Doe
age: 35
`
		yamlContent2 := `
timeout: 10s
features:
  - feature1
  - feature2
`
		tmpfile1, err := os.CreateTemp("", "test1*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpfile1.Name())

		tmpfile2, err := os.CreateTemp("", "test2*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpfile2.Name())

		_, err = tmpfile1.Write([]byte(yamlContent1))
		require.NoError(t, err)
		tmpfile1.Close()

		_, err = tmpfile2.Write([]byte(yamlContent2))
		require.NoError(t, err)
		tmpfile2.Close()

		loader := NewLoader[TestConfig]()
		cfg, err := loader.Load(tmpfile1.Name(), tmpfile2.Name())
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "John Doe", cfg.Name)
		assert.Equal(t, 35, cfg.Age)
		assert.Equal(t, 10*time.Second, cfg.Timeout)
		assert.Equal(t, []string{"feature1", "feature2"}, cfg.Features)
	})

	t.Run("Load with environment variables", func(t *testing.T) {
		os.Setenv("TEST_NAME", "Env User")
		os.Setenv("TEST_AGE", "40")
		defer os.Unsetenv("TEST_NAME")
		defer os.Unsetenv("TEST_AGE")

		loader := NewLoader[TestConfig](WithEnvPrefix[TestConfig]("TEST_"))
		cfg, err := loader.Load()
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "Env User", cfg.Name)
		assert.Equal(t, 40, cfg.Age)
	})

	t.Run("Load with custom unmarshal function", func(t *testing.T) {
		customUnmarshal := func(data interface{}) (*TestConfig, error) {
			return &TestConfig{Name: "Custom", Age: 50}, nil
		}

		loader := NewLoader[TestConfig](WithUnmarshal[TestConfig](customUnmarshal))
		cfg, err := loader.Load()
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "Custom", cfg.Name)
		assert.Equal(t, 50, cfg.Age)
	})

	t.Run("Load with validation", func(t *testing.T) {
		validate := func(cfg *TestConfig) error {
			if cfg.Age < 0 {
				return fmt.Errorf("age must be non-negative")
			}
			return nil
		}

		loader := NewLoader[TestConfig](WithValidator[TestConfig](validate))

		yamlContent := `
name: Invalid
age: -5
`
		tmpfile, err := os.CreateTemp("", "test*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.Write([]byte(yamlContent))
		require.NoError(t, err)
		tmpfile.Close()

		_, err = loader.Load(tmpfile.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "age must be non-negative")
	})

	t.Run("Load with transformation", func(t *testing.T) {
		transform := func(cfg *TestConfig) error {
			cfg.Name = "Transformed: " + cfg.Name
			return nil
		}

		loader := NewLoader[TestConfig](WithTransformer[TestConfig](transform))

		yamlContent := `
name: Original
age: 25
`
		tmpfile, err := os.CreateTemp("", "test*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.Write([]byte(yamlContent))
		require.NoError(t, err)
		tmpfile.Close()

		cfg, err := loader.Load(tmpfile.Name())
		require.NoError(t, err)
		assert.Equal(t, "Transformed: Original", cfg.Name)
	})

	t.Run("Load with defaults", func(t *testing.T) {
		loader := NewLoader[TestConfig]()

		yamlContent := `
name: Custom
`
		tmpfile, err := os.CreateTemp("", "test*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.Write([]byte(yamlContent))
		require.NoError(t, err)
		tmpfile.Close()

		cfg, err := loader.Load(tmpfile.Name())
		require.NoError(t, err)
		assert.Equal(t, "Custom", cfg.Name)
		assert.Equal(t, 30, cfg.Age)                // Default value
		assert.Equal(t, 5*time.Second, cfg.Timeout) // Default value
	})

	t.Run("Load with custom koanf setup", func(t *testing.T) {
		customSetup := func(k *koanf.Koanf) error {
			k.Set("name", "CustomSetup")
			return nil
		}
		loader := NewLoader[TestConfig](WithKoanfSetup[TestConfig](customSetup))
		cfg, err := loader.Load()
		require.NoError(t, err)
		assert.Equal(t, "CustomSetup", cfg.Name)
	})

	t.Run("Load with invalid file format", func(t *testing.T) {
		loader := NewLoader[TestConfig](WithFileFormat[TestConfig](999)) // Invalid format
		_, err := loader.Load("dummy.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported file format")
	})

	t.Run("Load with non-existent file", func(t *testing.T) {
		loader := NewLoader[TestConfig]()
		_, err := loader.Load("non_existent_file.yaml")
		assert.Error(t, err)
	})

	t.Run("Load with invalid YAML", func(t *testing.T) {
		yamlContent := `
name: Invalid YAML
age: [this is not a valid integer]
`
		tmpfile, err := os.CreateTemp("", "test*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.Write([]byte(yamlContent))
		require.NoError(t, err)
		tmpfile.Close()

		loader := NewLoader[TestConfig]()
		_, err = loader.Load(tmpfile.Name())
		assert.Error(t, err)
	})
}

func TestValidate(t *testing.T) {
	validate := func(cfg *TestConfig) error {
		if cfg.Age < 0 {
			return fmt.Errorf("age must be non-negative")
		}
		return nil
	}

	loader := NewLoader[TestConfig](WithValidator[TestConfig](validate))

	t.Run("Valid configuration", func(t *testing.T) {
		cfg := &TestConfig{Name: "Valid", Age: 25}
		err := loader.Validate(cfg)
		assert.NoError(t, err)
	})

	t.Run("Invalid configuration", func(t *testing.T) {
		cfg := &TestConfig{Name: "Invalid", Age: -5}
		err := loader.Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "age must be non-negative")
	})
}

func TestTransform(t *testing.T) {
	transform := func(cfg *TestConfig) error {
		cfg.Name = "Transformed: " + cfg.Name
		cfg.Age *= 2
		return nil
	}

	loader := NewLoader[TestConfig](WithTransformer[TestConfig](transform))

	t.Run("Transform configuration", func(t *testing.T) {
		cfg := &TestConfig{Name: "Original", Age: 25}
		err := loader.Transform(cfg)
		assert.NoError(t, err)
		assert.Equal(t, "Transformed: Original", cfg.Name)
		assert.Equal(t, 50, cfg.Age)
	})
}
