package structs

type SuperBlock struct {
	SFilesystemType  int32
	SInodesCount     int32
	SBlocksCount     int32
	SFreeBlocksCount int32
	SFreeInodesCount int32
	SMtime           [20]byte
	SUmTime          [20]byte
	SMntCount        int32
	SMagic           int32
	SInodeSize       int32
	SBlockSize       int32
	SFirstIno        int32
	SFirstBlo        int32
	SBmInodeStart    int32
	SBmBlockStart    int32
	SInodeStart      int32
	SBlockStart      int32
}

type Inode struct {
	IUid   int32
	IGid   int32
	ISize  int32
	IAtime [20]byte
	ICtime [20]byte
	IMtime [20]byte
	IBlock [16]int32
	IType  byte
	IPerm  [3]byte
}

type Content struct {
	BName  [12]byte
	BInodo int32
}

type FolderBlock struct {
	BContent [4]Content
}

type FileBlock struct {
	BContent [64]byte
}

type PointerBlock struct {
	BPointers [16]int32
}
