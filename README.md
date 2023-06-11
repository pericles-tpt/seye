# Seye (abbr. StorageEye)
`seye` allows you to scan one or more directories to identify current characteristics (largest size, number of files, duplicates, etc) as well as the change in its disk usage over time.

**WARNING**: This program is currently in a prototype state, some of the INTENDED functionality mentioned below may not be working as intended until testing is complete

## Building `seye`
Binary releases are not available yet to reinforce the prototype state of the code. However you can build the code from source:
1. Install `go`, by following these instructions: https://go.dev/doc/install
2. Clone this repo
2. In the root directory of this repo run `go build`, to generate a binary called `seye` that you can run from your terminal

## Information Gathered by Fiye
"Shallow" scans allow you to:
- get the largest files in a directory
- get the total size of a directory and the number of files within it
- get the change in disk usage of the target directory and its subdirectories between two points in time

"Comprehensive" scans provide additional data, including:
- the location of duplicates
- whether any files and/or directories have changed within the target directory, between two points in time 

## Shallow vs Comprehensive Scans
"Shallow" scans will collect information about directories and basic file information accessible with the `stat` system call, namely:
- file size
- last modified time

"Comprehensive" scans will collect the same information as "shallow" scans in addition to storing hashes (using the SHA256 algorithm) of any files that `seye` has permission to read.

## How it Works
When you run a `seye` `scan` on a specified directory it will traverse that directory and all its files and subdirectories.

In order to use the least disk space possible, `seye` only stores the:
- first full scan ("comprehensive")
- last full scan ("comprehensive" or "shallow")
- "diffs" of every scan between the first and last scans ("comprehensive" ONLY if the scans being "diff"ed are both "comprehensive")

This means for 'n' scans, the number of files generated are: 
$$1 * fullComprehensive + 1 * fullComprehensiveOrShallow + (n-1) * diffComprehensiveOrShallow$$

### Disk Usage
The amount of disk space used by `seye`, is dependent on a number of factors including:
- number of files in a target directory
  - how many files are non-empty 
- number of subdirectories in a target directory
- how many scans have been completed
- whether the last scan was "comprehensive"
- time between scans (i.e. higher -> more changes (maybe) -> larger diffs)
- if scans being "diff"ed are both "comprehensive" -> larger diff
- whether the last scan is comprehensve
- time between the first and last scan

## Limitations
- `seye` is only able to get information and hash files that it has permission to access. Files and directories that are not accessible to the user will not be scanned by `seye`
- `seye` can use a significant amount of memory for scans whilst generating them and the resultant scan files can be quite large
(for 2M files and 500k directories a comprehensive scan was ~600 - 700MB and a shallow scan was ~500 - 600MB)
- Once again `seye` is in a PROTOTYPE state, some functionality may not work as expected and there are likely bugs that still need to be ironed out.

## Current Tasks
- Finish writing tests in `diff_test.go`
- Correct classification of 'added', 'changed', 'renamed' and 'removed' changes
- Support for outputting data to files (text, csv, etc)

## Acknowledgement
- [@anthonyklepac](https://github.com/anthonyklepac) for his contributions to ideation and protoyping for an [earlier version of this program](https://github.com/pericles-tpt/StorageEye)
