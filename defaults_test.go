package konfetty

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

// Structs provided by the user
type Profile struct {
	Checks Checks
}

type Checks struct {
	Ping []PingCheck
}

type PingCheck struct {
	*BaseCheck
	Host string
}

type BaseCheck struct {
	Name     string
	Interval time.Duration
	Timeout  time.Duration
}

// Implement DefaultProvider for BaseCheck
func (b BaseCheck) Defaults() interface{} {
	return BaseCheck{
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
	}
}

func TestFillDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    *Profile
		expected *Profile
	}{
		{
			name: "Profile with partially filled checks",
			input: &Profile{
				Checks: Checks{
					Ping: []PingCheck{
						{
							BaseCheck: &BaseCheck{Name: "Custom Ping"},
							Host:      "example.com",
						},
					},
				},
			},
			expected: &Profile{
				Checks: Checks{
					Ping: []PingCheck{
						{
							BaseCheck: &BaseCheck{
								Name:     "Custom Ping",
								Interval: 30 * time.Second,
								Timeout:  5 * time.Second,
							},
							Host: "example.com",
						},
					},
				},
			},
		},
		{
			name: "Profile with nil checks",
			input: &Profile{
				Checks: Checks{
					Ping: []PingCheck{
						{
							Host: "example.com",
						},
					},
				},
			},
			expected: &Profile{
				Checks: Checks{
					Ping: []PingCheck{
						{
							BaseCheck: &BaseCheck{
								Interval: 30 * time.Second,
								Timeout:  5 * time.Second,
							},
							Host: "example.com",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fillDefaults(tt.input)
			require.NoError(t, err)

			diff := cmp.Diff(tt.expected, tt.input)
			if diff != "" {
				t.Errorf("FillDefaults() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
