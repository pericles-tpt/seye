package tree

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Fiye/stats"
	"github.com/Fiye/utility"
	"github.com/joomcode/errorx"
)

var (
	ignoredDirs   = []string{"proc", "Fiye"}
	chosenHash    = sha256.Size
	MEGABYTE      = 1024 * 1024
	copyBufferLen = 1 * MEGABYTE

	filesAddedForHashing = 0
	noHashOffset         = -1

	timeSpentStating time.Duration
	timeSpentReading time.Duration // also time spent hashing
	buildQ           = [][]FileTree{}
	walkQ            = [][]FileTree{}
	mainDone         = false
	totalBytesRead   float64
	totalFilesRead   int64
	totalFilesStated float64

	readdirTimeTaken time.Duration
	totalFilesFound  = 0
	totalDirsFound   = 0
	idleCount        = 0
)

/*
WARNING: This algorithm doesn't produce valid trees at the moment, tree order past the root is non-deterministic

A walk algorithm that runs `readdir` operations on immediate files (`stat`, "read" and "hash") for multiple dirs
across multiple threads.

Fastest for "shallow" scans, but slower than `WalkTreeIterativeFile` for "comprehensive"
*/
func WalkTreeIterativeDir(rootPath string, isComprehensive bool, walkStats *stats.WalkStats) (t *FileTree) {
	totalFilesFound = 0
	totalDirsFound = 0
	idleCount = 0

	rootPath = strings.TrimSuffix(rootPath, "/")
	var (
		allHashBytes = []byte{}
	)
	walkQ = make([][]FileTree, 1)
	buildQ = make([][]FileTree, 1)
	walkQ[0] = append(walkQ[0], FileTree{BasePath: rootPath})

	numThreads := maxNumThreadsComprehensive
	if !isComprehensive {
		numThreads = maxNumThreadsShallow
	}

	var wg sync.WaitGroup
	wg.Add(numThreads)
	dirJobQueues = make([][]DirJob, numThreads)
	dirJobQueueLocks = make([]sync.Mutex, numThreads)
	for i := 0; i < numThreads; i++ {
		dirJobQueues[i] = []DirJob{}
		go func(n int) {
			for {
				if len(dirJobQueues[n]) > 0 {
					doDirJob(n)
				} else if !mainDone {
					if isComprehensive {
						time.Sleep(idleThreadsForComprehensive)
					} else {
						time.Sleep(idleThreadsForShallow)
					}
					idleCount++
				} else {
					break
				}
			}
			wg.Done()
		}(i)
	}

	threadsBytesRead = make([]int64, numThreads)
	threadsFilesRead = make([]int64, numThreads)
	threadsCopyBuffer = make([][]byte, numThreads)
	for i := range threadsCopyBuffer {
		threadsCopyBuffer[i] = make([]byte, copyBufferLen)
	}
	totalBytesRead = 0
	totalFilesRead = 0
	totalFilesStated = 0
	a := time.Now()
	depth := len(walkQ) - 1
	// Append next level of walkQ, for threads to put their new nodes
	for {
		var t FileTree

		walkQLock.Lock()
		for depth+1 > len(walkQ)-1 {
			walkQ = append(walkQ, []FileTree{})
		}
		walkQLock.Unlock()

		// Problem is what if a thread is assigned from the previous level and not given anything from the new level... this won't trigger
		allThreadsEmpty := true
		for _, tq := range dirJobQueues {
			allThreadsEmpty = allThreadsEmpty && (len(tq) == 0)
		}

		// i.e. stops from moving onto next depth unless finished all items at this depth... we could go faster...
		walkQLock.Lock()
		if len(walkQ[depth]) == 0 && allThreadsEmpty {
			depth++
			fmt.Printf("At depth %d, time taken %d ms\n", depth, time.Since(a).Milliseconds())
		}
		walkQLock.Unlock()
		walkQLock.Lock()
		if len(walkQ[depth]) > 0 {
			walkQ, t = popFront2D(walkQ, depth)
		} else {
			walkQLock.Unlock()
			ae := true
			for _, tq := range dirJobQueues {
				ae = ae && (len(tq) == 0)
			}
			if ae {
				break
			}
			// fmt.Println("Nothing at depth: ", depth)
			time.Sleep(10 * time.Microsecond)
			continue
		}
		walkQLock.Unlock()

		if utility.Contains(ignoredDirs, path.Base(t.BasePath)) {
			continue
		}

		buildQLock.Lock()
		for depth+1 > len(buildQ)-1 {
			buildQ = append(buildQ, []FileTree{})
		}
		thisIndexBuildQ := len(buildQ[depth])
		buildQ[depth] = append(buildQ[depth], t)
		buildQLock.Unlock()

		// 2. Pick a thread to assign the task too
		chosenThread := chooseThreadDir(numThreads, totalDirsFound, isComprehensive)

		// TODO: Record which threads are handling, which dirs, so we can add the resultant `FileTree`s back in correct order

		// 3. Create job for thread + add it to the queue
		pushDirJobOnQueue(DirJob{
			ThisIndexBuildQ: thisIndexBuildQ,
			AllHashByte:     &allHashBytes,
			IsComprehensive: isComprehensive,
			Depth:           depth,
			WalkStats:       walkStats,
		}, chosenThread)

		totalDirsFound++
	}

	mainDone = true
	wg.Wait()

	// var speedStr string
	// if isComprehensive {
	// 	rate, _ := units.NewValue(float64(totalBytesRead), units.Byte).Convert(units.GigaByte)
	// 	speedStr = fmt.Sprintf("read %d files. Read and hashed %.2fGB total at an average rate of %.2fGB/s\n", totalFilesRead, rate.Float(), rate.Float()/time.Since(a).Seconds())
	// } else {
	// 	rate := int(totalFilesStated / time.Since(a).Seconds())
	// 	speedStr = fmt.Sprintf("stated %d files. Retrieved file info at %d files/s", int(totalFilesStated), rate)
	// }
	// fmt.Printf("Traversed %d directories, found %d files, %s\n", totalDirsFound, totalFilesFound, speedStr)

	newBuildQ := []FileTree{}
	for _, arr := range buildQ {
		newBuildQ = append(newBuildQ, arr...)
	}
	buildQ = nil

	tree := constructTreeFromIterativeQ(&newBuildQ)
	tree.AllHash = allHashBytes

	return &tree
}

