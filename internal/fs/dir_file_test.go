package fs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/structs"
)

func TestMkdirRecursiveAndManyChildren(t *testing.T) {
	resetMount(t)
	path := setupFormattedFS(t)
	actor := rootActor()
	before := testFileSize(t, path)

	if err := Mkdir(path, activePartitionStart(t), "/home/archivos/mia/fase2", true, actor); err != nil {
		t.Fatalf("Mkdir recursive failed: %v", err)
	}
	for i := 1; i <= 10; i++ {
		dirPath := "/home/archivos/mia/a" + string(rune('0'+i))
		if i == 10 {
			dirPath = "/home/archivos/mia/a10"
		}
		if err := Mkdir(path, activePartitionStart(t), dirPath, false, actor); err != nil {
			t.Fatalf("Mkdir child %d failed: %v", i, err)
		}
	}

	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	_, inode, err := ResolvePath(file, sb, "/home/archivos/mia/fase2")
	if err != nil {
		t.Fatalf("Resolve created path failed: %v", err)
	}
	if inode.IType != '0' || inode.IUid != 1 || inode.IGid != 1 || structs.FixedBytesToString(inode.IPerm[:]) != "664" {
		t.Fatalf("unexpected directory inode: %#v", inode)
	}
	block, err := ReadFolderBlock(file, sb, inode.IBlock[0])
	if err != nil {
		t.Fatalf("ReadFolderBlock failed: %v", err)
	}
	if structs.FixedBytesToString(block.BContent[0].BName[:]) != "." || structs.FixedBytesToString(block.BContent[1].BName[:]) != ".." {
		t.Fatalf("missing dot entries: %#v", block)
	}
	miaIndex, miaInode, err := ResolvePath(file, sb, "/home/archivos/mia")
	if err != nil {
		t.Fatalf("Resolve mia failed: %v", err)
	}
	_ = miaIndex
	usedBlocks := len(UsedDirectBlocks(miaInode))
	if usedBlocks < 4 {
		t.Fatalf("expected multiple folder blocks, got %d", usedBlocks)
	}
	if after := testFileSize(t, path); after != before {
		t.Fatalf("disk size changed from %d to %d", before, after)
	}
}

func TestMkdirWithoutParentFails(t *testing.T) {
	resetMount(t)
	path := setupFormattedFS(t)
	if err := Mkdir(path, activePartitionStart(t), "/noexiste/a", false, rootActor()); err == nil {
		t.Fatal("expected missing parent error")
	}
}

func TestMkdirExistingFinalDoesNotDuplicateEntry(t *testing.T) {
	resetMount(t)
	path := setupFormattedFS(t)
	actor := rootActor()

	if err := Mkdir(path, activePartitionStart(t), "/home/archivos/mia/carpeta2/a10", true, actor); err != nil {
		t.Fatalf("initial Mkdir failed: %v", err)
	}
	sbBefore := readSBByPath(t, path)
	err := Mkdir(path, activePartitionStart(t), "/home/archivos/mia/carpeta2/a10", false, actor)
	if err == nil || !strings.Contains(err.Error(), "la carpeta ya existe: /home/archivos/mia/carpeta2/a10") {
		t.Fatalf("expected existing folder error, got %v", err)
	}
	sbAfter := readSBByPath(t, path)
	if sbAfter.SFreeInodesCount != sbBefore.SFreeInodesCount || sbAfter.SFreeBlocksCount != sbBefore.SFreeBlocksCount {
		t.Fatalf("existing mkdir changed free counts: before=%#v after=%#v", sbBefore, sbAfter)
	}

	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	_, parent, err := ResolvePath(file, sb, "/home/archivos/mia/carpeta2")
	if err != nil {
		t.Fatalf("Resolve parent failed: %v", err)
	}
	if got := countDirectoryEntry(t, file, sb, parent, "a10"); got != 1 {
		t.Fatalf("expected one a10 entry, got %d", got)
	}
}

func TestMkdirFinalExistingFileFails(t *testing.T) {
	resetMount(t)
	path := setupFormattedFS(t)
	actor := rootActor()
	if err := Mkfile(path, activePartitionStart(t), "/home/a.txt", true, 10, "", actor); err != nil {
		t.Fatalf("Mkfile failed: %v", err)
	}
	err := Mkdir(path, activePartitionStart(t), "/home/a.txt", false, actor)
	if err == nil || !strings.Contains(err.Error(), "ya existe un archivo con ese nombre: /home/a.txt") {
		t.Fatalf("expected existing file error, got %v", err)
	}
}

