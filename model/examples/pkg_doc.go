// Package examples provides types and functions to facilitate the examples and test code in the model package.
//
// Important take aways from this example package are the instantiation of the global Models variable and registering
// type(s) with it.
//
// Model registration is performed via an init() function:
//	func init() {
// 		// Somewhere in your application you need to register all types to be used as models.
// 		Models.Register(&Address{})
// 		Models.Register(&Person{})
// 		Models.Register(&PersonAddress{})
// 	}
//
// Also important is the struct definition for the type Address; the struct definition combined with the Models global
// defines how SQL statements are generated and the expected column names in the generated SQL.
package examples
