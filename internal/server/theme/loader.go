package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadThemeFromFile loads a theme from a JSON or YAML file
func LoadThemeFromFile(path string) (*Theme, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file: %w", err)
	}

	// Determine file type by extension
	ext := strings.ToLower(filepath.Ext(path))

	var theme Theme

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &theme); err != nil {
			return nil, fmt.Errorf("failed to parse JSON theme file: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &theme); err != nil {
			return nil, fmt.Errorf("failed to parse YAML theme file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file extension %q (supported: .json, .yaml, .yml)", ext)
	}

	// Validate and fill in missing values with defaults
	if err := validateAndFillDefaults(&theme); err != nil {
		return nil, fmt.Errorf("theme validation failed: %w", err)
	}

	return &theme, nil
}

// validateAndFillDefaults validates a theme and fills in missing values with defaults
func validateAndFillDefaults(t *Theme) error {
	if t.Name == "" {
		return fmt.Errorf("theme name is required")
	}

	defaultTheme := GetDefaultTheme()

	// Fill in missing colors
	if t.Colors.Light.Background == "" {
		t.Colors.Light.Background = defaultTheme.Colors.Light.Background
	}
	if t.Colors.Light.Surface == "" {
		t.Colors.Light.Surface = defaultTheme.Colors.Light.Surface
	}
	if t.Colors.Light.Primary == "" {
		t.Colors.Light.Primary = defaultTheme.Colors.Light.Primary
	}
	if t.Colors.Light.Secondary == "" {
		t.Colors.Light.Secondary = defaultTheme.Colors.Light.Secondary
	}
	if t.Colors.Light.Text == "" {
		t.Colors.Light.Text = defaultTheme.Colors.Light.Text
	}
	if t.Colors.Light.TextSecondary == "" {
		t.Colors.Light.TextSecondary = defaultTheme.Colors.Light.TextSecondary
	}
	if t.Colors.Light.Border == "" {
		t.Colors.Light.Border = defaultTheme.Colors.Light.Border
	}
	if t.Colors.Light.Accent == "" {
		t.Colors.Light.Accent = defaultTheme.Colors.Light.Accent
	}
	if t.Colors.Light.AccentHover == "" {
		t.Colors.Light.AccentHover = defaultTheme.Colors.Light.AccentHover
	}
	if t.Colors.Light.Shadow == "" {
		t.Colors.Light.Shadow = defaultTheme.Colors.Light.Shadow
	}

	// Fill in missing dark colors
	if t.Colors.Dark.Background == "" {
		t.Colors.Dark.Background = defaultTheme.Colors.Dark.Background
	}
	if t.Colors.Dark.Surface == "" {
		t.Colors.Dark.Surface = defaultTheme.Colors.Dark.Surface
	}
	if t.Colors.Dark.Primary == "" {
		t.Colors.Dark.Primary = defaultTheme.Colors.Dark.Primary
	}
	if t.Colors.Dark.Secondary == "" {
		t.Colors.Dark.Secondary = defaultTheme.Colors.Dark.Secondary
	}
	if t.Colors.Dark.Text == "" {
		t.Colors.Dark.Text = defaultTheme.Colors.Dark.Text
	}
	if t.Colors.Dark.TextSecondary == "" {
		t.Colors.Dark.TextSecondary = defaultTheme.Colors.Dark.TextSecondary
	}
	if t.Colors.Dark.Border == "" {
		t.Colors.Dark.Border = defaultTheme.Colors.Dark.Border
	}
	if t.Colors.Dark.Accent == "" {
		t.Colors.Dark.Accent = defaultTheme.Colors.Dark.Accent
	}
	if t.Colors.Dark.AccentHover == "" {
		t.Colors.Dark.AccentHover = defaultTheme.Colors.Dark.AccentHover
	}
	if t.Colors.Dark.Shadow == "" {
		t.Colors.Dark.Shadow = defaultTheme.Colors.Dark.Shadow
	}

	// Fill in missing typography
	if t.Typography.FontFamily == "" {
		t.Typography.FontFamily = defaultTheme.Typography.FontFamily
	}
	if t.Typography.FontFamilyMono == "" {
		t.Typography.FontFamilyMono = defaultTheme.Typography.FontFamilyMono
	}
	if t.Typography.FontSizeBase == "" {
		t.Typography.FontSizeBase = defaultTheme.Typography.FontSizeBase
	}
	if t.Typography.LineHeight == "" {
		t.Typography.LineHeight = defaultTheme.Typography.LineHeight
	}

	// Fill in missing spacing
	if t.Spacing.HeaderHeight == "" {
		t.Spacing.HeaderHeight = defaultTheme.Spacing.HeaderHeight
	}
	if t.Spacing.ContentPadding == "" {
		t.Spacing.ContentPadding = defaultTheme.Spacing.ContentPadding
	}
	if t.Spacing.CardPadding == "" {
		t.Spacing.CardPadding = defaultTheme.Spacing.CardPadding
	}

	// Fill in missing component styles
	if t.Components.HeaderShadow == "" {
		t.Components.HeaderShadow = defaultTheme.Components.HeaderShadow
	}
	if t.Components.CardShadow == "" {
		t.Components.CardShadow = defaultTheme.Components.CardShadow
	}
	if t.Components.CardRadius == "" {
		t.Components.CardRadius = defaultTheme.Components.CardRadius
	}
	if t.Components.BorderWidth == "" {
		t.Components.BorderWidth = defaultTheme.Components.BorderWidth
	}

	return nil
}
