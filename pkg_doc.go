// Package sqlh provides some simple utility for database/sql.
//
// Refer to examples under Scanner.Select for row scanning.
//
// Refer to subdirectory model for the model abstraction layer.
//
// sqlh and associated packages use reflect but nearly all of the heavy lifting is offloaded
// to set @ https://www.github.com/nofeaturesonlybugs/set
//
// Both sqlh.Scanner and model.Models use set.Mapper; the examples typically demonstrate
// instantiating a set.Mapper.  If you design your Go destinations and models well
// then ideally your app will only need a single set.Mapper, a single sqlh.Scanner,
// and a single model.Models.  Since all of these can be instantiated without a database
// connection you may wish to define them as globals and register models as part
// of an init() function.
//
// set.Mapper supports some additional flexibility not shown in the examples for this package;
// if your project has extra convoluted Go structs in your database layer you may want to consult
// the package documentation for package set.
package sqlh
