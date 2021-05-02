# Block Exports

`core-go` uses tar based block exports to transport (large) parts of the blockchain offline.

The following constraints apply:
- The tar format is `ustar`.
- The first file is the `manifest.json`, padded to 1024 bytes with spaces.
- TODO write rest of the docs

In general, most tar archivers can read block exports just fine,
but will corrupt them when editing.
