# Image handling

## Accepted formats

PNG, JPEG, GIF are embedded directly. WebP, BMP, TIFF are converted
to PNG. SVG is converted to PNG immediately (never served raw).

## Web images

Images referenced by URL are downloaded once into an `assets/`
directory next to the notebook. The download has a 30-second timeout.
Subsequent builds use the cached copy.

## Security

- Paths must stay inside the notebook directory (no `../` or absolute)
- SVG is rasterized to PNG to prevent scripts and external references
- HTML `<img>` tags are treated like Markdown image references
- All images are embedded as data URIs in HTML output
