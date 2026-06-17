package fs

import (
	"bytes"
	"encoding/binary"
	"path/filepath"
	"strings"
	"testing"

	"mia_p1_201800996/internal/binio"
	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/partition"
	"mia_p1_201800996/internal/structs"
)

func TestFormatEXT2InitializesRootAndUsers(t *testing.T) {
	resetMount(t)
	path := createMountedPartition(t, 10, 5, "P")
	before := testFileSize(t, path)

	var out bytes.Buffer
	if err := Format(FormatOptions{ID: "961A", Type: "full"}, &out); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()

	mounted, ok := mount.Global.GetMounted("961a")
	if !ok {
		t.Fatal("expected mounted partition")
	}
	sb, err := ReadSuperBlock(file, int64(mounted.Start))
	if err != nil {
		t.Fatalf("ReadSuperBlock failed: %v", err)
	}

	inodeSize := int32(binary.Size(structs.Inode{}))
	if sb.SMagic != 0xEF53 {
		t.Fatalf("SMagic = %#x, want 0xEF53", sb.SMagic)
	}
	if sb.SFilesystemType != 2 {
		t.Fatalf("SFilesystemType = %d, want 2", sb.SFilesystemType)
	}
	if sb.SInodesCount <= 0 {
		t.Fatalf("SInodesCount = %d, want > 0", sb.SInodesCount)
	}
	if sb.SBlocksCount != 3*sb.SInodesCount {
		t.Fatalf("SBlocksCount = %d, want %d", sb.SBlocksCount, 3*sb.SInodesCount)
	}
	if sb.SFreeInodesCount != sb.SInodesCount-2 {
		t.Fatalf("SFreeInodesCount = %d, want %d", sb.SFreeInodesCount, sb.SInodesCount-2)
	}
	if sb.SFreeBlocksCount != sb.SBlocksCount-2 {
		t.Fatalf("SFreeBlocksCount = %d, want %d", sb.SFreeBlocksCount, sb.SBlocksCount-2)
	}
	if sb.SInodeSize != inodeSize {
		t.Fatalf("SInodeSize = %d, want %d", sb.SInodeSize, inodeSize)
	}
	if sb.SBlockSize != 64 || sb.SFirstIno != 2 || sb.SFirstBlo != 2 {
		t.Fatalf("unexpected block/free indexes: block=%d firstIno=%d firstBlo=%d", sb.SBlockSize, sb.SFirstIno, sb.SFirstBlo)
	}

	inodeBitmap, err := ReadBitmap(file, int64(sb.SBmInodeStart), sb.SInodesCount)
	if err != nil {
		t.Fatalf("Read inode bitmap failed: %v", err)
	}
	blockBitmap, err := ReadBitmap(file, int64(sb.SBmBlockStart), sb.SBlocksCount)
	if err != nil {
		t.Fatalf("Read block bitmap failed: %v", err)
	}
	assertBitmapPrefix(t, inodeBitmap)
	assertBitmapPrefix(t, blockBitmap)

	rootInode, err := ReadInode(file, sb, 0)
	if err != nil {
		t.Fatalf("Read root inode failed: %v", err)
	}
	if rootInode.IType != '0' || rootInode.IUid != 1 || rootInode.IGid != 1 || rootInode.IBlock[0] != 0 {
		t.Fatalf("unexpected root inode: %#v", rootInode)
	}
	if structs.FixedBytesToString(rootInode.IPerm[:]) != "777" {
		t.Fatalf("root perm = %q, want 777", structs.FixedBytesToString(rootInode.IPerm[:]))
	}

	rootBlock, err := ReadFolderBlock(file, sb, 0)
	if err != nil {
		t.Fatalf("Read root block failed: %v", err)
	}
	if structs.FixedBytesToString(rootBlock.BContent[0].BName[:]) != "." || rootBlock.BContent[0].BInodo != 0 {
		t.Fatalf("bad . entry: %#v", rootBlock.BContent[0])
	}
	if structs.FixedBytesToString(rootBlock.BContent[1].BName[:]) != ".." || rootBlock.BContent[1].BInodo != 0 {
		t.Fatalf("bad .. entry: %#v", rootBlock.BContent[1])
	}
	if structs.FixedBytesToString(rootBlock.BContent[2].BName[:]) != "users.txt" || rootBlock.BContent[2].BInodo != 1 {
		t.Fatalf("bad users.txt entry: %#v", rootBlock.BContent[2])
	}

	usersInode, err := ReadInode(file, sb, 1)
	if err != nil {
		t.Fatalf("Read users inode failed: %v", err)
	}
	if usersInode.IType != '1' || usersInode.IUid != 1 || usersInode.IGid != 1 || usersInode.IBlock[0] != 1 {
		t.Fatalf("unexpected users inode: %#v", usersInode)
	}
	if structs.FixedBytesToString(usersInode.IPerm[:]) != "664" {
		t.Fatalf("users perm = %q, want 664", structs.FixedBytesToString(usersInode.IPerm[:]))
	}

	users, err := ReadRootUsersFile(file, sb)
	if err != nil {
		t.Fatalf("ReadRootUsersFile failed: %v", err)
	}
	if !strings.Contains(users, "1,G,root") || !strings.Contains(users, "1,U,root,root,123") {
		t.Fatalf("unexpected users content: %q", users)
	}

	after := testFileSize(t, path)
	if after != before {
		t.Fatalf("disk size changed from %d to %d", before, after)
	}
}

