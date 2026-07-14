// Package rpcvalidation centralizes client-side validation for JSON-RPC
// arguments whose limits are part of the stable Zenon SDK contract.
//
// The helpers in this internal package reject invalid requests before a
// transport call is attempted. They are shared by the ledger and embedded API
// packages so every paginated endpoint applies the same canonical limits.
package rpcvalidation
