# Packages
## main
The entrypoint for the application, initialises required state for it to run

## config
Contains logic for loading and accessing config data

## command
...

## tree
Contains functions related to the `FileTree` structure (generated when a scan is performed)
The main functionality it contains is:
- structure: tree structure
- file: reading/writing tree data
- walk: generating the tree, by walking through its directories (currently iterative, BF approach)

# file
Contains the structure of a `File` which is contained in an array in the `FileTree`

# stats
Contains the logic and structure for `WalkStats` which is passed to a tree "walk" function to
(optionally) collect additional data during traversal i.e:
- basic information about duplicates
- basic information about the largest files

# diff
Handles tree comparison and diffing, diff functions will generate a `TreeDiff` which also contains
an array of `FileDiff`, it contains functionality for:
- add: adding a diff to a tree such that (tree_a - tree_b = diff_a_b -> tree_a + diff_a_b = tree_b)
- compare: populates `ScanDiff` which contains all differences between two `FileTree`
- diff: generates a `TreeDiff` from two `FileTree`
- file: read/write operations for `ScanDiff`
- structure: contains the structure of the aforementioned structs

# records
Handles recordkeeping for completed "scans" (i.e. `FileTree` generation) and "diffs", so that basic 
information can be obtained about the completed tasks, without loading the file

# utility
Contains general utility functions