func TestMkfileReadOverwriteAndCat(t *testing.T) {
	resetMount(t)
	path := setupFormattedFS(t)
	actor := rootActor()
	before := testFileSize(t, path)

	if err := Mkfile(path, activePartitionStart(t), "/home/b1.txt", true, 75, "", actor); err != nil {
		t.Fatalf("Mkfile 75 failed: %v", err)
	}
	content := readFileFromFS(t, path, "/home/b1.txt")
	if len(content) != 75 || string(content[:10]) != "0123456789" {
		t.Fatalf("unexpected b1 content len=%d content=%q", len(content), string(content))
	}
	if len(UsedDirectBlocks(resolveInode(t, path, "/home/b1.txt"))) != 2 {
		t.Fatal("expected 2 direct blocks for 75 bytes")
	}

	if err := Mkfile(path, activePartitionStart(t), "/home/vacio.txt", true, 0, "", actor); err != nil {
		t.Fatalf("Mkfile empty failed: %v", err)
	}
	if blocks := UsedDirectBlocks(resolveInode(t, path, "/home/vacio.txt")); len(blocks) != 0 {
		t.Fatalf("expected no blocks for empty file, got %#v", blocks)
	}

	if err := Mkfile(path, activePartitionStart(t), "/missing/b2.txt", false, 75, "", actor); err == nil {
		t.Fatal("expected missing parent error")
	}
	if err := Mkfile(path, activePartitionStart(t), "/home/deep/b2.txt", true, 175, "", actor); err != nil {
		t.Fatalf("Mkfile recursive failed: %v", err)
	}
	if len(readFileFromFS(t, path, "/home/deep/b2.txt")) != 175 {
		t.Fatal("expected 175-byte file")
	}

	sbBefore := readSBByPath(t, path)
	if err := Mkfile(path, activePartitionStart(t), "/home/deep/b2.txt", false, 10, "", actor); err != nil {
		t.Fatalf("overwrite shrink failed: %v", err)
	}
	sbAfterShrink := readSBByPath(t, path)
	if sbAfterShrink.SFreeBlocksCount <= sbBefore.SFreeBlocksCount {
		t.Fatalf("expected free blocks to increase after shrink")
	}
	if len(readFileFromFS(t, path, "/home/deep/b2.txt")) != 10 {
		t.Fatal("expected 10-byte overwritten file")
	}
	if err := Mkfile(path, activePartitionStart(t), "/home/deep/b2.txt", false, 175, "", actor); err != nil {
		t.Fatalf("overwrite grow failed: %v", err)
	}
	sbAfterGrow := readSBByPath(t, path)
	if sbAfterGrow.SFreeBlocksCount >= sbAfterShrink.SFreeBlocksCount {
		t.Fatalf("expected free blocks to decrease after grow")
	}

	hostContent := "contenido-host"
	hostPath := filepath.Join(t.TempDir(), "host.txt")
	if err := os.WriteFile(hostPath, []byte(hostContent), 0o644); err != nil {
		t.Fatalf("WriteFile host failed: %v", err)
	}
	if err := Mkfile(path, activePartitionStart(t), "/home/host.txt", false, 200, hostPath, actor); err != nil {
		t.Fatalf("Mkfile cont failed: %v", err)
	}
	if got := string(readFileFromFS(t, path, "/home/host.txt")); got != hostContent {
		t.Fatalf("cont should have priority, got %q", got)
	}

	out, err := Cat(path, activePartitionStart(t), map[string]string{"file": "/home/b1.txt"}, actor)
	if err != nil {
		t.Fatalf("Cat failed: %v", err)
	}
	if len(out) != 75 || strings.ContainsRune(out, '\x00') {
		t.Fatalf("cat should respect ISize, len=%d content=%q", len(out), out)
	}
	out, err = Cat(path, activePartitionStart(t), map[string]string{"file1": "/home/host.txt", "file2": "/home/deep/b2.txt"}, actor)
	if err != nil {
		t.Fatalf("Cat multi failed: %v", err)
	}
	if !strings.HasPrefix(out, hostContent+"\n") {
		t.Fatalf("cat order mismatch: %q", out)
	}

	if err := Mkfile(path, activePartitionStart(t), "/home/big.txt", false, int64(maxFileDataBlocks*BlockSize+1), "", actor); err == nil {
		t.Fatal("expected file capacity error")
	}
	if after := testFileSize(t, path); after != before {
		t.Fatalf("disk size changed from %d to %d", before, after)
	}
}

