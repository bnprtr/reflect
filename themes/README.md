# Custom Themes

This directory contains example custom theme files that can be loaded with the `--theme-file` flag.

## Usage

```bash
# Load a custom JSON theme
./reflect --theme-file themes/example-custom.json --proto-root /path/to/protos

# Load a custom YAML theme
./reflect --theme-file themes/example-minimal.yaml --proto-root /path/to/protos
```

## Theme File Format

Themes can be defined in either JSON or YAML format. Both formats support the same structure:

### Required Fields

- `name`: The name of your theme (string)
- `colors`: Color palettes for light and dark modes
  - `light`: Light mode colors
  - `dark`: Dark mode colors

### Color Properties

Each color mode (light/dark) requires these properties:

- `background`: Main background color
- `surface`: Surface/card background color
- `primary`: Primary text color
- `secondary`: Secondary text color
- `text`: Default text color
- `textSecondary`: Muted text color
- `border`: Border color
- `accent`: Accent/link color
- `accentHover`: Accent hover state color
- `shadow`: Shadow color (rgba format)

### Optional Fields

If not specified, these will use default values:

- `typography`: Font and typography settings
  - `fontFamily`: Main font stack
  - `fontFamilyMono`: Monospace font stack
  - `fontSizeBase`: Base font size
  - `lineHeight`: Line height

- `spacing`: Layout spacing
  - `headerHeight`: Header height
  - `contentPadding`: Content padding
  - `cardPadding`: Card padding

- `components`: Component-specific styles
  - `headerShadow`: Header shadow
  - `cardShadow`: Card shadow
  - `cardRadius`: Card border radius
  - `borderWidth`: Border width

- `customCSS`: Additional CSS to inject (string)

## Examples

See `example-custom.json` for a purple theme in JSON format, and `example-minimal.yaml` for a pink theme in YAML format.

## Creating Your Own Theme

1. Copy one of the example files
2. Modify the colors to match your desired palette
3. Optionally customize typography, spacing, and components
4. Save with a `.json` or `.yaml`/`.yml` extension
5. Load with `--theme-file path/to/your-theme.json`

## Tips

- Use online color palette generators to create cohesive color schemes
- Test your theme in both light and dark modes
- Ensure sufficient contrast for accessibility (WCAG AA: 4.5:1 for text)
- Use tools like https://coolors.co or https://paletton.com for color inspiration
