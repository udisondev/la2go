package zone

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDamagePerSecond(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
		want   int32
	}{
		{
			name:   "normal value",
			params: map[string]string{"damagePerSecond": "200"},
			want:   200,
		},
		{
			name:   "missing param",
			params: map[string]string{},
			want:   0,
		},
		{
			name:   "invalid number",
			params: map[string]string{"damagePerSecond": "abc"},
			want:   0,
		},
		{
			name:   "nil params",
			params: nil,
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &DamageZone{
				BaseZone: &BaseZone{
					id:       100,
					name:     "test-dmg",
					zoneType: TypeDamage,
					params:   tt.params,
				},
			}
			assert.Equal(t, tt.want, z.DamagePerSecond())
		})
	}
}

func TestDamageInterval(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
		want   int32
	}{
		{
			name:   "custom interval",
			params: map[string]string{"damageInterval": "5"},
			want:   5,
		},
		{
			name:   "default when missing",
			params: map[string]string{},
			want:   defaultDamageInterval,
		},
		{
			name:   "default on invalid",
			params: map[string]string{"damageInterval": "xyz"},
			want:   defaultDamageInterval,
		},
		{
			name:   "nil params defaults",
			params: nil,
			want:   defaultDamageInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &DamageZone{
				BaseZone: &BaseZone{
					id:       101,
					name:     "test-dmg",
					zoneType: TypeDamage,
					params:   tt.params,
				},
			}
			assert.Equal(t, tt.want, z.DamageInterval())
		})
	}
}
