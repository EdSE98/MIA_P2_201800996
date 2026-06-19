package fs

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/structs"
)

func Mkfile(diskPath string, partitionStart int64, path string, recursive bool, size int64, contPath string, actor Actor) error {
	if path == "" {
		return fmt.Errorf("mkfile requiere -path")
	}
	content, err := buildFileContent(size, contPath)
	if err != nil {
		return err
	}

	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		return err
	}
	defer file.Close()

	sb, err := ReadSuperBlock(file, partitionStart)
	if err != nil {
		return err
	}
	parentIndex, parentInode, name, err := ResolveParent(file, sb, path)
	if err != nil {
		if !recursive {
			return err
		}
		parts, cleanErr := CleanAbsPath(path)
		if cleanErr != nil {
			return cleanErr
		}
		if len(parts) <= 1 {
			return err
		}
		parentPath := "/" + strings.Join(parts[:len(parts)-1], "/")
		if mkdirErr := Mkdir(diskPath, partitionStart, parentPath, true, actor); mkdirErr != nil {
			return mkdirErr
		}
		sb, err = ReadSuperBlock(file, partitionStart)
		if err != nil {
			return err
		}
		parentIndex, parentInode, name, err = ResolveParent(file, sb, path)
		if err != nil {
			return err
		}
	}
	return CreateOrOverwriteFile(file, partitionStart, &sb, parentIndex, &parentInode, name, content, actor)
}

func CreateOrOverwriteFile(file *os.File, partitionStart int64, sb *structs.SuperBlock, parentIndex int32, parentInode *structs.Inode, name string, content []byte, actor Actor) error {
	if !CanWrite(*parentInode, actor) {
		return fmt.Errorf("permiso de escritura denegado")
	}
	existingIndex, existingInode, ok, err := FindEntryInDirectory(file, *sb, *parentInode, name)
	if err != nil {
		return err
	}
	if ok {
		if existingInode.IType != '1' {
			return fmt.Errorf("existe una carpeta con ese nombre")
		}
		return WriteFileContent(file, partitionStart, sb, existingIndex, &existingInode, content)
	}

	inodeIndex, err := AllocateInode(file, sb)
	if err != nil {
		return err
	}
	inode := structs.NewEmptyInode()
	inode.IUid = actor.UID
	inode.IGid = actor.GID
	inode.IType = '1'
	inode.IAtime = structs.NowDateBytes()
	inode.ICtime = inode.IAtime
	inode.IMtime = inode.IAtime
	structs.SetPerm(&inode.IPerm, "664")
	if err := WriteFileContent(file, partitionStart, sb, inodeIndex, &inode, content); err != nil {
		_ = FreeInode(file, sb, inodeIndex)
		return err
	}
	return AddDirectoryEntry(file, sb, parentIndex, parentInode, name, inodeIndex)
}

func ReadFileContent(file *os.File, sb structs.SuperBlock, inode structs.Inode) ([]byte, error) {
	if inode.IType != '1' {
		return nil, fmt.Errorf("la ruta no es archivo")
	}
	remaining := int(inode.ISize)
	content := make([]byte, 0, remaining)
	blocks, err := FileDataBlockIndices(file, sb, inode)
	if err != nil {
		return nil, err
	}
	for _, blockIndex := range blocks {
		if remaining <= 0 {
			break
		}
		block, err := ReadFileBlock(file, sb, blockIndex)
		if err != nil {
			return nil, err
		}
		take := remaining
		if take > BlockSize {
			take = BlockSize
		}
		content = append(content, block.BContent[:take]...)
		remaining -= take
	}
	if remaining > 0 {
		return nil, fmt.Errorf("contenido de archivo incompleto")
	}
	return content, nil
}

func WriteFileContent(file *os.File, partitionStart int64, sb *structs.SuperBlock, inodeIndex int32, inode *structs.Inode, content []byte) error {
	blocks := splitBytesIntoFileBlocks(content)
	if len(blocks) > maxFileDataBlocks {
		return fmt.Errorf("archivo excede capacidad soportada")
	}
	bitmap, err := ReadBlockBitmap(file, *sb)
	if err != nil {
		return err
	}
	currentBlocks, pointerIndex, _, hasPointer, err := currentFileBlockLayout(file, *sb, *inode)
	if err != nil {
		return err
	}
	needed := len(blocks)
	if needed > len(currentBlocks) {
		allocated, err := AllocateBlocks(bitmap, needed-len(currentBlocks))
		if err != nil {
			return err
		}
		currentBlocks = append(currentBlocks, allocated...)
		sb.SFreeBlocksCount -= int32(len(allocated))
	}
	if needed < len(currentBlocks) {
		toFree := currentBlocks[needed:]
		for _, blockIndex := range toFree {
			if blockIndex < 0 || blockIndex >= int32(len(bitmap)) {
				return fmt.Errorf("bloque fuera de rango: %d", blockIndex)
			}
			bitmap[blockIndex] = 0
			if err := WriteFileBlock(file, *sb, blockIndex, structs.NewEmptyFileBlock()); err != nil {
				return err
			}
		}
		sb.SFreeBlocksCount += int32(len(toFree))
		currentBlocks = currentBlocks[:needed]
	}

	needsPointer := needed > directBlockLimit
	if needsPointer && !hasPointer {
		allocated, err := AllocateBlocks(bitmap, 1)
		if err != nil {
			return err
		}
		pointerIndex = allocated[0]
		hasPointer = true
		sb.SFreeBlocksCount--
	}
	if !needsPointer && hasPointer {
		if pointerIndex < 0 || pointerIndex >= int32(len(bitmap)) {
			return fmt.Errorf("bloque de apuntadores fuera de rango: %d", pointerIndex)
		}
		bitmap[pointerIndex] = 0
		sb.SFreeBlocksCount++
		if err := WritePointerBlock(file, *sb, pointerIndex, structs.NewEmptyPointerBlock()); err != nil {
			return err
		}
		pointerIndex = -1
		hasPointer = false
	}

	for i := range inode.IBlock {
		inode.IBlock[i] = -1
	}
	for i := 0; i < needed && i < directBlockLimit; i++ {
		inode.IBlock[i] = currentBlocks[i]
	}
	if needsPointer {
		pointer := structs.NewEmptyPointerBlock()
		inode.IBlock[simpleIndirectIndex] = pointerIndex
		for i := directBlockLimit; i < needed; i++ {
			pointer.BPointers[i-directBlockLimit] = currentBlocks[i]
		}
		if err := WritePointerBlock(file, *sb, pointerIndex, pointer); err != nil {
			return err
		}
	}
	for i, blockIndex := range currentBlocks[:needed] {
		if err := WriteFileBlock(file, *sb, blockIndex, blocks[i]); err != nil {
			return err
		}
	}
	inode.ISize = int32(len(content))
	inode.IMtime = structs.NowDateBytes()
	sb.SFirstBlo = RecalculateFirstFree(bitmap)
	if err := WriteBitmap(file, int64(sb.SBmBlockStart), bitmap); err != nil {
		return err
	}
	if err := WriteInode(file, *sb, inodeIndex, *inode); err != nil {
		return err
	}
	return WriteSuperBlock(file, partitionStart, *sb)
}

