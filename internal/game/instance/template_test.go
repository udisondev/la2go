package instance

import (
	"errors"
	"testing"
	"time"
)

func TestTemplate_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    Template
		wantErr error
	}{
		{
			name:    "valid_minimal",
			tmpl:    Template{ID: 1, Name: "Test"},
			wantErr: nil,
		},
		{
			name: "valid_full",
			tmpl: Template{
				ID:          10,
				Name:        "Kamaloka",
				Duration:    30 * time.Minute,
				MaxPlayers:  9,
				MinLevel:    40,
				MaxLevel:    52,
				Cooldown:    24 * time.Hour,
				RemoveBuffs: true,
				SpawnX:      -56000,
				SpawnY:      -56000,
				SpawnZ:      -2000,
			},
			wantErr: nil,
		},
		{
			name:    "zero_id",
			tmpl:    Template{ID: 0, Name: "Test"},
			wantErr: ErrInvalidTemplateID,
		},
		{
			name:    "negative_id",
			tmpl:    Template{ID: -1, Name: "Test"},
			wantErr: ErrInvalidTemplateID,
		},
		{
			name:    "empty_name",
			tmpl:    Template{ID: 1, Name: ""},
			wantErr: ErrEmptyTemplateName,
		},
		{
			name:    "negative_max_players",
			tmpl:    Template{ID: 1, Name: "Test", MaxPlayers: -1},
			wantErr: ErrInvalidMaxPlayers,
		},
		{
			name:    "min_level_too_high",
			tmpl:    Template{ID: 1, Name: "Test", MinLevel: 81},
			wantErr: ErrInvalidLevel,
		},
		{
			name:    "max_level_too_high",
			tmpl:    Template{ID: 1, Name: "Test", MaxLevel: 81},
			wantErr: ErrInvalidLevel,
		},
		{
			name:    "min_greater_than_max",
			tmpl:    Template{ID: 1, Name: "Test", MinLevel: 60, MaxLevel: 40},
			wantErr: ErrInvalidLevel,
		},
		{
			name:    "negative_min_level",
			tmpl:    Template{ID: 1, Name: "Test", MinLevel: -1},
			wantErr: ErrInvalidLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tmpl.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Validate() error = %v; want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v; want %v", err, tt.wantErr)
			}
		})
	}
}
