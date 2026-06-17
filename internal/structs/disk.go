package structs

type MBR struct {
	MbrTamano        int32
	MbrFechaCreacion [20]byte
	MbrDskSignature  int32
	DskFit           byte
	MbrPartitions    [4]Partition
}

type Partition struct {
	PartStatus      byte
	PartType        byte
	PartFit         byte
	PartStart       int32
	PartSize        int32
	PartName        [16]byte
	PartCorrelative int32
	PartID          [4]byte
}

type EBR struct {
	PartMount byte
	PartFit   byte
	PartStart int32
	PartSize  int32
	PartNext  int32
	PartName  [16]byte
}
