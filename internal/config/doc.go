// Package config handles loading, saving, and defaulting of the global
// repomni configuration stored at $XDG_CONFIG_HOME/repomni/config.yaml.
//
// The configuration defines a source directory, an injection mode (symlink or
// copy), and a list of items (files or directories) to inject into target
// repositories. Paths in the config support tilde and environment variable
// expansion via [ExpandPath].
package config