func FileDataBlockIndices(file *os.File, sb structs.SuperBlock, inode structs.Inode) ([]int32, error) {
	blocks, _, _, _, err := currentFileBlockLayout(file, sb, inode)
	return blocks, err
}

func currentFileBlockLayout(file *os.File, sb structs.SuperBlock, inode structs.Inode) ([]int32, int32, structs.PointerBlock, bool, error) {
	blocks := make([]int32, 0, maxFileDataBlocks)
	for i := 0; i < directBlockLimit && i < len(inode.IBlock); i++ {
		if inode.IBlock[i] >= 0 {
			blocks = append(blocks, inode.IBlock[i])
		}
	}
	pointerIndex := int32(-1)
	pointer := structs.NewEmptyPointerBlock()
	hasPointer := false
	if simpleIndirectIndex < len(inode.IBlock) && inode.IBlock[simpleIndirectIndex] >= 0 {
		pointerIndex = inode.IBlock[simpleIndirectIndex]
		var err error
		pointer, err = ReadPointerBlock(file, sb, pointerIndex)
		if err != nil {
			return nil, -1, structs.PointerBlock{}, false, err
		}
		hasPointer = true
		for _, blockIndex := range pointer.BPointers {
			if blockIndex >= 0 {
				blocks = append(blocks, blockIndex)
			}
		}
	}
	return blocks, pointerIndex, pointer, hasPointer, nil
}

func Cat(diskPath string, partitionStart int64, params map[string]string, actor Actor) (string, error) {
	paths := catPaths(params)
	if len(paths) == 0 {
		return "", fmt.Errorf("cat requiere -file")
	}
	file, _, err := disk.OpenReadWrite(diskPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	sb, err := ReadSuperBlock(file, partitionStart)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	for i, path := range paths {
		_, inode, err := ResolvePath(file, sb, path)
		if err != nil {
			return "", err
		}
		if inode.IType != '1' {
			return "", fmt.Errorf("la ruta no es archivo: %s", path)
		}
		if !CanRead(inode, actor) {
			return "", fmt.Errorf("permiso de lectura denegado")
		}
		content, err := ReadFileContent(file, sb, inode)
		if err != nil {
			return "", err
		}
		if i > 0 {
			out.WriteByte('\n')
		}
		out.Write(content)
	}
	return out.String(), nil
}

func buildFileContent(size int64, contPath string) ([]byte, error) {
	if contPath != "" {
		return os.ReadFile(contPath)
	}
	if size < 0 {
		return nil, fmt.Errorf("size no puede ser negativo")
	}
	pattern := []byte("0123456789")
	content := make([]byte, size)
	for i := range content {
		content[i] = pattern[i%len(pattern)]
	}
	return content, nil
}

func splitBytesIntoFileBlocks(content []byte) []structs.FileBlock {
	if len(content) == 0 {
		return nil
	}
	blocks := make([]structs.FileBlock, 0, (len(content)+BlockSize-1)/BlockSize)
	for start := 0; start < len(content); start += BlockSize {
		end := start + BlockSize
		if end > len(content) {
			end = len(content)
		}
		block := structs.NewEmptyFileBlock()
		copy(block.BContent[:], content[start:end])
		blocks = append(blocks, block)
	}
	return blocks
}

func catPaths(params map[string]string) []string {
	var result []string
	if value, ok := params["file"]; ok {
		result = append(result, value)
	}
	type numbered struct {
		index int
		path  string
	}
	var numberedPaths []numbered
	for key, value := range params {
		if !strings.HasPrefix(key, "file") || key == "file" {
			continue
		}
		index, err := strconv.Atoi(strings.TrimPrefix(key, "file"))
		if err != nil {
			continue
		}
		numberedPaths = append(numberedPaths, numbered{index: index, path: value})
	}
	sort.Slice(numberedPaths, func(i, j int) bool {
		return numberedPaths[i].index < numberedPaths[j].index
	})
	for _, item := range numberedPaths {
		result = append(result, item.path)
	}
	return result
}
