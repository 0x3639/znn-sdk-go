// Package modelwire provides executable JSON decode/encode plumbing for the
// stable Zenon model conformance fixtures.
//
// The package instantiates the SDK's client response models and the canonical
// go-zenon dependency types used by the public API. Constructor-only primitive
// shapes and abstract collection models use small typed views because their
// fixture representation is not the JSON-RPC representation of the underlying
// Go primitive.
package modelwire
