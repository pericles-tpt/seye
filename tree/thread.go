package tree

import (
	"crypto/sha256"
	"io"
	"math"
	"os"
	"sync"
	"time"

	"github.com/joomcode/errorx"
	"github.com/pericles-tpt/seye/stats"
	"github.com/pericles-tpt/seye/utility"
)

/*
A "job" assigned to a thread's queue for it to perform a `readdir`, as well
as any subsequent `stat`, "read" and "hash" operations required on files in
the IMMEDIATE directory.

The 'path' for the `DirJob` to operate on is located in `buildQ[ThisIndexInBuildQ]`
*/
type DirJob struct {
	ThisIndexBuildQ int
	AllHashByte     *[]byte
	IsComprehensive bool
	Depth           int
	WalkStats       *stats.WalkStats
}

/*
A "job" assigned to a thread's queue, for it to perform a `stat`, "read" and
"hash" operations on a file, putting the result in:

The 'path' for the `FileJob` to operate on is located in `buildQ[ThisIndexInBuildQ]`
(where it adds the file to it's `.Files` properties and updates some other props)
*/
type FileJob struct {
	FullPath           string
	ParentIndexInQueue int
	ThisIndexInFiles   int
	Entry              os.DirEntry
	File               File
	ErrStrings         []string
	IsComprehensive    bool
	LastModified       time.Time
	AllHashByte        *[]byte

	WalkStats  *stats.WalkStats
	HashOffset int
	HashLength int
}

var (
	maxNumThreadsComprehensive  = 8
	maxNumThreadsShallow        = 8
	idleThreadsForComprehensive = 50 * time.Microsecond
	idleThreadsForShallow       = 50 * time.Microsecond
	buildQLock                  = sync.Mutex{}
	walkQLock                   = sync.Mutex{}
	threadsBytesRead            = []int64{}
	threadsFilesRead            = []int64{}

	dirJobQueues      = [][]DirJob{}
	dirJobQueueLocks  = []sync.Mutex{}
	threadsCopyBuffer = [][]byte{}
	allHashLock       = sync.Mutex{}

	fileJobQueues     = [][]FileJob{}
	fileJobQueueLocks = []sync.Mutex{}
	buildQS           = []FileTree{}
	buildQSLock       = sync.Mutex{}

	walkLock = sync.Mutex{}
)

/*
Performs "read" and "hash" operations on a file, supports MT
*/
func readHashFile(rl ReadLocation, threadNum int) (utility.HashLocation, []string) {
	var errStrings []string
	hl := utility.InitialiseHashLocation()

	timer := time.Now()
	fTmp, err := os.OpenFile(rl.FullPath, os.O_RDONLY, 0400)
	if err != nil {
		errStrings = append(errStrings, errorx.Decorate(err, "failed to open file to get `Hash` for 'Comprehensive' scan").Error())
	} else if rl.Size > 0 {
		h := sha256.New()
		n, err := io.CopyBuffer(h, fTmp, threadsCopyBuffer[threadNum])
		if err != nil {
			errStrings = append(errStrings, errorx.Decorate(err, "failed to read file to get `Hash` for 'Comprehensive' scan").Error())
		} else {
			timeSpentReading += time.Since(timer)
			totalBytesRead += float64(n)
			threadsBytesRead[threadNum] += n
			threadsFilesRead[threadNum]++
			totalFilesRead++

			hashedBytes := h.Sum(nil)
			for i := 0; i < rl.HashLength; i++ {
				(*rl.AllHashByte)[rl.HashOffset+i] = hashedBytes[i]
			}
			hl.Type = utility.SHA256
			hl.HashOffset = rl.HashOffset
			hl.HashLength = rl.HashLength

			walkLock.Lock()
			if rl.WalkStats != nil {
				rl.WalkStats.UpdateDuplicates(hashedBytes[:], rl.Size, rl.FullPath)
			}
			walkLock.Unlock()
		}
	}
	defer fTmp.Close()
	return hl, errStrings
}

/*
 */