func TestFormatRejectsInvalidIDAndType(t *testing.T) {
	resetMount(t)
	var out bytes.Buffer
	if err := Format(FormatOptions{ID: "961A", Type: "full"}, &out); err == nil {
		t.Fatal("expected invalid ID error")
	}

	path := createMountedPartition(t, 10, 5, "P")
	_ = path
	if err := Format(FormatOptions{ID: "961A", Type: "fast"}, &out); err == nil {
		t.Fatal("expected invalid type error")
	}
}

func TestFormatRejectsTooSmallPartition(t *testing.T) {
	resetMount(t)
	path := filepath.Join(t.TempDir(), "small.mia")
	if err := disk.MakeDisk(disk.MakeDiskOptions{Size: 1, Unit: "M", Path: path}); err != nil {
		t.Fatalf("MakeDisk failed: %v", err)
	}
	if err := partition.Create(partition.CreateOptions{Size: 512, Unit: "B", Path: path, Type: "P", Name: "Small"}); err != nil {
		t.Fatalf("Create partition failed: %v", err)
	}
	if _, err := mount.Global.Mount(path, "Small"); err != nil {
		t.Fatalf("Mount failed: %v", err)
	}

	var out bytes.Buffer
	if err := Format(FormatOptions{ID: "961A", Type: "full"}, &out); err == nil {
		t.Fatal("expected too-small partition error")
	}
}

func TestFormatAllowsExtendedWithWarning(t *testing.T) {
	resetMount(t)
	createMountedPartition(t, 10, 5, "E")

	var out bytes.Buffer
	if err := Format(FormatOptions{ID: "961A", Type: "full"}, &out); err != nil {
		t.Fatalf("Format extended failed: %v", err)
	}
	if !strings.Contains(out.String(), "Advertencia") {
		t.Fatalf("expected warning, got %q", out.String())
	}
}

func assertBitmapPrefix(t *testing.T, bitmap []byte) {
	t.Helper()
	if len(bitmap) < 3 {
		t.Fatalf("bitmap too small: %d", len(bitmap))
	}
	if bitmap[0] != 1 || bitmap[1] != 1 || bitmap[2] != 0 {
		t.Fatalf("unexpected bitmap prefix: %v", bitmap[:3])
	}
}

func createMountedPartition(t *testing.T, diskSizeMB int64, partitionSizeMB int64, partType string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fs.mia")
	if err := disk.MakeDisk(disk.MakeDiskOptions{Size: diskSizeMB, Unit: "M", Path: path}); err != nil {
		t.Fatalf("MakeDisk failed: %v", err)
	}
	if err := partition.Create(partition.CreateOptions{Size: partitionSizeMB, Unit: "M", Path: path, Type: partType, Name: "Part1"}); err != nil {
		t.Fatalf("Create partition failed: %v", err)
	}
	if _, err := mount.Global.Mount(path, "Part1"); err != nil {
		t.Fatalf("Mount failed: %v", err)
	}
	return path
}

func resetMount(t *testing.T) {
	t.Helper()
	old := mount.Global
	mount.Global = mount.NewManager()
	t.Cleanup(func() {
		mount.Global = old
	})
}

func testFileSize(t *testing.T, path string) int64 {
	t.Helper()
	file, _, err := disk.OpenReadWrite(path)
	if err != nil {
		t.Fatalf("OpenReadWrite failed: %v", err)
	}
	defer file.Close()
	size, err := binio.FileSize(file)
	if err != nil {
		t.Fatalf("FileSize failed: %v", err)
	}
	return size
}
