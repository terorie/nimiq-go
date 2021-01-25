// Package beserial implements encoding and decoding of
// the Nimiq "big-endian-serialization" wire protocol.
// It is based on simple binary big-endian encodings of numbers.
//
// The default encoding follows just a handful of rules:
//  - Numbers are two's-complement & big-endian encoded (un)signed 8/16/32/64-bit.
//  - Structures are just their elements appended to each other.
//  - Fixed-size arrays are individual beserial-encodings of their elements.
//  - Variable-length arrays/strings are same as above, but length-prefixed.
//  - Booleans are 8-bit integers of 0x00 for false and 0x01 for true.
//  - Optional items are prefixed by a boolean, and omitted if the bool is false.
// By implementing the Marshaler and/or Unmarshaler interfaces
// the default encoding is overwritten with the custom code.
// Examples can be found in the unit tests.
//
// The protocol does not describe its structure - unlike other binary
// encodings like CBOR or Protobuf - which makes its meaning entirely
// dependent on the supplied message structure.
// In other words, a beserial-blob on its own is garbage data and
// it's impossible to detect what data types it contains and whether it's valid.
//
// Reference Rust implementation of beserial: https://lib.rs/crates/beserial
// The original implementation can be found in https://github.com/nimiq/core-js,
// albeit only manual/handwritten (de)serializers.
package beserial
