# ShieldCI Warden base

This directory is a source snapshot of [Prof-bilal/Warden](https://github.com/Prof-bilal/Warden) used as the first execution base for ShieldCI.

- Source commit: `65680c37cb69019a5ee46d8eb7d8bb2102d3d994`
- Go module: `github.com/warden-sandbox/warden`
- Go version: `1.24`
- Module root: `warden-starter/warden/`

ShieldCI product services must integrate through an adapter and preserve Warden’s fail-closed and deny-by-default guarantees. Treat this snapshot as upstream code: keep product changes outside it unless a specific upstream integration fix is required.