func TestMkfileSimpleIndirectReadAndShrink(t *testing.T) {
	resetMount(t)
	path := setupFormattedFS(t)
	actor := rootActor()
	before := testFileSize(t, path)
	largeSize := int64(directBlockLimit*BlockSize + 75)

	if err := Mkfile(path, activePartitionStart(t), "/home/grande.txt", true, largeSize, "", actor); err != nil {
		t.Fatalf("Mkfile large failed: %v", err)
	}
	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	sb := activeSuperBlock(t, file)
	index, inode, err := ResolvePath(file, sb, "/home/grande.txt")
	if err != nil {
		t.Fatalf("Resolve large failed: %v", err)
	}
	if inode.IBlock[simpleIndirectIndex] < 0 {
		t.Fatal("expected simple indirect pointer block")
	}
	pointerIndex := inode.IBlock[simpleIndirectIndex]
	pointer, err := ReadPointerBlock(file, sb, pointerIndex)
	if err != nil {
		t.Fatalf("ReadPointerBlock failed: %v", err)
	}
	if pointer.BPointers[0] < 0 {
		t.Fatal("expected first indirect data block")
	}
	content, err := ReadFileContent(file, sb, inode)
	if err != nil {
		t.Fatalf("ReadFileContent failed: %v", err)
	}
	if len(content) != int(largeSize) || string(content[:10]) != "0123456789" {
		t.Fatalf("unexpected large content len=%d", len(content))
	}
	file.Close()

	out, err := Cat(path, activePartitionStart(t), map[string]string{"file": "/home/grande.txt"}, actor)
	if err != nil {
		t.Fatalf("Cat large failed: %v", err)
	}
	if len(out) != int(largeSize) || strings.ContainsRune(out, '\x00') {
		t.Fatalf("cat should read complete indirect content, len=%d", len(out))
	}

	sbBeforeShrink := readSBByPath(t, path)
	if err := Mkfile(path, activePartitionStart(t), "/home/grande.txt", false, 10, "", actor); err != nil {
		t.Fatalf("overwrite shrink large failed: %v", err)
	}
	file, _, err = disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	sb = activeSuperBlock(t, file)
	inode, err = ReadInode(file, sb, index)
	if err != nil {
		t.Fatalf("ReadInode failed: %v", err)
	}
	if inode.IBlock[simpleIndirectIndex] != -1 {
		t.Fatalf("expected pointer block freed, got %d", inode.IBlock[simpleIndirectIndex])
	}
	bitmap, err := ReadBlockBitmap(file, sb)
	if err != nil {
		t.Fatalf("ReadBlockBitmap failed: %v", err)
	}
	if bitmap[pointerIndex] != 0 {
		t.Fatalf("expected pointer block bitmap to be free")
	}
	if sb.SFreeBlocksCount <= sbBeforeShrink.SFreeBlocksCount {
		t.Fatalf("expected free block count to increase after freeing pointer")
	}
	if after := testFileSize(t, path); after != before {
		t.Fatalf("disk size changed from %d to %d", before, after)
	}
}

func TestCatErrorsAndPermissions(t *testing.T) {
	resetMount(t)
	path := setupFormattedFS(t)
	actor := rootActor()
	if err := Mkfile(path, activePartitionStart(t), "/home/a.txt", true, 10, "", actor); err != nil {
		t.Fatalf("Mkfile failed: %v", err)
	}
	if _, err := Cat(path, activePartitionStart(t), map[string]string{"file": "/home/missing.txt"}, actor); err == nil {
		t.Fatal("expected missing file error")
	}
	if _, err := Cat(path, activePartitionStart(t), map[string]string{"file": "/home"}, actor); err == nil {
		t.Fatal("expected cat directory error")
	}

	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	index, inode, err := ResolvePath(file, sb, "/home/a.txt")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	structs.SetPerm(&inode.IPerm, "000")
	if err := WriteInode(file, sb, index, inode); err != nil {
		t.Fatalf("WriteInode failed: %v", err)
	}
	if _, err := Cat(path, activePartitionStart(t), map[string]string{"file": "/home/a.txt"}, Actor{User: "u", UID: 2, GID: 2}); err == nil {
		t.Fatal("expected read permission error")
	}
	if _, err := Cat(path, activePartitionStart(t), map[string]string{"file": "/home/a.txt"}, actor); err != nil {
		t.Fatalf("root should read everything: %v", err)
	}
}

