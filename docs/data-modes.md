# Data modes: live and runtime files

## data=live cells

A cell with `data=live` is compiled into the report binary but not
executed at build time. It runs at render time against the live data
source, using embedded secrets.

## reads(name:runtime) files

Files declared with `:runtime` mode are resolved relative to the
binary's working directory at render time, not embedded in the binary.

## Implementation

The kernel and build system check the `data` attribute from the
cell's info string. In build mode:

- `data=embedded` cells run at build time, results encrypted and embedded
- `data=live` cells are compiled only, run at render time
- `reads(name:embed)` files are embedded at build time
- `reads(name:runtime)` files are resolved at render time
