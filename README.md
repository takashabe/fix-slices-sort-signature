# fix-slices-sort-signature

This tool updates the signature of comparison functions passed to `slices.SortFunc` and `slices.SortStableFunc` in the old x/exp/slices package to the new standard.

For more details, see: [golang/go#61374](https://github.com/golang/go/issues/61374).

## Usage

To use this tool, you simply run it on your Go source files. It will automatically convert the old comparison function signatures to the new format using the `cmp` package for comparison.

:memo: TODO: Installation, usage CLI

## Example

Here is an example of how the tool modifies your code:

```diff
diff --git a/examples/main.go b/examples/main.go
index 19b372c..17e0b8b 100644
--- a/examples/main.go
+++ b/examples/main.go
@@ -1,6 +1,7 @@
 package main

 import (
+	"cmp"
 	"slices"
 	"time"
 )

 func main() {
 	as := []A{
 		{"b"},
 		{"a"},
 		{"c"},
 	}

-	slices.SortFunc(as, func(a, b A) bool {
-		return a.Name < b.Name
+	slices.SortFunc(as, func(a, b A) int {
+		return cmp.Compare(a.Name, b.Name)
 	})

-	slices.SortStableFunc(as, func(a, b A) bool {
-		return a.Name > b.Name
+	slices.SortStableFunc(as, func(a, b A) int {
+		return cmp.Compare(b.Name, a.Name)
 	})

 	// slices.SortFunc(as, func(a, b A) bool {
```

In this example:

- The comparison functions for `slices.SortFunc` and `slices.SortStableFunc` are changed from returning a `bool` to returning an `int`.
- The `cmp.Compare` function from the `cmp` package is used for the comparisons, which returns a negative number if the first argument is less, a positive number if it is greater, and zero if they are equal.