func doFileJob(threadNum int) {
	// Dequeue file job
	fileJobQueueLocks[threadNum].Lock()
	currJob := fileJobQueues[threadNum][0]
	fileJobQueueLocks[threadNum].Unlock()

	currJob.File = File{
		Name: currJob.FullPath,
		Hash: utility.InitialiseHashLocation(),
	}

	timer := time.Now()
	// stat(), syscall
	fStat, err := currJob.Entry.Info()
	if err != nil {
		currJob.ErrStrings = append(currJob.ErrStrings, errorx.Decorate(err, "failed to stat file").Error())
	} else {
		timeSpentStating += time.Since(timer)
		totalFilesStated++
		currJob.File.LastModified = fStat.ModTime()
		currJob.File.Size = fStat.Size()
		if currJob.IsComprehensive {
			currJob.File.Hash = utility.InitialiseHashLocation()
			errStrings := []string{}
			// do "read" and "hash" of file
			currJob.File.Hash, errStrings = getFileDataS(ReadLocation{
				WalkStats:   currJob.WalkStats,
				HashOffset:  currJob.HashOffset,
				HashLength:  currJob.HashLength,
				FullPath:    currJob.FullPath,
				Size:        fStat.Size(),
				AllHashByte: currJob.AllHashByte,
			}, threadNum)
			if len(errStrings) > 0 {
				currJob.ErrStrings = append(currJob.ErrStrings, errStrings...)
			}
		}
	}

	walkLock.Lock()
	if currJob.WalkStats != nil {
		currJob.WalkStats.UpdateLargestFiles(stats.BasicFile{Path: currJob.FullPath, Size: currJob.File.Size})
	}
	walkLock.Unlock()

	buildQSLock.Lock()
	buildQS[currJob.ParentIndexInQueue].Files[currJob.ThisIndexInFiles] = currJob.File
	buildQS[currJob.ParentIndexInQueue].ErrStrings = append(buildQS[currJob.ParentIndexInQueue].ErrStrings, currJob.ErrStrings...)
	buildQS[currJob.ParentIndexInQueue].LastModifiedDirect = utility.GetNewestTime(buildQS[currJob.ParentIndexInQueue].LastModifiedDirect, currJob.File.LastModified)
	buildQS[currJob.ParentIndexInQueue].SizeDirect += currJob.File.Size
	buildQSLock.Unlock()

	dequeueFileJob(threadNum)
}

func pushFileJobOnQueue(j FileJob, threadNum int) {
	fileJobQueueLocks[threadNum].Lock()
	fileJobQueues[threadNum] = append(fileJobQueues[threadNum], j)
	fileJobQueueLocks[threadNum].Unlock()
}

func dequeueFileJob(threadNum int) {
	fileJobQueueLocks[threadNum].Lock()
	fileJobQueues[threadNum] = fileJobQueues[threadNum][1:]
	fileJobQueueLocks[threadNum].Unlock()
}

func doDirJob(threadNum int) {
	// Dequeue dir job
	dirJobQueueLocks[threadNum].Lock()
	currJob := dirJobQueues[threadNum][0]
	dirJobQueueLocks[threadNum].Unlock()

	startWalk := time.Now()

	newNodesDepth := currJob.Depth + 1

	buildQLock.Lock()
	currTree := buildQ[currJob.Depth][currJob.ThisIndexBuildQ]
	buildQLock.Unlock()

	currTree.Comprehensive = currJob.IsComprehensive

	pathChildren, err := os.ReadDir(currTree.BasePath)
	currTree.ErrStrings = []string{}
	if err != nil {
		currTree.ErrStrings = append(currTree.ErrStrings, errorx.Decorate(err, "failed to open `BasePath`").Error())
	}

	currTree.Depth = currJob.Depth
	currTree.LastVisited = time.Now()

	childrenFiles := []os.DirEntry{}
	for _, c := range pathChildren {
		if c.IsDir() {
			fullPath := getFullPath(currTree.BasePath, c.Name())

			walkQLock.Lock()
			walkQ[newNodesDepth] = append(walkQ[newNodesDepth], FileTree{BasePath: fullPath})
			walkQLock.Unlock()
		} else if c.Type().IsRegular() {
			childrenFiles = append(childrenFiles, c)
		}
	}

	allHashLock.Lock()
	lenAllBytesBeforeChildren := len(*currJob.AllHashByte)
	if currTree.Comprehensive {
		*currJob.AllHashByte = append(*currJob.AllHashByte, make([]byte, len(childrenFiles)*chosenHash)...)
	}
	allHashLock.Unlock()
	for i, cf := range childrenFiles {
		fullPath := getFullPath(currTree.BasePath, cf.Name())

		// 1. Add file placeholder to parent `Files`
		buildQLock.Lock()
		buildQ[currJob.Depth][currJob.ThisIndexBuildQ].Files = append(buildQ[currJob.Depth][currJob.ThisIndexBuildQ].Files, File{})
		buildQLock.Unlock()

		// Do the stat
		nf := File{
			Name: fullPath,
			Hash: utility.InitialiseHashLocation(),
		}

		timer := time.Now()
		fStat, err := cf.Info()
		if err != nil {
			currTree.ErrStrings = append(currTree.ErrStrings, errorx.Decorate(err, "failed to stat file").Error())
		} else {
			timeSpentStating += time.Since(timer)
			totalFilesStated++
			nf.LastModified = fStat.ModTime()
			nf.Size = fStat.Size()
			if currTree.Comprehensive {
				nf.Hash = utility.InitialiseHashLocation()

				var errStrings []string
				nf.Hash, errStrings = readHashFile(ReadLocation{
					WalkStats:   currJob.WalkStats,
					HashOffset:  lenAllBytesBeforeChildren + i*chosenHash,
					HashLength:  chosenHash,
					FullPath:    fullPath,
					Size:        nf.Size,
					AllHashByte: currJob.AllHashByte,
				}, threadNum)
				currTree.ErrStrings = append(currTree.ErrStrings, errStrings...)
			}
		}

		walkLock.Lock()
		if currJob.WalkStats != nil {
			currJob.WalkStats.UpdateLargestFiles(stats.BasicFile{Path: nf.Name, Size: nf.Size})
		}
		walkLock.Unlock()

		buildQLock.Lock()
		buildQ[currJob.Depth][currJob.ThisIndexBuildQ].Files[i] = nf
		buildQ[currJob.Depth][currJob.ThisIndexBuildQ].ErrStrings = append(buildQ[currJob.Depth][currJob.ThisIndexBuildQ].ErrStrings, currTree.ErrStrings...)
		buildQ[currJob.Depth][currJob.ThisIndexBuildQ].LastModifiedDirect = utility.GetNewestTime(buildQ[currJob.Depth][currJob.ThisIndexBuildQ].LastModifiedDirect, nf.LastModified)
		buildQ[currJob.Depth][currJob.ThisIndexBuildQ].SizeDirect += nf.Size
		buildQLock.Unlock()

		totalFilesFound++
	}

	buildQLock.Lock()
	buildQ[currJob.Depth][currJob.ThisIndexBuildQ].NumFilesDirect = int64(len(buildQ[currJob.Depth][currJob.ThisIndexBuildQ].Files))
	buildQ[currJob.Depth][currJob.ThisIndexBuildQ].TimeTaken = time.Since(startWalk) // This is a bit less acurrate now with MT...
	buildQLock.Unlock()

	dequeueDirJob(threadNum)
}

