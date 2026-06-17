package structs

func NewEmptyPartition() Partition {
	return Partition{
		PartStatus:      '0',
		PartStart:       -1,
		PartSize:        0,
		PartCorrelative: 0,
	}
}

func NewEmptyMBR() MBR {
	mbr := MBR{}
	for i := range mbr.MbrPartitions {
		mbr.MbrPartitions[i] = NewEmptyPartition()
	}
	return mbr
}

func NewEmptyEBR() EBR {
	return EBR{
		PartStart: -1,
		PartSize:  0,
		PartNext:  -1,
	}
}

func NewEmptyInode() Inode {
	inode := Inode{}
	for i := range inode.IBlock {
		inode.IBlock[i] = -1
	}
	return inode
}

func NewEmptyFolderBlock() FolderBlock {
	block := FolderBlock{}
	for i := range block.BContent {
		block.BContent[i].BInodo = -1
	}
	return block
}

func NewEmptyFileBlock() FileBlock {
	return FileBlock{}
}

func NewEmptyPointerBlock() PointerBlock {
	block := PointerBlock{}
	for i := range block.BPointers {
		block.BPointers[i] = -1
	}
	return block
}
