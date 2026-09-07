package docs

import (
	"strings"

	"github.com/DonaldMurillo/gofastr/core/render"
)

// BrandConfig describes the small set of identity hooks a white-label site
// needs. Empty fields intentionally fall back to the framework-neutral mark
// and the Router's site name.
type BrandConfig struct {
	// Name is the site name in the header, footer, and PWA.
	Name string
	// LogoURL is the header logo image; empty falls back to the text
	// mark.
	LogoURL string
	// LogoAlt is the logo's alt text.
	LogoAlt string
	// FaviconURL is the bookmark and tab icon.
	FaviconURL string
	// ThemeColor tints the browser and OS chrome around the page.
	ThemeColor string
	// AccentColor overrides the theme's accent for brand-specific
	// highlights.
	AccentColor string
}

// WithBrand applies white-label identity without coupling the docs package to
// a particular product or logo asset.
func WithBrand(config BrandConfig) Option {
	return func(r *Router) {
		if strings.TrimSpace(config.Name) != "" {
			r.siteName = strings.TrimSpace(config.Name)
		}
		r.brand = mergeBrand(r.brand, config)
	}
}

func mergeBrand(base, override BrandConfig) BrandConfig {
	merged := base
	if strings.TrimSpace(override.Name) != "" {
		merged.Name = strings.TrimSpace(override.Name)
	}
	if strings.TrimSpace(override.LogoURL) != "" {
		merged.LogoURL = strings.TrimSpace(override.LogoURL)
	}
	if strings.TrimSpace(override.LogoAlt) != "" {
		merged.LogoAlt = strings.TrimSpace(override.LogoAlt)
	}
	if strings.TrimSpace(override.FaviconURL) != "" {
		merged.FaviconURL = strings.TrimSpace(override.FaviconURL)
	}
	if strings.TrimSpace(override.ThemeColor) != "" {
		merged.ThemeColor = strings.TrimSpace(override.ThemeColor)
	}
	if strings.TrimSpace(override.AccentColor) != "" {
		merged.AccentColor = strings.TrimSpace(override.AccentColor)
	}
	return merged
}

// Brand returns a copy of the configured white-label identity.
func (r *Router) Brand() BrandConfig { return r.brand }

// BrandCSS emits the optional token overrides for a host stylesheet.
func (r *Router) BrandCSS() string {
	if r == nil || r.brand.AccentColor == "" {
		return ""
	}
	return ":root { --color-accent: " + cssToken(r.brand.AccentColor) + "; --color-primary: " + cssToken(r.brand.AccentColor) + "; }"
}

func cssToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, char := range value {
		if !(char == '#' || char == '-' || char == '_' || char == '.' || char == ',' || char == '(' || char == ')' || char == '%' || char == ' ' || (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return ""
		}
	}
	return value
}

func (h *docsHeader) brandMark() render.HTML {
	brand := h.router.Brand()
	if logo := safeMetadataURL(brand.LogoURL); logo != "" {
		alt := brand.LogoAlt
		if alt == "" {
			alt = h.router.SiteName()
		}
		return render.Tag("img", map[string]string{
			"class": "fastr-docs-brand__logo", "src": logo, "alt": alt,
		})
	}
	return render.Raw(`<span class="fastr-docs-brand__mark" aria-hidden="true"><i></i><i></i><i></i></span>`)
}
