package configs

var FilenameCharMap = make(map[string]string)

var (
	// StoragesLoaded loaded success if empty
	StoragesLoaded = false
	// 单个Buffer最大限制
	MaxBufferLimit = 16 * 1024 * 1024
	// 超过该阈值的Buffer将使用 mmap 分配，可主动释放内存
	MmapThreshold = 4 * 1024 * 1024
)