/*
A walk algorithm that runs file operations (`stat`, "read" and "hash") for multiple files across multiple threads.

Fastest for "comprehensive" scans, but slower than `WalkTreeIterativeDir` for "shallow" scans
*/
func WalkTreeIterativeFile(rootPath string, depth int, isComprehensive bool, walkStats *stats.WalkStats) (t *FileTree) {
	rootPath = strings.TrimSuffix(rootPath, "/")
	var (
		allHashBytes = []byte{}
		walkQ        = []FileTree{{BasePath: rootPath}}
	)

	numThreads := maxNumThreadsComprehensive
	if !isComprehensive {
		numThreads = maxNumThreadsShallow
	}

	var wg sync.WaitGroup
	wg.Add(numThreads)
	fileJobQueues = make([][]FileJob, numThreads)
	fileJobQueueLocks = make([]sync.Mutex, numThreads)
	for i := 0; i < numThreads; i++ {
		fileJobQueues[i] = []FileJob{}
		go func(n int) {
			for {
				if len(fileJobQueues[n]) > 0 {
					doFileJob(n)
				} else if !mainDone {
					time.Sleep(100 * time.Microsecond)
				} else {
					break
				}
			}
			wg.Done()
		}(i)
	}

	threadsBytesRead = make([]int64, numThreads)
	threadsFilesRead = make([]int64, numThreads)
	threadsCopyBuffer = make([][]byte, numThreads)
	for i := range threadsCopyBuffer {
		threadsCopyBuffer[i] = make([]byte, copyBufferLen)
	}
	totalFilesFound := 0
	totalDirs := 0
	totalBytesRead = 0
	totalFilesRead = 0
	totalFilesStated = 0
	// a := time.Now()
	for len(walkQ) > 0 {
		currLevelItems := len(walkQ)
		for i := 0; i < currLevelItems; i++ {
			var t FileTree
			walkQ, t = popFront1D(walkQ)
			startWalk := time.Now()

			if utility.Contains(ignoredDirs, path.Base(t.BasePath)) {
				continue
			}

			tempTime := time.Now()
			pathChildren, err := os.ReadDir(t.BasePath)
			readdirTimeTaken += time.Since(tempTime)
			if err != nil {
				(t).ErrStrings = append((t).ErrStrings, errorx.Decorate(err, "failed to open `tree.BasePath`").Error())
				if depth == 0 {
					return &t
				}
			}

			(t).Comprehensive = isComprehensive
			(t).Depth = depth
			(t).LastVisited = time.Now()

			buildQSLock.Lock()
			buildQS = pushBack1D(buildQS, t)
			thisIndexbuildQS := len(buildQS) - 1
			buildQSLock.Unlock()

			childrenFiles := []os.DirEntry{}
			for _, c := range pathChildren {
				if c.IsDir() {
					fullPath := getFullPath(t.BasePath, c.Name())

					walkQ = pushBack1D(walkQ, FileTree{BasePath: fullPath})
					totalDirs++
				} else if c.Type().IsRegular() {
					childrenFiles = append(childrenFiles, c)
				}
			}

			var (
				fileIndex                 = 0
				lenAllBytesBeforeChildren = len(allHashBytes)
			)
			if isComprehensive {
				allHashBytes = append(allHashBytes, make([]byte, len(childrenFiles)*chosenHash)...)
			}
			for i, cf := range childrenFiles {
				fullPath := getFullPath(t.BasePath, cf.Name())

				// 1. Add file placeholder to parent `Files`
				buildQSLock.Lock()
				buildQS[len(buildQS)-1].Files = append(buildQS[len(buildQS)-1].Files, File{})
				buildQSLock.Unlock()

				// 2. Pick a thread to assign the task too
				chosenThread := chooseThreadFile(numThreads, totalFilesFound, isComprehensive)

				// 3. Create job for thread + add it to the queue
				pushFileJobOnQueue(FileJob{
					FullPath:           fullPath,
					ParentIndexInQueue: thisIndexbuildQS,
					ThisIndexInFiles:   fileIndex,
					Entry:              cf,
					AllHashByte:        &allHashBytes,
					IsComprehensive:    isComprehensive,

					HashOffset: lenAllBytesBeforeChildren + i*chosenHash,
					HashLength: chosenHash,
				}, chosenThread)

				fileIndex++
				totalFilesFound++
			}

			buildQSLock.Lock()
			buildQS[thisIndexbuildQS].NumFilesDirect = int64(len(buildQS[thisIndexbuildQS].Files))
			buildQS[thisIndexbuildQS].TimeTaken = time.Since(startWalk) // This is a bit less acurrate now with MT...
			buildQSLock.Unlock()
		}
		depth++
	}

	mainDone = true
	wg.Wait()

	// var speedStr string
	// if isComprehensive {
	// 	rate, _ := units.NewValue(float64(totalBytesRead), units.Byte).Convert(units.GigaByte)
	// 	speedStr = fmt.Sprintf("read %d files. Read and hashed %.2fGB total at an average rate of %.2fGB/s\n", totalFilesRead, rate.Float(), rate.Float()/time.Since(a).Seconds())
	// } else {
	// 	rate := int(totalFilesStated / time.Since(a).Seconds())
	// 	speedStr = fmt.Sprintf("stated %d files. Retrieved file info at %d files/s", int(totalFilesStated), rate)
	// }
	// fmt.Printf("Traversed %d directories, found %d files, %s\n", totalDirs, totalFilesFound, speedStr)

	// a = time.Now()

	tree := constructTreeFromIterativeQ(&buildQS)

	tree.AllHash = allHashBytes

	return &tree
}