func setupFormattedFS(t *testing.T) string {
	t.Helper()
	path := createMountedPartition(t, 20, 15, "P")
	var noop testWriter
	if err := FormatFromParams(map[string]string{"id": "961A"}, noop); err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	return path
}

func rootActor() Actor {
	return Actor{User: "root", UID: 1, GID: 1}
}

func activePartitionStart(t *testing.T) int64 {
	t.Helper()
	mounted, ok := mount.Global.GetMounted("961A")
	if !ok {
		t.Fatal("expected active mount")
	}
	return int64(mounted.Start)
}

func activeSuperBlock(t *testing.T, file *os.File) structs.SuperBlock {
	t.Helper()
	sb, err := ReadSuperBlock(file, activePartitionStart(t))
	if err != nil {
		t.Fatalf("ReadSuperBlock failed: %v", err)
	}
	return sb
}

func readSBByPath(t *testing.T, path string) structs.SuperBlock {
	t.Helper()
	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	return activeSuperBlock(t, file)
}

func countDirectoryEntry(t *testing.T, file *os.File, sb structs.SuperBlock, inode structs.Inode, name string) int {
	t.Helper()
	count := 0
	for i := 0; i < directBlockLimit && i < len(inode.IBlock); i++ {
		blockIndex := inode.IBlock[i]
		if blockIndex < 0 {
			continue
		}
		block, err := ReadFolderBlock(file, sb, blockIndex)
		if err != nil {
			t.Fatalf("ReadFolderBlock failed: %v", err)
		}
		for _, content := range block.BContent {
			if content.BInodo >= 0 && structs.FixedBytesToString(content.BName[:]) == name {
				count++
			}
		}
	}
	return count
}

func readFileFromFS(t *testing.T, path string, fsPath string) []byte {
	t.Helper()
	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	_, inode, err := ResolvePath(file, sb, fsPath)
	if err != nil {
		t.Fatalf("ResolvePath failed: %v", err)
	}
	content, err := ReadFileContent(file, sb, inode)
	if err != nil {
		t.Fatalf("ReadFileContent failed: %v", err)
	}
	return content
}

func resolveInode(t *testing.T, path string, fsPath string) structs.Inode {
	t.Helper()
	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	sb := activeSuperBlock(t, file)
	_, inode, err := ResolvePath(file, sb, fsPath)
	if err != nil {
		t.Fatalf("ResolvePath failed: %v", err)
	}
	return inode
}

func TestCanReadWriteUGO(t *testing.T) {
	inode := structs.NewEmptyInode()
	inode.IUid = 10
	inode.IGid = 20
	structs.SetPerm(&inode.IPerm, "640")
	if !CanRead(inode, Actor{User: "owner", UID: 10, GID: 1}) || !CanWrite(inode, Actor{User: "owner", UID: 10, GID: 1}) {
		t.Fatal("owner should read/write")
	}
	if !CanRead(inode, Actor{User: "group", UID: 2, GID: 20}) || CanWrite(inode, Actor{User: "group", UID: 2, GID: 20}) {
		t.Fatal("group should read only")
	}
	if CanRead(inode, Actor{User: "other", UID: 2, GID: 3}) || CanWrite(inode, Actor{User: "other", UID: 2, GID: 3}) {
		t.Fatal("other should have no permissions")
	}
	if !CanRead(inode, rootActor()) || !CanWrite(inode, rootActor()) {
		t.Fatal("root should have all permissions")
	}
}

func TestFileWriteDoesNotGrowDisk(t *testing.T) {
	resetMount(t)
	path := setupFormattedFS(t)
	before := testFileSize(t, path)
	if err := Mkfile(path, activePartitionStart(t), "/home/a.txt", true, 100, "", rootActor()); err != nil {
		t.Fatalf("Mkfile failed: %v", err)
	}
	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	after, err := binio.FileSize(file)
	if err != nil {
		t.Fatalf("FileSize failed: %v", err)
	}
	if after != before {
		t.Fatalf("disk grew from %d to %d", before, after)
	}
}
