// Package config provides configuration structures for document processing components.
//
// This package defines configuration types that are used to initialize domain objects
// in other packages. Configuration structures support JSON serialization, default values,
// and merge semantics for layered configuration.
//
// Configuration objects should only exist during initialization and should be transformed
// into domain objects at system boundaries using transformation functions (e.g., NewRenderer).
package config
