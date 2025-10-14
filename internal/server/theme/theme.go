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

// GetOceanTheme returns a blue/teal ocean-inspired theme
func GetOceanTheme() *Theme {
	return &Theme{
		Name: "ocean",
		Colors: ColorScheme{
			Light: LightColors{
				Background:    "#f0f9ff",
				Surface:       "#ffffff",
				Primary:       "#0c4a6e",
				Secondary:     "#0e7490",
				Text:          "#0c4a6e",
				TextSecondary: "#0e7490",
				Border:        "#bae6fd",
				Accent:        "#0284c7",
				AccentHover:   "#0369a1",
				Shadow:        "rgba(2, 132, 199, 0.1)",
			},
			Dark: DarkColors{
				Background:    "#082f49",
				Surface:       "#0c4a6e",
				Primary:       "#e0f2fe",
				Secondary:     "#7dd3fc",
				Text:          "#e0f2fe",
				TextSecondary: "#7dd3fc",
				Border:        "#164e63",
				Accent:        "#0ea5e9",
				AccentHover:   "#38bdf8",
				Shadow:        "rgba(14, 165, 233, 0.3)",
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
			HeaderShadow: "0 1px 3px 0 rgba(2, 132, 199, 0.1), 0 1px 2px 0 rgba(2, 132, 199, 0.06)",
			CardShadow:   "0 1px 3px 0 rgba(2, 132, 199, 0.1), 0 1px 2px 0 rgba(2, 132, 199, 0.06)",
			CardRadius:   "0.5rem",
			BorderWidth:  "1px",
		},
	}
}

// GetForestTheme returns a green nature-inspired theme
func GetForestTheme() *Theme {
	return &Theme{
		Name: "forest",
		Colors: ColorScheme{
			Light: LightColors{
				Background:    "#f0fdf4",
				Surface:       "#ffffff",
				Primary:       "#14532d",
				Secondary:     "#166534",
				Text:          "#14532d",
				TextSecondary: "#166534",
				Border:        "#bbf7d0",
				Accent:        "#16a34a",
				AccentHover:   "#15803d",
				Shadow:        "rgba(22, 163, 74, 0.1)",
			},
			Dark: DarkColors{
				Background:    "#052e16",
				Surface:       "#14532d",
				Primary:       "#dcfce7",
				Secondary:     "#86efac",
				Text:          "#dcfce7",
				TextSecondary: "#86efac",
				Border:        "#166534",
				Accent:        "#22c55e",
				AccentHover:   "#4ade80",
				Shadow:        "rgba(34, 197, 94, 0.3)",
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
			HeaderShadow: "0 1px 3px 0 rgba(22, 163, 74, 0.1), 0 1px 2px 0 rgba(22, 163, 74, 0.06)",
			CardShadow:   "0 1px 3px 0 rgba(22, 163, 74, 0.1), 0 1px 2px 0 rgba(22, 163, 74, 0.06)",
			CardRadius:   "0.5rem",
			BorderWidth:  "1px",
		},
	}
}

// GetSunsetTheme returns an orange/purple warm sunset theme
func GetSunsetTheme() *Theme {
	return &Theme{
		Name: "sunset",
		Colors: ColorScheme{
			Light: LightColors{
				Background:    "#fff7ed",
				Surface:       "#ffffff",
				Primary:       "#7c2d12",
				Secondary:     "#c2410c",
				Text:          "#7c2d12",
				TextSecondary: "#c2410c",
				Border:        "#fed7aa",
				Accent:        "#ea580c",
				AccentHover:   "#c2410c",
				Shadow:        "rgba(234, 88, 12, 0.1)",
			},
			Dark: DarkColors{
				Background:    "#431407",
				Surface:       "#7c2d12",
				Primary:       "#ffedd5",
				Secondary:     "#fdba74",
				Text:          "#ffedd5",
				TextSecondary: "#fdba74",
				Border:        "#9a3412",
				Accent:        "#f97316",
				AccentHover:   "#fb923c",
				Shadow:        "rgba(249, 115, 22, 0.3)",
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
			HeaderShadow: "0 1px 3px 0 rgba(234, 88, 12, 0.1), 0 1px 2px 0 rgba(234, 88, 12, 0.06)",
			CardShadow:   "0 1px 3px 0 rgba(234, 88, 12, 0.1), 0 1px 2px 0 rgba(234, 88, 12, 0.06)",
			CardRadius:   "0.5rem",
			BorderWidth:  "1px",
		},
	}
}

// GetMonochromeTheme returns a grayscale-only minimalist theme
func GetMonochromeTheme() *Theme {
	return &Theme{
		Name: "monochrome",
		Colors: ColorScheme{
			Light: LightColors{
				Background:    "#ffffff",
				Surface:       "#fafafa",
				Primary:       "#171717",
				Secondary:     "#525252",
				Text:          "#171717",
				TextSecondary: "#525252",
				Border:        "#e5e5e5",
				Accent:        "#404040",
				AccentHover:   "#262626",
				Shadow:        "rgba(0, 0, 0, 0.08)",
			},
			Dark: DarkColors{
				Background:    "#0a0a0a",
				Surface:       "#171717",
				Primary:       "#fafafa",
				Secondary:     "#a3a3a3",
				Text:          "#fafafa",
				TextSecondary: "#a3a3a3",
				Border:        "#262626",
				Accent:        "#d4d4d4",
				AccentHover:   "#e5e5e5",
				Shadow:        "rgba(255, 255, 255, 0.05)",
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
			HeaderShadow: "0 1px 3px 0 rgba(0, 0, 0, 0.08), 0 1px 2px 0 rgba(0, 0, 0, 0.04)",
			CardShadow:   "0 1px 3px 0 rgba(0, 0, 0, 0.08), 0 1px 2px 0 rgba(0, 0, 0, 0.04)",
			CardRadius:   "0.5rem",
			BorderWidth:  "1px",
		},
	}
}

// GetThemeByName returns a theme by its name
func GetThemeByName(name string) *Theme {
	switch name {
	case "minimal":
		return GetMinimalTheme()
	case "high-contrast":
		return GetHighContrastTheme()
	case "ocean":
		return GetOceanTheme()
	case "forest":
		return GetForestTheme()
	case "sunset":
		return GetSunsetTheme()
	case "monochrome":
		return GetMonochromeTheme()
	default:
		return GetDefaultTheme()
	}
}

// GetAllThemes returns a list of all available theme names
func GetAllThemes() []string {
	return []string{
		"default",
		"minimal",
		"high-contrast",
		"ocean",
		"forest",
		"sunset",
		"monochrome",
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
