package explorer

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"mia_p1_201800996/internal/disk"
	"mia_p1_201800996/internal/fs"
	"mia_p1_201800996/internal/mount"
	"mia_p1_201800996/internal/structs"
	"mia_p1_201800996/internal/users"
)

type identityMaps struct {
	users  map[int32]string
	groups map[int32]string
}

func List(id string, targetPath string) (DirectoryListing, error) {
	ctx, err := openContext(id)
	if err != nil {
		return DirectoryListing{}, err
	}
	defer ctx.file.Close()

	inodeIndex, inode, err := fs.ResolvePath(ctx.file, ctx.sb, targetPath)
	if err != nil {
		return DirectoryListing{}, err
	}
	_ = inodeIndex
	if inode.IType != '0' {
		return DirectoryListing{}, fmt.Errorf("la ruta no es carpeta")
	}

	entries, err := fs.ListDirectoryEntries(ctx.file, ctx.sb, inode)
	if err != nil {
		return DirectoryListing{}, err
	}
	identities := loadIdentities(ctx.file, ctx.sb)

	items := make([]Item, 0, len(entries))
	for _, entry := range entries {
		if entry.Name == "." || entry.Name == ".." {
			continue
		}
		itemPath := joinExt2Path(targetPath, entry.Name)
		items = append(items, itemFromInode(entry.Name, itemPath, entry.InodeIndex, entry.Inode, identities))
	}

	return DirectoryListing{
		ID:    id,
		Path:  cleanDisplayPath(targetPath),
		Items: items,
	}, nil
}

func Read(id string, targetPath string) (FileContent, error) {
	ctx, err := openContext(id)
	if err != nil {
		return FileContent{}, err
	}
	defer ctx.file.Close()

	_, inode, err := fs.ResolvePath(ctx.file, ctx.sb, targetPath)
	if err != nil {
		return FileContent{}, err
	}
	if inode.IType != '1' {
		return FileContent{}, fmt.Errorf("la ruta no es archivo")
	}
	content, err := fs.ReadFileContent(ctx.file, ctx.sb, inode)
	if err != nil {
		return FileContent{}, err
	}

	return FileContent{
		ID:      id,
		Path:    cleanDisplayPath(targetPath),
		Name:    baseName(targetPath),
		Content: string(content),
		Size:    inode.ISize,
	}, nil
}

func Stat(id string, targetPath string) (Metadata, error) {
	ctx, err := openContext(id)
	if err != nil {
		return Metadata{}, err
	}
	defer ctx.file.Close()

	inodeIndex, inode, err := fs.ResolvePath(ctx.file, ctx.sb, targetPath)
	if err != nil {
		return Metadata{}, err
	}
	identities := loadIdentities(ctx.file, ctx.sb)
	item := itemFromInode(baseName(targetPath), cleanDisplayPath(targetPath), inodeIndex, inode, identities)
	return Metadata{
		ID:          id,
		Path:        item.Path,
		Name:        item.Name,
		Type:        item.Type,
		Inode:       item.Inode,
		Size:        item.Size,
		Permissions: item.Permissions,
		Owner:       item.Owner,
		Group:       item.Group,
	}, nil
}

type context struct {
	file *os.File
	sb   structs.SuperBlock
}

func openContext(id string) (context, error) {
	mounted, ok := mount.Global.GetMounted(id)
	if !ok {
		return context{}, fmt.Errorf("no existe montaje con id %q", id)
	}
	file, _, err := disk.OpenReadWrite(mounted.DiskPath)
	if err != nil {
		return context{}, err
	}
	sb, err := fs.ReadSuperBlock(file, int64(mounted.Start))
	if err != nil {
		file.Close()
		return context{}, err
	}
	if sb.SMagic != 0xEF53 || sb.SFilesystemType != 2 {
		file.Close()
		return context{}, fmt.Errorf("la particion no tiene formato EXT2")
	}
	return context{file: file, sb: sb}, nil
}

func loadIdentities(file *os.File, sb structs.SuperBlock) identityMaps {
	result := identityMaps{
		users:  map[int32]string{},
		groups: map[int32]string{},
	}
	content, err := fs.ReadRootUsersFile(file, sb)
	if err != nil {
		return result
	}
	groupRecords, userRecords, err := users.ParseUsersFile(content)
	if err != nil {
		return result
	}
	for _, group := range groupRecords {
		if group.Active {
			result.groups[group.ID] = group.Name
		}
	}
	for _, user := range userRecords {
		if user.Active {
			result.users[user.ID] = user.Username
		}
	}
	return result
}

func itemFromInode(name string, itemPath string, inodeIndex int32, inode structs.Inode, identities identityMaps) Item {
	return Item{
		Name:        displayName(name, itemPath),
		Path:        cleanDisplayPath(itemPath),
		Type:        inodeType(inode),
		Size:        inode.ISize,
		Inode:       inodeIndex,
		Permissions: structs.FixedBytesToString(inode.IPerm[:]),
		Owner:       identities.owner(inode.IUid),
		Group:       identities.group(inode.IGid),
	}
}

func (i identityMaps) owner(uid int32) string {
	if name, ok := i.users[uid]; ok {
		return name
	}
	return strconv.FormatInt(int64(uid), 10)
}

func (i identityMaps) group(gid int32) string {
	if name, ok := i.groups[gid]; ok {
		return name
	}
	return strconv.FormatInt(int64(gid), 10)
}

func inodeType(inode structs.Inode) string {
	if inode.IType == '0' {
		return "directory"
	}
	return "file"
}

func joinExt2Path(parent string, name string) string {
	parent = cleanDisplayPath(parent)
	if parent == "/" {
		return "/" + name
	}
	return parent + "/" + name
}

func cleanDisplayPath(value string) string {
	if strings.TrimSpace(value) == "" {
		return "/"
	}
	clean := path.Clean(value)
	if !strings.HasPrefix(clean, "/") {
		return "/" + clean
	}
	return clean
}

func baseName(value string) string {
	clean := cleanDisplayPath(value)
	if clean == "/" {
		return "/"
	}
	return path.Base(clean)
}

func displayName(name string, itemPath string) string {
	if name != "" {
		return name
	}
	return baseName(itemPath)
}