func pushDirJobOnQueue(j DirJob, threadNum int) {
	dirJobQueueLocks[threadNum].Lock()
	dirJobQueues[threadNum] = append(dirJobQueues[threadNum], j)
	dirJobQueueLocks[threadNum].Unlock()
}

func dequeueDirJob(threadNum int) {
	dirJobQueueLocks[threadNum].Lock()
	dirJobQueues[threadNum] = dirJobQueues[threadNum][1:]
	dirJobQueueLocks[threadNum].Unlock()
}

func getFileDataS(rl ReadLocation, threadNum int) (utility.HashLocation, []string) {
	var errStrings []string
	hl := utility.InitialiseHashLocation()

	timer := time.Now()
	fTmp, err := os.OpenFile(rl.FullPath, os.O_RDONLY, 0400)
	if err != nil {
		errStrings = append(errStrings, errorx.Decorate(err, "failed to open file to get `Hash` for 'Comprehensive' scan").Error())
	} else if rl.Size > 0 {
		h := sha256.New()
		n, err := io.CopyBuffer(h, fTmp, threadsCopyBuffer[threadNum])
		if err != nil {
			errStrings = append(errStrings, errorx.Decorate(err, "failed to read file to get `Hash` for 'Comprehensive' scan").Error())
		} else {
			timeSpentReading += time.Since(timer)
			totalBytesRead += float64(n)
			threadsBytesRead[threadNum] += n
			threadsFilesRead[threadNum]++
			totalFilesRead++
			hashedBytes := h.Sum(nil)
			for i := 0; i < rl.HashLength; i++ {
				(*rl.AllHashByte)[rl.HashOffset+i] = hashedBytes[i]
			}
			hl.Type = utility.SHA256
			hl.HashOffset = rl.HashOffset
			hl.HashLength = rl.HashLength

			walkLock.Lock()
			if rl.WalkStats != nil {
				rl.WalkStats.UpdateDuplicates(hashedBytes[:], rl.Size, rl.FullPath)
			}
			walkLock.Unlock()
		}
	}
	defer fTmp.Close()
	return hl, errStrings
}

func chooseThreadDir(numThreads, totalDirsFound int, isComprehensive bool) int {
	chosen := int(math.Mod(float64(totalDirsFound), float64(numThreads)))
	if isComprehensive {
		leastJobs := math.MaxInt
		for j, q := range dirJobQueues {
			if len(q) < leastJobs {
				leastJobs = len(q)
				chosen = j
			}
		}
	}
	return chosen
}

func chooseThreadFile(numThreads, totalFilesFound int, isComprehensive bool) int {
	chosen := int(math.Mod(float64(totalFilesFound), float64(numThreads)))
	if isComprehensive {
		leastJobs := math.MaxInt
		for j, q := range fileJobQueues {
			if len(q) < leastJobs {
				leastJobs = len(q)
				chosen = j
			}
		}
	}
	return chosen
}
