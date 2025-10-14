package theme

// Theme represents a visual theme configuration for the UI
type Theme struct {
	Name        string
	Colors      ColorScheme
	Typography  Typography
	Spacing     Spacing
	Components  Components
	CustomCSS   string // Additional CSS to inject
}

// ColorScheme defines color palettes for light and dark modes
type ColorScheme struct {
	Light LightColors
	Dark  DarkColors
}

// LightColors defines the color palette for light mode
type LightColors struct {
	Background    string
	Surface       string
	Primary       string
	Secondary     string
	Text          string
	TextSecondary string
	Border        string
	Accent        string
	AccentHover   string
	Shadow        string
}

// DarkColors defines the color palette for dark mode
type DarkColors struct {
	Background    string
	Surface       string
	Primary       string
	Secondary     string
	Text          string
	TextSecondary string
	Border        string
	Accent        string
	AccentHover   string
	Shadow        string
}

// Typography defines font and text styling
type Typography struct {
	FontFamily     string
	FontFamilyMono string
	FontSizeBase   string
	LineHeight     string
}

// Spacing defines layout spacing values
type Spacing struct {
	HeaderHeight string
	ContentPadding string
	CardPadding    string
}

// Components defines component-specific styles
type Components struct {
	HeaderShadow  string
	CardShadow    string
	CardRadius    string
	BorderWidth   string
}

// Config holds the current theme configuration
type Config struct {
	CurrentTheme *Theme
	AllowCustom  bool
}

// GetDefaultTheme returns the default Reflect theme
func GetDefaultTheme() *Theme {
	return &Theme{
		Name: "default",
		Colors: ColorScheme{
			Light: LightColors{
				Background:    "#f9fafb",
				Surface:       "#ffffff",
				Primary:       "#111827",
				Secondary:     "#6b7280",
				Text:          "#111827",
				TextSecondary: "#6b7280",
				Border:        "#e5e7eb",
				Accent:        "#2563eb",
				AccentHover:   "#1d4ed8",
				Shadow:        "rgba(0, 0, 0, 0.1)",
			},
			Dark: DarkColors{
				Background:    "#0f172a",
				Surface:       "#1e293b",
				Primary:       "#f1f5f9",
				Secondary:     "#94a3b8",
				Text:          "#f1f5f9",
				TextSecondary: "#94a3b8",
				Border:        "#334155",
				Accent:        "#3b82f6",
				AccentHover:   "#60a5fa",
				Shadow:        "rgba(0, 0, 0, 0.5)",
			},
		},
		Typography: Typography{
			FontFamily:     "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
			FontFamilyMono: "'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace",
			FontSizeBase:   "16px",
			LineHeight:     "1.6",
		},
		Spacing: Spacing{
			HeaderHeight:   "4rem",
			ContentPadding: "2rem",
			CardPadding:    "1.5rem",
		},
		Components: Components{
			HeaderShadow: "0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)",
			CardShadow:   "0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)",
			CardRadius:   "0.5rem",
			BorderWidth:  "1px",
		},
	}
}

// GetMinimalTheme returns a minimal theme with less visual noise
func GetMinimalTheme() *Theme {
	theme := GetDefaultTheme()
	theme.Name = "minimal"

	// Lighter borders and shadows
	theme.Colors.Light.Border = "#f3f4f6"
	theme.Colors.Dark.Border = "#1f2937"
	theme.Components.HeaderShadow = "0 1px 2px 0 rgba(0, 0, 0, 0.05)"
	theme.Components.CardShadow = "none"
	theme.Components.BorderWidth = "1px"

	return theme
}

// GetHighContrastTheme returns a high contrast theme for better accessibility
func GetHighContrastTheme() *Theme {
	theme := GetDefaultTheme()
	theme.Name = "high-contrast"

	// Stronger contrasts
	theme.Colors.Light.Background = "#ffffff"
	theme.Colors.Light.Surface = "#f9fafb"
	theme.Colors.Light.Primary = "#000000"
	theme.Colors.Light.Border = "#d1d5db"
	theme.Colors.Light.Accent = "#1e40af"

	theme.Colors.Dark.Background = "#000000"
	theme.Colors.Dark.Surface = "#1a1a1a"
	theme.Colors.Dark.Primary = "#ffffff"
	theme.Colors.Dark.Border = "#404040"
	theme.Colors.Dark.Accent = "#60a5fa"

	theme.Components.BorderWidth = "2px"

	return theme
}

// GetThemeByName returns a theme by its name
func GetThemeByName(name string) *Theme {
	switch name {
	case "minimal":
		return GetMinimalTheme()
	case "high-contrast":
		return GetHighContrastTheme()
	default:
		return GetDefaultTheme()
	}
}

// ToCSSVariables converts a theme to CSS custom properties
func (t *Theme) ToCSSVariables() map[string]string {
	vars := make(map[string]string)

	// Light mode colors
	vars["--color-bg-light"] = t.Colors.Light.Background
	vars["--color-surface-light"] = t.Colors.Light.Surface
	vars["--color-primary-light"] = t.Colors.Light.Primary
	vars["--color-secondary-light"] = t.Colors.Light.Secondary
	vars["--color-text-light"] = t.Colors.Light.Text
	vars["--color-text-secondary-light"] = t.Colors.Light.TextSecondary
	vars["--color-border-light"] = t.Colors.Light.Border
	vars["--color-accent-light"] = t.Colors.Light.Accent
	vars["--color-accent-hover-light"] = t.Colors.Light.AccentHover
	vars["--color-shadow-light"] = t.Colors.Light.Shadow

	// Dark mode colors
	vars["--color-bg-dark"] = t.Colors.Dark.Background
	vars["--color-surface-dark"] = t.Colors.Dark.Surface
	vars["--color-primary-dark"] = t.Colors.Dark.Primary
	vars["--color-secondary-dark"] = t.Colors.Dark.Secondary
	vars["--color-text-dark"] = t.Colors.Dark.Text
	vars["--color-text-secondary-dark"] = t.Colors.Dark.TextSecondary
	vars["--color-border-dark"] = t.Colors.Dark.Border
	vars["--color-accent-dark"] = t.Colors.Dark.Accent
	vars["--color-accent-hover-dark"] = t.Colors.Dark.AccentHover
	vars["--color-shadow-dark"] = t.Colors.Dark.Shadow

	// Typography
	vars["--font-family"] = t.Typography.FontFamily
	vars["--font-family-mono"] = t.Typography.FontFamilyMono
	vars["--font-size-base"] = t.Typography.FontSizeBase
	vars["--line-height"] = t.Typography.LineHeight

	// Spacing
	vars["--header-height"] = t.Spacing.HeaderHeight
	vars["--content-padding"] = t.Spacing.ContentPadding
	vars["--card-padding"] = t.Spacing.CardPadding

	// Components
	vars["--header-shadow"] = t.Components.HeaderShadow
	vars["--card-shadow"] = t.Components.CardShadow
	vars["--card-radius"] = t.Components.CardRadius
	vars["--border-width"] = t.Components.BorderWidth

	return vars
}
