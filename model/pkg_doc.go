// Package model allows Go structs to behave as database models.
//
// While this package exports several types the only one you currently need
// to be concerned with is type Models.  All of the examples in this package
// use a global instance of Models defined in the examples subpackage; you may
// refer to that global instance for an instantiation example.
//
// Note that in the examples for this package when you see examples.Models
// or examples.Connect() it is referring the examples subdirectory for
// this package and NOT the subdirectory for sqlh (i.e. both sqlh and sqlh/model
// have an examples subdirectory.)
package model
